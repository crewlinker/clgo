package clauthz_test

import (
	"context"
	"embed"
	"io/fs"
	"testing"

	"github.com/crewlinker/clgo/clauthz"
	"github.com/crewlinker/clgo/clzap"
	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuthz(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clauthz")
}

//go:embed testdata/bundles
var bundle embed.FS

type TestAuthzInput struct {
	IsAdmin bool `json:"is_admin"`
}

var _ = Describe("authz (mocked)", func() {
	var autz *clauthz.Authz
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&autz, &obs),
			clzap.TestProvide(),
			clauthz.TestProvide(map[string]string{
				"main.rego": `
				package authz
				import rego.v1
			
				default allow := false
			
				allow if {
					input.is_admin == true
				}
				`,
			}),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should authz to false and decision be logged", func(ctx context.Context) {
		Expect(autz.IsAuthorized(ctx, TestAuthzInput{})).To(BeFalse())
		Expect(obs.FilterMessageSnippet("Decision Log").All()).To(HaveLen(1))
	})

	It("should authz to true with the right input", func(ctx context.Context) {
		Expect(autz.IsAuthorized(ctx, TestAuthzInput{IsAdmin: true})).To(BeTrue())
	})
})

var _ = Describe("authz (served)", func() {
	var autz *clauthz.Authz
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&autz, &obs),
			clauthz.BundleProvide(lo.Must(fs.Sub(bundle, "testdata"))),
			clauthz.Provide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should authz to true with the right input", func(ctx context.Context) {
		Expect(autz.IsAuthorized(ctx, TestAuthzInput{IsAdmin: true})).To(BeTrue())
	})
})
