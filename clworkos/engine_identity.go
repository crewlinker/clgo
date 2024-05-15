package clworkos

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mitchellh/mapstructure"
)

// Impersonator describes who is impersonating the user (if any).
type Impersonator struct {
	// Email of the impersonator.
	Email string `mapstructure:"sub"`
}

// Identity describes the identity as determined by the authentication process.
type Identity struct {
	// IsValid is set to true when the identity could be determined and the identity is valid.
	IsValid bool `mapstructure:"-"`
	// describes when the token that backs this identity expires.
	ExpiresAt time.Time `mapstructure:"-"`
	// WorkOS user id
	UserID string `mapstructure:"-"`
	// WorkOS organization id the user is a member of
	OrganizationID string `mapstructure:"org_id"`
	// ID of the session, used for logout.
	SessionID string `mapstructure:"sid"`
	// Role of the user in the organization.
	Role string `mapstructure:"role"`
	// Impersonator describes who is impersonating the user (if any)
	Impersonator Impersonator `mapstructure:"act"`
}

// BuildSessionToken builds an signed and encrypted session token.
func (e Engine) BuildSessionToken(refreshToken string) (string, error) {
	encryptKey, ok := e.keys.encrypt.public.LookupKeyID(e.cfg.DefaultEncryptKeyID)
	if !ok {
		return "", KeyNotFoundError{id: e.cfg.DefaultEncryptKeyID}
	}

	signKey, ok := e.keys.signing.private.LookupKeyID(e.cfg.DefaultSignKeyID)
	if !ok {
		return "", KeyNotFoundError{id: e.cfg.DefaultSignKeyID}
	}

	tok, err := jwt.NewBuilder().Claim("rt", refreshToken).Build()
	if err != nil {
		return "", fmt.Errorf("failed to build session token: %w", err)
	}

	serialized, err := jwt.NewSerializer().
		Encrypt(jwt.WithKey(encryptKey.Algorithm(), encryptKey)).
		Sign(jwt.WithKey(signKey.Algorithm(), signKey)).
		Serialize(tok)
	if err != nil {
		return "", fmt.Errorf("failed to serialize session token: %w", err)
	}

	return string(serialized), nil
}

// authenticatedSessionFromCookie returns the session from the cookie.
func (e Engine) authenticatedSessionFromCookie(_ context.Context, r *http.Request) (string, error) {
	cookie, err := r.Cookie(e.cfg.SessionCookieName)
	if err != nil {
		return "", fmt.Errorf("failed to get session cookie: %w", err)
	}

	verified, err := jws.Verify([]byte(cookie.Value), jws.WithKeySet(e.keys.signing.public))
	if err != nil {
		return "", fmt.Errorf("failed to verify session token: %w", err)
	}

	decrypted, err := jwe.Decrypt(verified, jwe.WithKeySet(e.keys.encrypt.private))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt session token: %w", err)
	}

	// note: verification is node above, see the discussion in this issue:
	// https://github.com/lestrrat-go/jwx/issues/1133#issuecomment-2112063384
	parsed, err := jwt.Parse(decrypted, jwt.WithVerify(false))
	if err != nil {
		return "", fmt.Errorf("failed to parse session token: %w", err)
	}

	refreshToken, err := fromToken[string](parsed, "rt")
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token from session token: %w", err)
	}

	return refreshToken, nil
}

// identityFromCookie returns the identity of the user based on something carrying cookies.
func (e Engine) identityFromAccessToken(ctx context.Context, token string) (idn Identity, err error) {
	tok, err := e.parseAccessToken(ctx, token)
	if err != nil {
		return idn, fmt.Errorf("failed to parse access token: %w", err)
	}

	idn = Identity{IsValid: true, UserID: tok.Subject(), ExpiresAt: tok.Expiration()}
	claims := tok.PrivateClaims()

	if err := mapstructure.Decode(claims, &idn); err != nil {
		return idn, fmt.Errorf("failed to decode claims as identity: %w", err)
	}

	return idn, nil
}

// parseAccessToken will verify the access token.
func (e Engine) parseAccessToken(_ context.Context, accessToken string) (jwt.Token, error) {
	tok, err := jwt.ParseString(accessToken, jwt.WithKeySet(e.keys.workos.public),
		jwt.WithClock(e.clock),
		jwt.WithIssuer(e.cfg.AccessTokenIssuer),
		jwt.WithAcceptableSkew(e.cfg.TokenValidationAcceptableSkew),
		jwt.WithVerify(true),
		jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("failed to parse, verify and validate JWT (%s): %w", accessToken, err)
	}

	return tok, nil
}

// addAuthenticatedCookies adds cookies to the response that form the user's session.
func (e Engine) addAuthenticatedCookies(
	_ context.Context,
	accessToken, refreshToken string,
	w http.ResponseWriter,
) error {
	serialized, err := e.BuildSessionToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to build session token: %w", err)
	}

	// store the encrypted session token so other requests can use it.
	http.SetCookie(w, &http.Cookie{
		Path:     e.cfg.SessionCookiesPath,
		Domain:   e.cfg.AllCookieDomain,
		SameSite: http.SameSiteLaxMode,
		Name:     e.cfg.SessionCookieName,
		Value:    serialized,
		HttpOnly: true,
		Secure:   e.cfg.AllCookieSecure,
	})

	// set the access token cookie directly
	http.SetCookie(w, &http.Cookie{
		Path:     e.cfg.SessionCookiesPath,
		Domain:   e.cfg.AllCookieDomain,
		SameSite: http.SameSiteLaxMode,
		Name:     e.cfg.AccessTokenCookieName,
		Value:    accessToken,
		HttpOnly: true,
		Secure:   e.cfg.AllCookieSecure,
	})

	return nil
}
