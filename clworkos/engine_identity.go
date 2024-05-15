package clworkos

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

// addAuthenticatedCookies adds cookies to the response that form the user's session.
func (e Engine) addAuthenticatedCookies(
	_ context.Context,
	accessToken, refreshToken string,
	w http.ResponseWriter,
) error {
	encryptKey, ok := e.keys.encrypt.public.LookupKeyID(e.cfg.DefaultEncryptKeyID)
	if !ok {
		return KeyNotFoundError{id: e.cfg.DefaultEncryptKeyID}
	}

	signKey, ok := e.keys.signing.private.LookupKeyID(e.cfg.DefaultSignKeyID)
	if !ok {
		return KeyNotFoundError{id: e.cfg.DefaultSignKeyID}
	}

	tok, err := jwt.NewBuilder().Claim("rt", refreshToken).Build()
	if err != nil {
		return fmt.Errorf("failed to build session token: %w", err)
	}

	serialized, err := jwt.NewSerializer().
		Encrypt(jwt.WithKey(encryptKey.Algorithm(), encryptKey)).
		Sign(jwt.WithKey(signKey.Algorithm(), signKey)).
		Serialize(tok)
	if err != nil {
		return fmt.Errorf("failed to serialize session token: %w", err)
	}

	// store the encrypted session token so other requests can use it.
	http.SetCookie(w, &http.Cookie{
		Domain:   e.cfg.AllCookieDomain,
		SameSite: http.SameSiteLaxMode,
		Name:     e.cfg.SessionCookieName,
		Value:    string(serialized),
		HttpOnly: true,
		Secure:   e.cfg.AllCookieSecure,
	})

	// set the access token cookie directly
	http.SetCookie(w, &http.Cookie{
		Domain:   e.cfg.AllCookieDomain,
		SameSite: http.SameSiteLaxMode,
		Name:     e.cfg.AccessTokenCookieName,
		Value:    accessToken,
		HttpOnly: true,
		Secure:   e.cfg.AllCookieSecure,
	})

	return nil
}
