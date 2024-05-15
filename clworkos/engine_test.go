package clworkos_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/crewlinker/clgo/clworkos"
	"github.com/crewlinker/clgo/clworkos/clworkosmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/fx"
)

var _ = Describe("engine", func() {
	var engine *clworkos.Engine
	var umm *clworkosmock.MockUserManagement
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&engine, &umm),
			Provide(1715748368)) // provide at a wall-clock where tokens have not expired
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

	Describe("continue session", func() {
		It("should return error when access token cookie is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to get access token cookie`)))
		})

		It("should return zero identity when invalid access token", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: "cl_access_token", Value: "invalid.access.token"})

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to parse access token`)))
			Expect(idn).To(Equal(clworkos.Identity{}))
		})

		It("should not refresh and use valid access token", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: "cl_access_token", Value: "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJhY3QiOnsic3ViIjoiYWRtaW5AY3Jld2xpbmtlci5jb20ifSwiaXNzIjoiaHR0cHM6Ly9hcGkud29ya29zLmNvbSIsInN1YiI6InVzZXJfMDFISlRENFZTOFQ2REtBSzVCM0FaVlFGQ1YiLCJzaWQiOiJzZXNzaW9uXzAxSFhYOFZROFNORDVUQ05aQ042TkE3VkZRIiwianRpIjoiMDFIWFg4VlFRSEZOSFI5Q04xQzZOTVRZOEIiLCJvcmdfaWQiOiJvcmdfMDFISlRCUEszWVFNWlk5S0gzMEVYVjlHSE4iLCJyb2xlIjoibWVtYmVyIiwiZXhwIjoxNzE1NzQ4MzY5LCJpYXQiOjE3MTU3NDgwNjl9.DDwEWaHIabk7Uzg9VYce3eX1Kh-x99eKDGH_qbE1QOoy68U3nM9PxDEIIAxdUaT3v91nJtIn-lGa2Woq-wFZGrsd58tfWBmH5kv2SXxaojo1FL-JmDox8eu5Aw1SguVuXPU3r6PawwGScUeDqZ9pAT3qGqS7LyT-jtw_-8nns4D6QttDOF-CzAS4vi9JKujCtBPYLOR_m5axkXp4PEiWMMz5qAoKOpEWFTtfm-X7bD-Yk00hllp7sjk8m5ebpVlDDcT0uL-8Rzp-W64eyvvDfxmp6ZuEaSzvA20AvYPjTKAOGcBJ2V84Ql-5vvWLhEl2-J4IgvxUzfn9dFsUWGwh2Q"})

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).NotTo(HaveOccurred())
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

			Expect(rec.Result().Cookies()).To(BeEmpty())
		})
	})
})

var _ = Describe("engine in present", func() {
	var engine *clworkos.Engine
	var umm *clworkosmock.MockUserManagement
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&engine, &umm),
			Provide(1715772708))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	Describe("continue session", func() {
		It("should complain about session cookie not present", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithExpiredAccessToken(req)

			_, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`named cookie not present`)))
		})

		It("should fail if refresh token is invalid", func(ctx context.Context) {
			oldSessionToken := lo.Must(engine.BuildSessionToken("some.refresh.token"))
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithExpiredAccessToken(req)
			req.AddCookie(&http.Cookie{
				Name:  "cl_session",
				Value: oldSessionToken[:8] + oldSessionToken[9:],
			})

			_, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to verify`)))
		})

		It("should parse the session token", func(ctx context.Context) {
			oldSessionToken := lo.Must(engine.BuildSessionToken("some.refresh.token"))
			newAccessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJpc3MiOiJodHRwczovL2FwaS53b3Jrb3MuY29tIiwic3ViIjoidXNlcl8wMUhKVEQ0VlM4VDZES0FLNUIzQVpWUUZDViIsInNpZCI6InNlc3Npb25fMDFIWFkwQVJKMEJIVDZFQ0hBQ1NLRTZLSDciLCJqdGkiOiIwMUhYWTBBUllXNE5ONFk5UzhIOU1UTUJUWCIsIm9yZ19pZCI6Im9yZ18wMUhKVEJQSzNZUU1aWTlLSDMwRVhWOUdITiIsInJvbGUiOiJtZW1iZXIiLCJleHAiOjE3MTU3NzI5NzksImlhdCI6MTcxNTc3MjY3OX0.iNot86Q5gUmbVqgIqiTuqbHSOCmgbY3XyRCnFXbe6S9kvDvYeBtf5yX9CcG7-6bi8xXmHU0Qv6yCKAcnteNCQNhSlYAZGcbh-yPhoPu_xu6t0tfPEcKRL9OE9HPu3WNt15DLKimjf9Ag0c8tX4HDocLPxn7kkBsWq_BArktM6OQgiQd1dC4jyVYnvGii_fbtiKyiPb9TaRaksKu3saWIML5KA6g4wcLdA91kre4etPWFzRoEEs4RdvSSmeZ23a6ILPHpwvE8PBtlAIXONmgrBqWduT-5Um7OAULF90by8fwZcGE7YevmpEiKcJ8l30IKJs9ymdFWZ0DenXvNGoyMFg"

			umm.EXPECT().AuthenticateWithRefreshToken(mock.Anything,
				usermanagement.AuthenticateWithRefreshTokenOpts{
					RefreshToken: "some.refresh.token",
				}).
				Return(usermanagement.RefreshAuthenticationResponse{
					AccessToken:  newAccessToken,
					RefreshToken: "some.new.refresh_token",
				}, nil).
				Once()

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithExpiredAccessToken(req)
			req.AddCookie(&http.Cookie{
				Name:  "cl_session",
				Value: oldSessionToken,
			})

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).NotTo(HaveOccurred())

			By("checking the cookies being re-set")
			Expect(rec.Result().Cookies()).To(HaveLen(2))
			Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_session"))
			Expect(rec.Result().Cookies()[0].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
			Expect(rec.Result().Cookies()[1].Name).To(Equal("cl_access_token"))
			Expect(rec.Result().Cookies()[1].Value).To(Equal(newAccessToken))

			By("making sure the session got replaced")
			Expect(rec.Result().Cookies()[0].Value).ToNot(Equal(oldSessionToken))

			By("checking the identity")
			Expect(idn).To(Equal(clworkos.Identity{
				IsValid:        true,
				ExpiresAt:      lo.Must(time.Parse(time.RFC3339, "2024-05-15T11:36:19Z")),
				UserID:         "user_01HJTD4VS8T6DKAK5B3AZVQFCV",
				OrganizationID: "org_01HJTBPK3YQMZY9KH30EXV9GHN",
				SessionID:      "session_01HXY0ARJ0BHT6ECHACSKE6KH7",
				Role:           "member",
				Impersonator:   clworkos.Impersonator{},
			}))
		})
	})
})

func WithExpiredAccessToken(req *http.Request) {
	req.AddCookie(&http.Cookie{
		Name:  "cl_access_token",
		Value: "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJhY3QiOnsic3ViIjoiYWRtaW5AY3Jld2xpbmtlci5jb20ifSwiaXNzIjoiaHR0cHM6Ly9hcGkud29ya29zLmNvbSIsInN1YiI6InVzZXJfMDFISlRENFZTOFQ2REtBSzVCM0FaVlFGQ1YiLCJzaWQiOiJzZXNzaW9uXzAxSFhYOFZROFNORDVUQ05aQ042TkE3VkZRIiwianRpIjoiMDFIWFg4VlFRSEZOSFI5Q04xQzZOTVRZOEIiLCJvcmdfaWQiOiJvcmdfMDFISlRCUEszWVFNWlk5S0gzMEVYVjlHSE4iLCJyb2xlIjoibWVtYmVyIiwiZXhwIjoxNzE1NzQ4MzY5LCJpYXQiOjE3MTU3NDgwNjl9.DDwEWaHIabk7Uzg9VYce3eX1Kh-x99eKDGH_qbE1QOoy68U3nM9PxDEIIAxdUaT3v91nJtIn-lGa2Woq-wFZGrsd58tfWBmH5kv2SXxaojo1FL-JmDox8eu5Aw1SguVuXPU3r6PawwGScUeDqZ9pAT3qGqS7LyT-jtw_-8nns4D6QttDOF-CzAS4vi9JKujCtBPYLOR_m5axkXp4PEiWMMz5qAoKOpEWFTtfm-X7bD-Yk00hllp7sjk8m5ebpVlDDcT0uL-8Rzp-W64eyvvDfxmp6ZuEaSzvA20AvYPjTKAOGcBJ2V84Ql-5vvWLhEl2-J4IgvxUzfn9dFsUWGwh2Q",
	})
}
