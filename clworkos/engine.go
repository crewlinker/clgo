package clworkos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lestrrat-go/jwx/v2/jwt"
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
}

// NewUserManagement creates a new UserManagement implementation with the provided configuration.
func NewUserManagement(cfg Config) *usermanagement.Client {
	return usermanagement.NewClient(cfg.APIKey)
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
}

// NewEngine creates a new Engine with the provided UserManagement implementation.
func NewEngine(cfg Config, logs *zap.Logger, keys *Keys, clock Clock, um UserManagement) *Engine {
	return &Engine{
		cfg:   cfg,
		logs:  logs.Named("engine"),
		keys:  keys,
		um:    um,
		clock: clock,
	}
}

// StartSignInFlow starts the sign-in flow as the user is redirected to WorkOS.
func (e Engine) StartSignInFlow(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	redirectTo := r.URL.Query().Get("redirect_to")
	if redirectTo == "" {
		return ErrRedirectToNotProvided
	}

	// make sure the response has a state cookie with a nonce
	nonce, err := e.addStateCookie(ctx, w, redirectTo)
	if err != nil {
		return fmt.Errorf("failed to set up state cookie: %w", err)
	}

	// get the authorization URL from WorkOS
	loc, err := e.um.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    e.cfg.MainClientID,
		RedirectURI: e.cfg.CallbackURL.String(),
		Provider:    "authkit",
		State:       nonce,
	})
	if err != nil {
		return fmt.Errorf("failed to get authorization URL: %w", err)
	}

	// redirect the user to the auth provider
	http.Redirect(w, r, loc.String(), http.StatusFound)

	return nil
}

// HandleSignInCallback handles the sign-in callback as the user returns from WorkOS.
func (e Engine) HandleSignInCallback(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	errorCode, errorDescription := r.URL.Query().Get("error"), r.URL.Query().Get("error_description")
	if errorCode != "" {
		return WorkOSCallbackError{code: errorCode, description: errorDescription}
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return ErrCallbackCodeNotProvided
	}

	// exchange grant (code) for access token and refresh token
	resp, err := e.um.AuthenticateWithCode(ctx, usermanagement.AuthenticateWithCodeOpts{
		ClientID: e.cfg.MainClientID,
		Code:     code,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with code: %w", err)
	}

	// check that the state cookie is valid, and remove it
	redirectTo, err := e.checkAndConsumeStateCookie(ctx, r.URL.Query().Get("state"), resp.Impersonator != nil, w, r)
	if err != nil {
		return fmt.Errorf("failed to verify and consume state cookie: %w", err)
	}

	// add the session cookie to the response, the user is now authenticated
	if err := e.addAuthenticatedCookies(ctx, resp.AccessToken, resp.RefreshToken, w); err != nil {
		return fmt.Errorf("failed to add session cookie: %w", err)
	}

	// redirect the user to the original location (possibly from the state cookie)
	http.Redirect(w, r, redirectTo.String(), http.StatusFound)

	return nil
}

// ContinueSession will continue the user's session, potentially by refreshing it. It is expected to be called
// on every request as part of some middleware logic.
func (e Engine) ContinueSession(ctx context.Context, w http.ResponseWriter, r *http.Request) (idn Identity, err error) {
	atCookie, err := r.Cookie(e.cfg.AccessTokenCookieName)
	if err != nil {
		return idn, InputErrorf("failed to get access token cookie: %w", err)
	}

	idn, err = e.identityFromAccessToken(ctx, atCookie.Value)

	switch {
	case err == nil:
		// valid access token, return identity right away
		return idn, nil
	case !errors.Is(err, jwt.ErrTokenExpired()):
		// some other error with the access token, return the error
		return Identity{}, err
	}

	// read the refresh token from the encrypted session token
	refreshToken, err := e.authenticatedSessionFromCookie(ctx, r)
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

	// add the session cookie to the response, the user is now authenticated
	if err := e.addAuthenticatedCookies(ctx, resp.AccessToken, resp.RefreshToken, w); err != nil {
		return idn, fmt.Errorf("failed to add session cookie: %w", err)
	}

	// return the new identity immediately
	return e.identityFromAccessToken(ctx, resp.AccessToken)
}

// StartSignOutFlow starts the sign-out flow as the user is redirected to WorkOS.
func (e Engine) StartSignOutFlow(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	atCookie, err := r.Cookie(e.cfg.AccessTokenCookieName)
	if err != nil {
		return InputErrorf("failed to get access token cookie: %w", err)
	}

	idn, err := e.identityFromAccessToken(ctx, atCookie.Value)
	if err != nil {
		return fmt.Errorf("failed to get identity from acces token: %w", err)
	}

	logoutURL, err := e.um.GetLogoutURL(usermanagement.GetLogoutURLOpts{SessionID: idn.SessionID})
	if err != nil {
		return fmt.Errorf("failed to get logout URL: %w", err)
	}

	// clear refresh and access token cookies
	http.SetCookie(w, &http.Cookie{MaxAge: -1, Name: e.cfg.AccessTokenCookieName})
	http.SetCookie(w, &http.Cookie{MaxAge: -1, Name: e.cfg.SessionCookieName})

	// redirect to WorkOS to finalize the logout
	http.Redirect(w, r, logoutURL.String(), http.StatusFound)

	return nil
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