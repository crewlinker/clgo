package clworkos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gobwas/glob"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/iter"
	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/zap"
)

// UserManagement interface describes the interface with the WorkOS User Management API.
type UserManagement interface {
	GetAuthorizationURL(
		opts usermanagement.GetAuthorizationURLOpts) (*url.URL, error)
	GetLogoutURL(
		opts usermanagement.GetLogoutURLOpts) (*url.URL, error)
	AuthenticateWithCode(
		ctx context.Context,
		opts usermanagement.AuthenticateWithCodeOpts) (usermanagement.AuthenticateResponse, error)
	AuthenticateWithRefreshToken(
		ctx context.Context,
		opts usermanagement.AuthenticateWithRefreshTokenOpts) (usermanagement.RefreshAuthenticationResponse, error)
	AuthenticateWithPassword(
		ctx context.Context,
		opts usermanagement.AuthenticateWithPasswordOpts,
	) (usermanagement.AuthenticateResponse, error)
	GetUser(ctx context.Context, opts usermanagement.GetUserOpts) (usermanagement.User, error)
	ListUsers(ctx context.Context, opts usermanagement.ListUsersOpts) (usermanagement.ListUsersResponse, error)
	CreateOrganizationMembership(
		ctx context.Context, opts usermanagement.CreateOrganizationMembershipOpts,
	) (usermanagement.OrganizationMembership, error)
	UpdateUser(ctx context.Context, opts usermanagement.UpdateUserOpts) (usermanagement.User, error)
}

// Organizations interface provides organization information from WorkOS.
type Organizations interface {
	CreateOrganization(ctx context.Context, opts organizations.CreateOrganizationOpts) (organizations.Organization, error)
	GetOrganization(ctx context.Context, opts organizations.GetOrganizationOpts) (organizations.Organization, error)
	ListOrganizations(
		ctx context.Context,
		opts organizations.ListOrganizationsOpts,
	) (organizations.ListOrganizationsResponse, error)
	UpdateOrganization(
		ctx context.Context,
		opts organizations.UpdateOrganizationOpts,
	) (organizations.Organization, error)
}

// NewUserManagement creates a new UserManagement implementation with the provided configuration.
func NewUserManagement(cfg Config) *usermanagement.Client {
	return usermanagement.NewClient(cfg.APIKey)
}

// NewOrganizations creates a new UserManagement implementation with the provided configuration.
func NewOrganizations(cfg Config) *organizations.Client {
	return &organizations.Client{
		APIKey: cfg.APIKey,
	}
}

// Clock is an interface for fetching the wall-clock time.
type Clock interface{ jwt.Clock }

// Engine implements the core business logic for WorkOS-powered authentication.
type Engine struct {
	cfg   Config
	logs  *zap.Logger
	keys  *Keys
	clock jwt.Clock
	um    UserManagement
	orgs  Organizations
	hooks Hooks
	globs struct {
		allowedRedirectTo []glob.Glob
	}
}

// NewEngine creates a new Engine with the provided UserManagement implementation.
func NewEngine(
	cfg Config,
	logs *zap.Logger,
	hooks Hooks,
	keys *Keys,
	clock Clock,
	um UserManagement,
	orgs Organizations,
) (eng *Engine, err error) {
	eng = &Engine{
		cfg:   cfg,
		logs:  logs.Named("engine"),
		keys:  keys,
		um:    um,
		orgs:  orgs,
		clock: clock,
		hooks: hooks,
	}

	eng.globs.allowedRedirectTo, err = iter.MapErr(cfg.RedirectToAllowedHosts, func(pat *string) (glob.Glob, error) {
		return glob.Compile(*pat)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to compile allowed redirect_to patterns: %w", err)
	}

	if eng.hooks == nil {
		eng.hooks = NoOpHooks{}
	}

	return eng, nil
}

// StartAuthenticationFlow starts the sign-in flow as the user is redirected to WorkOS.
func (e Engine) StartAuthenticationFlow(
	ctx context.Context, w http.ResponseWriter, r *http.Request, screenHint string,
) (*url.URL, error) {
	redirectTo := r.URL.Query().Get("redirect_to")
	if redirectTo == "" {
		return nil, ErrRedirectToNotProvided
	}

	// make sure the response has a state cookie with a nonce
	nonce, err := e.addStateCookie(ctx, w, redirectTo)
	if err != nil {
		return nil, fmt.Errorf("failed to set up state cookie: %w", err)
	}

	// get the authorization URL from WorkOS
	loc, err := e.um.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    e.cfg.MainClientID,
		RedirectURI: e.cfg.CallbackURL.String(),
		Provider:    "authkit",
		State:       nonce,
		ScreenHint:  usermanagement.ScreenHint(screenHint),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization URL: %w", err)
	}

	return loc, nil
}

// HandleSignInCallback handles the sign-in callback as the user returns from WorkOS.
func (e Engine) HandleSignInCallback(ctx context.Context, w http.ResponseWriter, r *http.Request) (*url.URL, error) {
	errorCode, errorDescription := r.URL.Query().Get("error"), r.URL.Query().Get("error_description")
	if errorCode != "" {
		return nil, WorkOSCallbackError{code: errorCode, description: errorDescription}
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, ErrCallbackCodeNotProvided
	}

	// exchange grant (code) for access token and refresh token
	resp, err := e.um.AuthenticateWithCode(ctx, usermanagement.AuthenticateWithCodeOpts{
		ClientID: e.cfg.MainClientID,
		Code:     code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with code: %w", err)
	}

	idn, err := e.identityFromAccessToken(ctx, resp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity from access token: %w", err)
	}

	user, err := e.um.GetUser(ctx, usermanagement.GetUserOpts{User: idn.UserID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var orgp *organizations.Organization

	if idn.OrganizationID != "" {
		org, err := e.orgs.GetOrganization(ctx, organizations.GetOrganizationOpts{Organization: resp.OrganizationID})
		if err != nil {
			return nil, fmt.Errorf("failed to get organization: %w", err)
		}

		orgp = &org
	}

	if err := e.hooks.AuthenticateWithCodeDidSucceed(ctx, idn, user, orgp); err != nil {
		return nil, fmt.Errorf("failed to run hook: %w", err)
	}

	// determine the redirect based on the state cookie existence
	redirectTo, err := e.checkAndConsumeStateCookie(ctx, r.URL.Query().Get("state"), w, r)
	if err != nil {
		return nil, fmt.Errorf("failed to verify and consume state cookie: %w", err)
	}

	// add the session cookie to the response, the user is now authenticated
	if err := e.addAuthenticatedCookies(ctx, resp.AccessToken, resp.RefreshToken, w); err != nil {
		return nil, fmt.Errorf("failed to add session cookie: %w", err)
	}

	return redirectTo, nil
}

// ErrBasicAuthNotAllowed is returned when a user is not allowed to use basic auth.
var ErrBasicAuthNotAllowed = errors.New("basic auth is not allowed")

// AuthenticateUsernamePassword is used to authenticate a user with a username and password. This method is only allowed
// for certain white-listed usernames since it is generally considered insure when used wrongly.
func (e Engine) AuthenticateUsernamePassword(ctx context.Context, uname, passwd string) (idn Identity, err error) {
	if !lo.Contains(e.cfg.BasicAuthWhitelist, uname) {
		return idn, ErrBasicAuthNotAllowed
	}

	resp, err := e.um.AuthenticateWithPassword(ctx, usermanagement.AuthenticateWithPasswordOpts{
		ClientID: e.cfg.MainClientID,
		Email:    uname,
		Password: passwd,
	})
	if err != nil {
		return idn, fmt.Errorf("failed to authenticate with password: %w", err)
	}

	return e.identityFromAccessToken(ctx, resp.AccessToken)
}

// ContinueSession will continue the user's session, potentially by refreshing it. It is expected to be called
// on every request as part of some middleware logic.
func (e Engine) ContinueSession(ctx context.Context, w http.ResponseWriter, r *http.Request) (idn Identity, err error) {
	// if there is any unexpected error while using the session we clear the tokens. This is to prevent
	// every request from failing if we're in a bad state. It also forces the removal of cookies when
	// the WorkOS backend has deemed the tokens invalid.
	defer func() {
		if err != nil && !errors.Is(err, ErrNoAuthentication) {
			e.clearSessionTokens(ctx, w)
		}
	}()

	// attempt to use the access token first.
	atCookie, atCookieExists, err := readCookie(r, e.cfg.AccessTokenCookieName)
	if err != nil && atCookieExists {
		return idn, InputErrorf("failed to get access token cookie: %w", err)
	}

	// if there is an access token, try to use it to get the identity. It will fail explicitedly
	// if the token is invalid, in the specific case of an expired token we don't return so we
	// can try the refresh token.
	if atCookieExists {
		idn, err = e.identityFromAccessToken(ctx, atCookie.Value)

		switch {
		case err == nil:
			// valid access token, return identity right away
			return idn, nil
		case !errors.Is(err, jwt.ErrTokenExpired()):
			// some unexpected error with the access token, return the error
			return idn, err
		}
	}

	// to refresh we need our sesion cookie holding the refresh token
	rtCookie, rtCookieExists, err := readCookie(r, e.cfg.SessionCookieName)
	if err != nil && rtCookieExists {
		return idn, fmt.Errorf("failed to get session cookie: %w", err)
	} else if !rtCookieExists {
		// no access token, AND no refresh token. Request is not authenticated in any
		// way so we return an error.
		return idn, ErrNoAuthentication
	}

	// read the refresh token from the encrypted session token
	refreshToken, err := e.authenticatedSessionFromCookie(ctx, rtCookie)
	if err != nil {
		return idn, fmt.Errorf("failed to get authenticated session from cookie: %w", err)
	}

	// exchange the refresh token for new tokens.
	resp, err := e.um.AuthenticateWithRefreshToken(ctx, usermanagement.AuthenticateWithRefreshTokenOpts{
		ClientID:     e.cfg.MainClientID,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return idn, fmt.Errorf("failed to authenticate with refresh token: %w", err)
	}

	// (re)add the session cookie to the response, the user is now authenticated
	if err := e.addAuthenticatedCookies(ctx, resp.AccessToken, resp.RefreshToken, w); err != nil {
		return idn, fmt.Errorf("failed to add session cookie: %w", err)
	}

	// return the new identity immediately
	return e.identityFromAccessToken(ctx, resp.AccessToken)
}

// StartSignOutFlow starts the sign-out flow as the user is redirected to WorkOS.
func (e Engine) StartSignOutFlow(ctx context.Context, w http.ResponseWriter, r *http.Request) (*url.URL, error) {
	defer e.clearSessionTokens(ctx, w) // always clear the tokens

	atCookie, atCookieExists, err := readCookie(r, e.cfg.AccessTokenCookieName)
	if err != nil && atCookieExists {
		return nil, InputErrorf("failed to get access token cookie: %w", err)
	} else if !atCookieExists {
		return nil, ErrNoAccessTokenForSignOut
	}

	idn, err := e.identityFromAccessToken(ctx, atCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity from acces token: %w", err)
	}

	logoutURL, err := e.um.GetLogoutURL(usermanagement.GetLogoutURLOpts{SessionID: idn.SessionID})
	if err != nil {
		return nil, fmt.Errorf("failed to get logout URL: %w", err)
	}

	return logoutURL, nil
}

// readCookie allows for reading a cookie and easily asserting if it existed.
func readCookie(r *http.Request, name string) (*http.Cookie, bool, error) {
	cookie, err := r.Cookie(name)
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		return nil, false, err //nolint:wrapcheck
	} else if err != nil {
		return nil, true, fmt.Errorf("failed to read cookie '%s': %w", name, err)
	}

	return cookie, true, nil
}

// fromToken will read 'key' from token and errors if it doesn't exist or is the wrong type.
func fromToken[T any](tok jwt.Token, key string) (r T, err error) {
	v, ok := tok.Get(key)
	if !ok {
		return r, fmt.Errorf("key '%s' not in openid token", key) //nolint:goerr113
	}

	r, ok = v.(T)
	if !ok {
		return r, fmt.Errorf("invalid type for '%s' in openid token: %T", key, v) //nolint:goerr113
	}

	return
}
