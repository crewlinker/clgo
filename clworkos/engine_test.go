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

var (
	// AccessToken for testing, valid at 06:46:08 (CET).
	AccessToken1ValidFor06_46_09GMT = "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJhY3QiOnsic3ViIjoiYWRtaW5AY3Jld2xpbmtlci5jb20ifSwiaXNzIjoiaHR0cHM6Ly9hcGkud29ya29zLmNvbSIsInN1YiI6InVzZXJfMDFISlRENFZTOFQ2REtBSzVCM0FaVlFGQ1YiLCJzaWQiOiJzZXNzaW9uXzAxSFhYOFZROFNORDVUQ05aQ042TkE3VkZRIiwianRpIjoiMDFIWFg4VlFRSEZOSFI5Q04xQzZOTVRZOEIiLCJvcmdfaWQiOiJvcmdfMDFISlRCUEszWVFNWlk5S0gzMEVYVjlHSE4iLCJyb2xlIjoibWVtYmVyIiwiZXhwIjoxNzE1NzQ4MzY5LCJpYXQiOjE3MTU3NDgwNjl9.DDwEWaHIabk7Uzg9VYce3eX1Kh-x99eKDGH_qbE1QOoy68U3nM9PxDEIIAxdUaT3v91nJtIn-lGa2Woq-wFZGrsd58tfWBmH5kv2SXxaojo1FL-JmDox8eu5Aw1SguVuXPU3r6PawwGScUeDqZ9pAT3qGqS7LyT-jtw_-8nns4D6QttDOF-CzAS4vi9JKujCtBPYLOR_m5axkXp4PEiWMMz5qAoKOpEWFTtfm-X7bD-Yk00hllp7sjk8m5ebpVlDDcT0uL-8Rzp-W64eyvvDfxmp6ZuEaSzvA20AvYPjTKAOGcBJ2V84Ql-5vvWLhEl2-J4IgvxUzfn9dFsUWGwh2Q"
	// AccessToken for testing, valid at 13:31:48 (CET).
	AccessToken2ValidFor13_36_19GMT = "eyJhbGciOiJSUzI1NiIsImtpZCI6InNzb19vaWRjX2tleV9wYWlyXzAxSEpUOFFENVdCOVdFTlZYMEE4QTM2UUFNIn0.eyJpc3MiOiJodHRwczovL2FwaS53b3Jrb3MuY29tIiwic3ViIjoidXNlcl8wMUhKVEQ0VlM4VDZES0FLNUIzQVpWUUZDViIsInNpZCI6InNlc3Npb25fMDFIWFkwQVJKMEJIVDZFQ0hBQ1NLRTZLSDciLCJqdGkiOiIwMUhYWTBBUllXNE5ONFk5UzhIOU1UTUJUWCIsIm9yZ19pZCI6Im9yZ18wMUhKVEJQSzNZUU1aWTlLSDMwRVhWOUdITiIsInJvbGUiOiJtZW1iZXIiLCJleHAiOjE3MTU3NzI5NzksImlhdCI6MTcxNTc3MjY3OX0.iNot86Q5gUmbVqgIqiTuqbHSOCmgbY3XyRCnFXbe6S9kvDvYeBtf5yX9CcG7-6bi8xXmHU0Qv6yCKAcnteNCQNhSlYAZGcbh-yPhoPu_xu6t0tfPEcKRL9OE9HPu3WNt15DLKimjf9Ag0c8tX4HDocLPxn7kkBsWq_BArktM6OQgiQd1dC4jyVYnvGii_fbtiKyiPb9TaRaksKu3saWIML5KA6g4wcLdA91kre4etPWFzRoEEs4RdvSSmeZ23a6ILPHpwvE8PBtlAIXONmgrBqWduT-5Um7OAULF90by8fwZcGE7YevmpEiKcJ8l30IKJs9ymdFWZ0DenXvNGoyMFg"
	// fixed wall time used for testing in various places.
	WallTime06_46_08GMT int64 = 1715748368
	WallTime11_31_48GMT int64 = 1715772708
)

var _ = Describe("engine", func() {
	var engine *clworkos.Engine
	var umm *clworkosmock.MockUserManagement
	var orgm *clworkosmock.MockOrganizations
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&engine, &umm, &orgm),
			Provide(WallTime06_46_08GMT)) // provide at a wall-clock where tokens have not expired
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(engine).NotTo(BeNil())
	})

	Describe("start sign-in flow", func() {
		It("should return error when redirect_to is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)

			_, err := engine.StartAuthenticationFlow(ctx, rec, req, "sign-in")
			Expect(err).To(MatchError(clworkos.ErrRedirectToNotProvided))
			Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
		})

		It("should return error when redirect_to is invalid url", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?redirect_to=:", nil)

			_, err := engine.StartAuthenticationFlow(ctx, rec, req, "sign-in")
			Expect(err).To(MatchError(MatchRegexp(`failed to parse`)))
			Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
		})

		It("should return error when redirect_to is not allowed", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?redirect_to=http://example.com", nil)

			_, err := engine.StartAuthenticationFlow(ctx, rec, req, "sign-in")
			Expect(err).To(MatchError(MatchRegexp(`redirect URL is not allowed`)))
			Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
		})

		for _, redirectTo := range []string{
			"localhost",
			"localhost:8080",
			"x.y.z.foo.com",
			"deploy-preview-20--atsdash.netlify.app",
		} {
			It("should add state cookie, and redirect to provider", func(ctx context.Context) {
				umm.EXPECT().GetAuthorizationURL(mock.Anything).Return(lo.Must(
					url.Parse("http://localhost:5354/some/redirect/url"),
				), nil).Once()

				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?redirect_to=http://"+redirectTo, nil)
				loc, err := engine.StartAuthenticationFlow(ctx, rec, req, "sign-in")
				Expect(err).To(Succeed())
				Expect(loc).To(Equal(lo.Must(url.Parse("http://localhost:5354/some/redirect/url"))))

				Expect(rec.Result().Cookies()).To(HaveLen(1))
				Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_auth_state"))
				Expect(rec.Result().Cookies()[0].Value).NotTo(BeEmpty())
				Expect(rec.Result().Cookies()[0].Domain).To(Equal("localhost"))
				Expect(rec.Result().Cookies()[0].Path).To(Equal("/"))
			})
		}
	})

	Describe("handle callback", func() {
		It("should return error when code is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := engine.HandleSignInCallback(ctx, rec, req)
			Expect(err).To(MatchError(clworkos.ErrCallbackCodeNotProvided))
			Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
		})

		It("should return error when provider error is present", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo&error=error&error_description=description", nil)
			_, err := engine.HandleSignInCallback(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`callback with error from WorkOS`)))
		})

		It("should authenticate when without state cookie", func(ctx context.Context) {
			umm.EXPECT().AuthenticateWithCode(mock.Anything, mock.Anything).
				Return(usermanagement.AuthenticateResponse{
					Impersonator: &usermanagement.Impersonator{Email: "admin@admin.com"},
					AccessToken:  AccessToken1ValidFor06_46_09GMT,
					RefreshToken: "some.refresh.token",
				}, nil).
				Once()

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
			loc, err := engine.HandleSignInCallback(ctx, rec, req)
			Expect(err).To(Succeed())

			Expect(rec.Result().Cookies()).To(HaveLen(2))
			Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_session"))
			Expect(rec.Result().Cookies()[0].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
			Expect(rec.Result().Cookies()[1].Name).To(Equal("cl_access_token"))
			Expect(rec.Result().Cookies()[1].Value).To(Equal(AccessToken1ValidFor06_46_09GMT))

			Expect(loc.String()).To(Equal("http://localhost:8080/healthz"))
		})

		Describe("with state cookie", func() {
			var stateToken string
			BeforeEach(func(ctx context.Context) {
				umm.EXPECT().AuthenticateWithCode(mock.Anything, mock.Anything).
					Return(usermanagement.AuthenticateResponse{
						AccessToken:  AccessToken1ValidFor06_46_09GMT,
						RefreshToken: "some.refresh.token",
					}, nil).
					Once()

				stateToken = lo.Must(engine.BuildSignedStateToken("some.nonce", "http://localhost:3834/some/dst"))
			})

			It("with invalid state cookie", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
				WithState(req, "invalid.state.token")

				_, err := engine.HandleSignInCallback(ctx, rec, req)
				Expect(err).To(MatchError(MatchRegexp(`failed to parse, verify and validate the state cookie`)))
			})

			It("with invalid nonce", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo", nil)
				WithState(req, stateToken)

				_, err := engine.HandleSignInCallback(ctx, rec, req)
				Expect(err).To(MatchError(clworkos.ErrStateNonceMismatch))
				Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
			})

			It("with valid state cookie", func(ctx context.Context) {
				rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?code=foo&state=some.nonce", nil)
				WithState(req, stateToken)

				loc, err := engine.HandleSignInCallback(ctx, rec, req)
				Expect(err).To(Succeed())
				Expect(loc.String()).To(Equal("http://localhost:3834/some/dst"))

				Expect(rec.Result().Cookies()).To(HaveLen(3))
				Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_auth_state"))
				Expect(rec.Result().Cookies()[0].MaxAge).To(Equal(-1))
				Expect(rec.Result().Cookies()[0].Domain).To(Equal("localhost"))
				Expect(rec.Result().Cookies()[0].Path).To(Equal("/"))
				Expect(rec.Result().Cookies()[1].Name).To(Equal("cl_session"))
				Expect(rec.Result().Cookies()[1].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
				Expect(rec.Result().Cookies()[1].Path).To(Equal("/"))
				Expect(rec.Result().Cookies()[2].Name).To(Equal("cl_access_token"))
				Expect(rec.Result().Cookies()[2].Value).To(Equal(AccessToken1ValidFor06_46_09GMT))
				Expect(rec.Result().Cookies()[2].Path).To(Equal("/"))
			})
		})
	})

	Describe("continue session", func() {
		It("should error when neither tokens are present", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := engine.ContinueSession(ctx, rec, req)

			Expect(err).To(MatchError(clworkos.ErrNoAuthentication))
			Expect(rec.Result().Cookies()).To(BeEmpty()) // no reset cookies
		})

		It("should return zero identity when invalid access token", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, "invalid.access.token")

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to parse access token`)))
			Expect(idn).To(Equal(clworkos.Identity{}))
		})

		It("should not refresh and use valid access token", func(ctx context.Context) {
			sessionToken := lo.Must(engine.BuildSessionToken(clworkos.Session{
				RefreshToken:            "some.refresh.token",
				OrganizationIDOverwrite: "org_01J3PZ6T0NTYHHKJP0HVZTYJ1E",
				RoleOverwrite:           "foo",
			}))

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, AccessToken1ValidFor06_46_09GMT)
			WithSession(req, sessionToken)

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(idn).To(Equal(clworkos.Identity{
				IsValid:        true,
				ExpiresAt:      lo.Must(time.Parse(time.RFC3339, "2024-05-15T04:46:09Z")),
				UserID:         "user_01HJTD4VS8T6DKAK5B3AZVQFCV",
				OrganizationID: "org_01J3PZ6T0NTYHHKJP0HVZTYJ1E",
				SessionID:      "session_01HXX8VQ8SND5TCNZCN6NA7VFQ",
				Role:           "foo",
				Impersonator: clworkos.Impersonator{
					Email: "admin@crewlinker.com",
				},
			}))

			Expect(rec.Result().Cookies()).To(BeEmpty())
		})
	})

	Describe("sign out", func() {
		It("should return error when access token cookie is missing", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := engine.StartSignOutFlow(ctx, rec, req)
			Expect(err).To(MatchError(clworkos.ErrNoAccessTokenForSignOut))
			Expect(clworkos.IsBadRequestError(err)).To(BeTrue())
		})

		It("should error when invalid access token", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, "invalid.access.token")

			_, err := engine.StartSignOutFlow(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to parse access token`)))
		})

		It("start logout flow successfully", func(ctx context.Context) {
			umm.EXPECT().GetLogoutURL(mock.Anything).Return(lo.Must(url.Parse("http://localhost:8080/logout")), nil).Once()

			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, AccessToken1ValidFor06_46_09GMT)

			loc, err := engine.StartSignOutFlow(ctx, rec, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(loc).To(Equal(lo.Must(url.Parse("http://localhost:8080/logout"))))

			By("checking the cookies being re-set")
			ExpectSessionClear(rec)
		})
	})

	Describe("username password", func() {
		It("should not allow if not in whitelist", func(ctx context.Context) {
			idn, fromCache, err := engine.AuthenticateUsernamePassword(ctx, "foo", "bar")
			Expect(err).To(MatchError(clworkos.ErrBasicAuthNotAllowed))
			Expect(idn.IsValid).To(BeFalse())
			Expect(fromCache).To(BeFalse())
		})

		It("should do once if on whitelist and token valid", func(ctx context.Context) {
			umm.EXPECT().AuthenticateWithPassword(mock.Anything, mock.Anything).
				Return(usermanagement.AuthenticateResponse{
					AccessToken:  AccessToken1ValidFor06_46_09GMT, // wall time: WallTime06_46_08GMT
					RefreshToken: "some.refresh.token",
				}, nil).
				Once()

			idn, fromCache, err := engine.AuthenticateUsernamePassword(ctx, "admin+system1@crewlinker.com", "bar")
			Expect(err).ToNot(HaveOccurred())
			Expect(idn.IsValid).To(BeTrue())
			Expect(idn.UserID).To(Equal(`user_01HJTD4VS8T6DKAK5B3AZVQFCV`))
			Expect(fromCache).To(BeFalse())

			idn, fromCache, err = engine.AuthenticateUsernamePassword(ctx, "admin+system1@crewlinker.com", "bar")
			Expect(err).ToNot(HaveOccurred())
			Expect(idn.IsValid).To(BeTrue())
			Expect(idn.UserID).To(Equal(`user_01HJTD4VS8T6DKAK5B3AZVQFCV`))
			Expect(fromCache).To(BeTrue())
		})
	})
})

var ExampleSession1 = clworkos.Session{RefreshToken: "some.refresh.token"}

var _ = Describe("engine in present", func() {
	var engine *clworkos.Engine
	var umm *clworkosmock.MockUserManagement
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&engine, &umm),
			Provide(WallTime11_31_48GMT))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	Describe("continue session", func() {
		It("should be unauthenticated with expired access token en missing refresh token", func(ctx context.Context) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, AccessToken1ValidFor06_46_09GMT)

			_, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(clworkos.ErrNoAuthentication))
		})

		It("should fail if refresh token is invalid", func(ctx context.Context) {
			oldSessionToken := lo.Must(engine.BuildSessionToken(ExampleSession1))
			rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			WithAccessToken(req, AccessToken1ValidFor06_46_09GMT)
			WithSession(req, oldSessionToken[:8]+oldSessionToken[9:])

			_, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).To(MatchError(MatchRegexp(`failed to verify`)))
			ExpectSessionClear(rec) // clear session on any error
		})

		It("should succeed with new access token", func(ctx context.Context) {
			oldSessionToken := lo.Must(engine.BuildSessionToken(ExampleSession1))
			newAccessToken := AccessToken2ValidFor13_36_19GMT

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
			WithAccessToken(req, AccessToken1ValidFor06_46_09GMT)
			WithSession(req, oldSessionToken)

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).NotTo(HaveOccurred())
			ExpectRefreshedSession(rec, newAccessToken, oldSessionToken, idn)
		})

		It("should refresh without access token if session token is present", func(ctx context.Context) {
			oldSessionToken := lo.Must(engine.BuildSessionToken(ExampleSession1))
			newAccessToken := AccessToken2ValidFor13_36_19GMT

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
			WithSession(req, oldSessionToken)

			idn, err := engine.ContinueSession(ctx, rec, req)
			Expect(err).NotTo(HaveOccurred())
			ExpectRefreshedSession(rec, newAccessToken, oldSessionToken, idn)
		})
	})
})

func ExpectSessionClear(rec *httptest.ResponseRecorder) {
	at, ok := lo.Find(rec.Result().Cookies(), func(c *http.Cookie) bool { return c.Name == "cl_access_token" })
	Expect(ok).To(BeTrue())
	st, ok := lo.Find(rec.Result().Cookies(), func(c *http.Cookie) bool { return c.Name == "cl_session" })
	Expect(ok).To(BeTrue())

	for _, c := range []*http.Cookie{at, st} {
		Expect(c.MaxAge).To(Equal(-1))
		Expect(c.Path).To(Equal("/"))
		Expect(c.Value).To(BeEmpty())
		Expect(c.Domain).To(Equal("localhost"))
		Expect(c.SameSite).To(Equal(http.SameSiteNoneMode))
	}
}

func ExpectRefreshedSession(
	rec *httptest.ResponseRecorder,
	newAccessToken, oldSessionToken string,
	idn clworkos.Identity,
) {
	By("checking the cookies being re-set")
	Expect(rec.Result().Cookies()).To(HaveLen(2))
	Expect(rec.Result().Cookies()[0].SameSite).To(Equal(http.SameSiteNoneMode))
	Expect(rec.Result().Cookies()[0].Name).To(Equal("cl_session"))
	Expect(rec.Result().Cookies()[0].Value).To(MatchRegexp(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`))
	Expect(rec.Result().Cookies()[1].SameSite).To(Equal(http.SameSiteNoneMode))
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
}

func WithAccessToken(req *http.Request, token string) {
	req.AddCookie(&http.Cookie{
		Name:  "cl_access_token",
		Value: token,
	})
}

func WithState(req *http.Request, state string) {
	req.AddCookie(&http.Cookie{
		Name:  "cl_auth_state",
		Value: state,
	})
}

func WithSession(req *http.Request, sessionToken string) {
	req.AddCookie(&http.Cookie{
		Name:  "cl_session",
		Value: sessionToken,
	})
}
