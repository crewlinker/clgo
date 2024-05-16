package clworkos_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clworkos"
	"github.com/crewlinker/clgo/clworkos/clworkosmock"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

func TestClworkos(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clworkos")
}

var _ = Describe("handler", func() {
	var hdlr *clworkos.Handler
	var mmu *clworkosmock.MockUserManagement
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&hdlr, &mmu, &obs), Provide(1715748368))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(hdlr).NotTo(BeNil())
	})

	Describe("sign in", func() {
		It("should serve bad request error", func() {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/sign-in", nil)
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(rec.Body.String()).To(ContainSubstring("missing redirect_to query parameter"))
		})

		It("should serve redirect", func() {
			mmu.EXPECT().GetAuthorizationURL(mock.Anything).Return(&url.URL{Scheme: "https", Host: "workos.com"}, nil)

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/sign-in?redirect_to=http://localhost:8080", nil)
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusFound))
			Expect(rec.Header().Get("Location")).To(ContainSubstring("https://workos.com"))
		})
	})

	Describe("callback", func() {
		It("should serve bad request error", func() {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/callback", nil)
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(rec.Body.String()).To(ContainSubstring("missing code query parameter"))
		})

		It("should serve redirect", func() {
			mmu.EXPECT().AuthenticateWithCode(mock.Anything, mock.Anything).Return(usermanagement.AuthenticateResponse{
				Impersonator: &usermanagement.Impersonator{Email: "a@a.com"},
			}, nil)

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/callback?code=123", nil)
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusFound))
			Expect(rec.Header().Get("Location")).To(ContainSubstring("http://localhost:8080"))
		})
	})

	Describe("sign out", func() {
		It("should serve bad request", func() {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/sign-out", nil)
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(rec.Body.String()).To(ContainSubstring("no credentials"))
		})

		It("should serve redirect", func() {
			mmu.EXPECT().GetLogoutURL(mock.Anything).Return(&url.URL{Scheme: "https", Host: "workos.com"}, nil)

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/sign-out", nil)
			req.AddCookie(&http.Cookie{Name: "cl_access_token", Value: AccessToken1ValidFor06_46_08})
			hdlr.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(http.StatusFound))
			Expect(rec.Header().Get("Location")).To(ContainSubstring("https://workos.com"))
		})
	})

	Describe("middleware", func() {
		It("should have zero-value identity unauthenticated requests, and no logs", func() {
			var idn *clworkos.Identity
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/foo", nil)

			hdlr.Authenticate()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				v := clworkos.IdentityFromContext(r.Context())
				idn = &v
			})).ServeHTTP(rec, req)

			Expect(idn).To(Equal(&clworkos.Identity{}))
			Expect(obs.FilterMessage("middleware failed to continue session").Len()).To(Equal(0))
		})

		It("should set identity on authenticated request", func() {
			var idn clworkos.Identity
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.AddCookie(&http.Cookie{Name: "cl_access_token", Value: "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJhY3QiOnsic3ViIjoiYWRtaW5AY3Jld2xpbmtlci5jb20ifSwiaXNzIjoiaHR0cHM6Ly9hcGkud29ya29zLmNvbSIsInN1YiI6InVzZXJfMDFISlRENFZTOFQ2REtBSzVCM0FaVlFGQ1YiLCJzaWQiOiJzZXNzaW9uXzAxSFhYOFZROFNORDVUQ05aQ042TkE3VkZRIiwianRpIjoiMDFIWFg4VlFRSEZOSFI5Q04xQzZOTVRZOEIiLCJvcmdfaWQiOiJvcmdfMDFISlRCUEszWVFNWlk5S0gzMEVYVjlHSE4iLCJyb2xlIjoibWVtYmVyIiwiZXhwIjoxNzE1NzQ4MzY5LCJpYXQiOjE3MTU3NDgwNjl9.DDwEWaHIabk7Uzg9VYce3eX1Kh-x99eKDGH_qbE1QOoy68U3nM9PxDEIIAxdUaT3v91nJtIn-lGa2Woq-wFZGrsd58tfWBmH5kv2SXxaojo1FL-JmDox8eu5Aw1SguVuXPU3r6PawwGScUeDqZ9pAT3qGqS7LyT-jtw_-8nns4D6QttDOF-CzAS4vi9JKujCtBPYLOR_m5axkXp4PEiWMMz5qAoKOpEWFTtfm-X7bD-Yk00hllp7sjk8m5ebpVlDDcT0uL-8Rzp-W64eyvvDfxmp6ZuEaSzvA20AvYPjTKAOGcBJ2V84Ql-5vvWLhEl2-J4IgvxUzfn9dFsUWGwh2Q"})
			hdlr.Authenticate()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idn = clworkos.IdentityFromContext(r.Context())
			})).ServeHTTP(rec, req)

			Expect(obs.FilterMessage("authenticated identity").Len()).To(Equal(1))

			Expect(idn).To(Equal(clworkos.Identity{
				IsValid:        true,
				ExpiresAt:      lo.Must(time.Parse(time.RFC3339, "2024-05-15T04:46:09Z")),
				UserID:         "user_01HJTD4VS8T6DKAK5B3AZVQFCV",
				OrganizationID: "org_01HJTBPK3YQMZY9KH30EXV9GHN",
				SessionID:      "session_01HXX8VQ8SND5TCNZCN6NA7VFQ",
				Role:           "member",
				Impersonator: clworkos.Impersonator{
					Email: "admin@crewlinker.com",
				},
			}))
		})
	})
})

func Provide(clockAt int64) fx.Option {
	return fx.Options(
		clworkos.TestProvide(GinkgoTB(), clockAt),
		clzap.TestProvide(),
	)
}
