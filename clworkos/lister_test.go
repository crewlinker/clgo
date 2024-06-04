package clworkos_test

import (
	"context"

	"github.com/crewlinker/clgo/clworkos"
	"github.com/crewlinker/clgo/clworkos/clworkosmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"github.com/workos/workos-go/v4/pkg/common"
	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/fx"
)

var _ = Describe("lister", func() {
	var lister *clworkos.Lister
	var umm *clworkosmock.MockUserManagement
	var orgm *clworkosmock.MockOrganizations
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&lister, &umm, &orgm),
			Provide(1715748368)) // provide at a wall-clock where tokens have not expired
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(lister).NotTo(BeNil())
	})

	It("should list users", func(ctx context.Context) {
		umm.EXPECT().ListUsers(mock.Anything,
			usermanagement.ListUsersOpts{After: "", Limit: 100}).Return(usermanagement.ListUsersResponse{
			ListMetadata: common.ListMetadata{
				After: "user_1",
			},
			Data: []usermanagement.User{
				{},
				{},
			},
		}, nil).Once()
		umm.EXPECT().ListUsers(mock.Anything,
			usermanagement.ListUsersOpts{After: "user_1", Limit: 100}).Return(usermanagement.ListUsersResponse{
			ListMetadata: common.ListMetadata{},
			Data: []usermanagement.User{
				{},
			},
		}, nil).Once()

		users, err := lister.ListAllUsers(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(users).To(HaveLen(3))
	})

	It("should list organizations", func(ctx context.Context) {
		orgm.EXPECT().ListOrganizations(mock.Anything,
			organizations.ListOrganizationsOpts{After: "", Limit: 100}).Return(organizations.ListOrganizationsResponse{
			ListMetadata: common.ListMetadata{
				After: "user_1",
			},
			Data: []organizations.Organization{
				{},
				{},
			},
		}, nil).Once()
		orgm.EXPECT().ListOrganizations(mock.Anything,
			organizations.ListOrganizationsOpts{After: "user_1", Limit: 100}).Return(organizations.ListOrganizationsResponse{
			ListMetadata: common.ListMetadata{},
			Data: []organizations.Organization{
				{},
			},
		}, nil).Once()

		orgs, err := lister.ListAllOrganizations(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(orgs).To(HaveLen(3))
	})
})
