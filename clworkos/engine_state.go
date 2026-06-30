package clworkos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gobwas/glob"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/samber/lo"
)

// stateNonceSize is the size of the nonce used for the auth flow.
const stateNonceSize = 32

// stateCookieMaxAge is the max age of the nonce cookie.
const stateCookieMaxAge = 7200

// BuildSignedStateToken builds a signed state token with the provided nonce and redirect_to values. It is
// a public method to make black box testing easier.
func (e Engine) BuildSignedStateToken(nonce, redirectTo string) (string, error) {
	token, err := jwt.NewBuilder().
		Claim("nonce", nonce).
		Claim("redirect_to", redirectTo).
		Build()
	if err != nil {
		return "", fmt.Errorf("failed to build state token: %w", err)
	}

	key, ok := e.keys.signing.private.LookupKeyID(e.cfg.DefaultSignKeyID)
	if !ok {
		return "", KeyNotFoundError{id: e.cfg.DefaultSignKeyID}
	}

	signed, err := jwt.Sign(token, jwt.WithKey(key.Algorithm(), key))
	if err != nil {
		return "", fmt.Errorf("failed to sign state token: %w", err)
	}

	return string(signed), nil
}

// addStateCookie sets up the state cookie on 'w'. It returns the raw token value.
func (e Engine) addStateCookie(
	_ context.Context, w http.ResponseWriter, redirectTo string,
) (nonce string, err error) {
	redirectURL, err := url.Parse(redirectTo)
	if err != nil {
		return "", InputErrorf("failed to parse redirect URL: %w", err)
	}

	if _, found := lo.Find(e.globs.allowedRedirectTo, func(pat glob.Glob) bool {
		return pat.Match(redirectURL.Hostname())
	}); !found {
		return "", RedirectToNotAllowedError{actual: redirectURL.Hostname(), allowed: e.cfg.RedirectToAllowedHosts}
	}

	nonce = lo.RandomString(stateNonceSize, lo.AlphanumericCharset)

	signed, err := e.BuildSignedStateToken(nonce, redirectURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to build signed state token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Domain:   e.cfg.AllCookieDomain,
		SameSite: e.cfg.AllCookieSameSite,
		Name:     e.cfg.StateCookieName,
		Value:    signed,
		HttpOnly: true,
		Secure:   e.cfg.AllCookieSecure,
		MaxAge:   stateCookieMaxAge,
		Path:     e.cfg.AllCookiePath,
	})

	return nonce, nil
}

// checkAndConsumeStateCookie compares the nonce from the query with the nonce stored in the state cookie.
func (e Engine) checkAndConsumeStateCookie(
	_ context.Context,
	nonceFromQuery string,
	w http.ResponseWriter,
	r *http.Request,
) (redirectTo *url.URL, err error) {
	cookie, cookieExists, err := readCookie(r, e.cfg.StateCookieName)
	if cookieExists && err != nil {
		return nil, fmt.Errorf("failed to read state cookie: %w", err)
	}

	// If WorkOS did not echo back our state nonce, the flow was not initiated by us through
	// StartAuthenticationFlow (e.g. password-reset, invitation or impersonation flows). There is
	// nothing to verify in that case. We still clear any stale state cookie that might linger from a
	// previously abandoned sign-in attempt (otherwise the nonce comparison below would fail with
	// ErrStateNonceMismatch) and fall back to the default redirect.
	if nonceFromQuery == "" {
		if cookieExists {
			e.clearStateCookie(w)
		}

		return e.cfg.DefaultRedirectTo, nil
	}

	// in case the user is impersonated, or invited. The state cookie will not be present. In that case
	// well have to redirect the user to a default URL.
	if !cookieExists {
		return e.cfg.DefaultRedirectTo, nil
	}

	// check validity of the state cookie
	stateToken, err := jwt.ParseString(cookie.Value,
		jwt.WithClock(e.clock),
		jwt.WithKeySet(e.keys.signing.public))
	if err != nil {
		return nil, fmt.Errorf("failed to parse, verify and validate the state cookie: %w", err)
	}

	nonceFromCookie, err := fromToken[string](stateToken, "nonce")
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce from state token: %w", err)
	}

	if nonceFromQuery != nonceFromCookie {
		return nil, ErrStateNonceMismatch
	}

	redirectValue, err := fromToken[string](stateToken, "redirect_to")
	if err != nil {
		return nil, fmt.Errorf("failed to get redirect_to from state token: %w", err)
	}

	redirectTo, err = url.Parse(redirectValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redirect_to URL: %w", err)
	}

	// clear the state (nonce) cookie, only useful until the callback
	e.clearStateCookie(w)

	return redirectTo, nil
}

// clearStateCookie expires the state (nonce) cookie. It is only useful until the callback has been handled.
func (e Engine) clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		MaxAge:   -1,
		Name:     e.cfg.StateCookieName,
		Path:     e.cfg.AllCookiePath,
		Domain:   e.cfg.AllCookieDomain,
		SameSite: e.cfg.AllCookieSameSite,
		Secure:   e.cfg.AllCookieSecure,
	})
}
