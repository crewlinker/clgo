package clworkos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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

	if !lo.Contains(e.cfg.RedirectToAllowedHosts, redirectURL.Hostname()) {
		return "", RedirectToNotAllowedError{actual: redirectURL.Hostname(), allowed: e.cfg.RedirectToAllowedHosts}
	}

	nonce = lo.RandomString(stateNonceSize, lo.AlphanumericCharset)

	signed, err := e.BuildSignedStateToken(nonce, redirectURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to build signed state token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Domain:   e.cfg.AllCookieDomain,
		SameSite: http.SameSiteLaxMode,
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
	isImpersonated bool,
	w http.ResponseWriter,
	r *http.Request,
) (redirectTo *url.URL, err error) {
	if isImpersonated {
		return e.cfg.RedirectToIfImpersonated, nil
	}

	// only complete the flow if the nonce in the users cookie matches the nonce in the callback
	cookie, _, err := readCookie(r, e.cfg.StateCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		return nil, ErrStateCookieNotPresentOrInvalid
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
	http.SetCookie(w, &http.Cookie{
		MaxAge: -1,
		Name:   e.cfg.StateCookieName,
		Path:   e.cfg.AllCookiePath,
		Domain: e.cfg.AllCookieDomain,
	})

	return redirectTo, nil
}
