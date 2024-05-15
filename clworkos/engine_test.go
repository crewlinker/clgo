package clworkos_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/crewlinker/clgo/clworkos"
	"github.com/crewlinker/clgo/clworkos/clworkosmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/fx"
)

var _ = Describe("handler", func() {
	var engine *clworkos.Engine
	var umm *clworkosmock.MockUserManagement
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&engine, &umm), Provide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(engine).NotTo(BeNil())
	})

	Describe("start sign-in flow", func() {
		It("should return error when redirect_to is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			Expect(engine.StartSignInFlow(ctx, rec, req)).To(MatchError(clworkos.ErrRedirectToNotProvided))
		})

		It("should return error when redirect_to is not allowed", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?redirect_to=http://example.com", nil)
			Expect(engine.StartSignInFlow(ctx, rec, req)).To(MatchError(MatchRegexp(`redirect URL is not allowed`)))
		})

		It("should add state cookie, and redirect to provider", func(ctx context.Context) {
			umm.EXPECT().GetAuthorizationURL(mock.Anything).Return(lo.Must(
				url.Parse("http://localhost:5354/some/redirect/url"),
			), nil).Once()

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?redirect_to=http://localhost", nil)
			Expect(engine.StartSignInFlow(ctx, rec, req)).To(Succeed())

			Expect(rec.Result().StatusCode).To(Equal(http.StatusFound))
			Expect(rec.Result().Header.Get("Location")).To(Equal("http://localhost:5354/some/redirect/url"))

			Expect(rec.Result().Cookies()).To(HaveLen(1))
			Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_auth_state"))
			Expect(rec.Result().Cookies()[0].Value).NotTo(BeEmpty())
		})
	})

	Describe("handle sign-in callback", func() {
		It("should return error when code is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			Expect(engine.HandleSignInCallback(ctx, rec, req)).To(MatchError(clworkos.ErrCallbackCodeNotProvided))
		})

		It("should return error when error is present", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo&error=error&error_description=description", nil)
			Expect(engine.HandleSignInCallback(ctx, rec, req)).To(MatchError(MatchRegexp(`callback with error from WorkOS`)))
		})

		It("should authenticate when impersonated", func(ctx context.Context) {
			umm.EXPECT().AuthenticateWithCode(mock.Anything, mock.Anything).
				Return(usermanagement.AuthenticateResponse{
					Impersonator: &usermanagement.Impersonator{Email: "admin@admin.com"},
					AccessToken:  "some.access.token",
					RefreshToken: "some.refresh.token",
				}, nil).
				Once()

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
			Expect(engine.HandleSignInCallback(ctx, rec, req)).To(Succeed())

			Expect(rec.Result().Cookies()).To(HaveLen(2))
			Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_session"))
			Expect(rec.Result().Cookies()[0].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
			Expect(rec.Result().Cookies()[1].Name).To(Equal("cl_access_token"))
			Expect(rec.Result().Cookies()[1].Value).To(Equal(`some.access.token`))

			Expect(rec.Result().StatusCode).To(Equal(http.StatusFound))
			Expect(rec.Result().Header.Get("Location")).To(Equal("http://localhost:8080/healthz"))
		})

		Describe("non-impersonated", func() {
			var stateToken string
			BeforeEach(func(ctx context.Context) {
				umm.EXPECT().AuthenticateWithCode(mock.Anything, mock.Anything).
					Return(usermanagement.AuthenticateResponse{
						AccessToken:  "some.access.token",
						RefreshToken: "some.refresh.token",
					}, nil).
					Once()

				stateToken = lo.Must(engine.BuildSignedStateToken("some.nonce", "http://localhost:3834/some/dst"))
			})

			It("without state cookie", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
				Expect(engine.HandleSignInCallback(ctx, rec, req)).To(MatchError(clworkos.ErrStateCookieNotPresentOrInvalid))
			})

			It("with invalid state cookie", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
				req.AddCookie(&http.Cookie{Name: "cl_auth_state", Value: "invalid.state.token"})

				Expect(engine.HandleSignInCallback(ctx, rec, req)).To(
					MatchError(MatchRegexp(`failed to parse, verify and validate the state cookie`)))
			})

			It("with invalid nonce", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
				req.AddCookie(&http.Cookie{Name: "cl_auth_state", Value: stateToken})

				Expect(engine.HandleSignInCallback(ctx, rec, req)).To(
					MatchError(clworkos.ErrStateNonceMismatch))
			})

			It("with valid state cookie", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo&state=some.nonce", nil)
				req.AddCookie(&http.Cookie{Name: "cl_auth_state", Value: stateToken})

				Expect(engine.HandleSignInCallback(ctx, rec, req)).To(Succeed())

				Expect(rec.Result().StatusCode).To(Equal(http.StatusFound))
				Expect(rec.Result().Header.Get("Location")).To(Equal("http://localhost:3834/some/dst"))

				Expect(rec.Result().Cookies()).To(HaveLen(3))
				Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_auth_state"))
				Expect(rec.Result().Cookies()[0].MaxAge).To(Equal(-1)) // expire the state cookie
				Expect(rec.Result().Cookies()[1].Name).To(Equal("cl_session"))
				Expect(rec.Result().Cookies()[1].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
				Expect(rec.Result().Cookies()[2].Name).To(Equal("cl_access_token"))
				Expect(rec.Result().Cookies()[2].Value).To(Equal(`some.access.token`))
			})
		})
	})
})
