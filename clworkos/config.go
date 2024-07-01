package clworkos

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Config configures the package components.
type Config struct {
	// WorkOS API key.
	APIKey string `env:"API_KEY"`
	// WorkOS main client ID.
	MainClientID string `env:"MAIN_CLIENT_ID"`
	// Full url to where the user will be send after WorkOS has done the authorization.
	CallbackURL *url.URL `env:"CALLBACK_URL" envDefault:"http://localhost:8080/auth/callback"`
	// AccessTokenIssuer is used to check the access token issuer.
	AccessTokenIssuer string `env:"ISSUER" envDefault:"https://api.workos.com"`
	// TokenValidationAcceptableSkew provides some slack in checking token times.
	TokenValidationAcceptableSkew time.Duration `env:"TOKEN_VALIDATION_ACCEPTABLE_SKEW" envDefault:"10s"`
	// AllCookieSecure will set the secure flag on all cookies.
	AllCookieSecure bool `env:"ALL_COOKIE_SECURE" envDefault:"true"`
	// AllCookieDomain configures the domain for all cookies
	AllCookieDomain string `env:"ALL_COOKIE_DOMAIN" envDefault:"localhost"`
	// AllCookiePath configures the path attribute for session cookies
	AllCookiePath string `env:"ALL_COOKIE_PATH" envDefault:"/"`
	// AllCookieSameSite configures the same-site attribute for all cookies
	AllCookieSameSite http.SameSite `env:"ALL_COOKIE_SAME_SITE" envDefault:"4"`
	// ShowErrorMessagesToClient will show server errors to the client, should only be visible in development.
	ShowErrorMessagesToClient bool `env:"SHOW_ERROR_MESSAGES_TO_CLIENT" envDefault:"false"`
	// RedirectToAllowedHosts is a list of hosts that are allowed to be redirected to.
	RedirectToAllowedHosts []string `env:"REDIRECT_TO_ALLOWED_HOSTS" envDefault:"localhost"`
	// JWKEndpoint is the endpoint for fetching the public key set for verifying the access key.
	JWKEndpoint string `env:"JWK_ENDPOINT" envDefault:"https://api.workos.com/sso/jwks/"`
	// PubPrivSigningKeySetB64JSON will hold private keys for JWE encryption
	PubPrivEncryptKeySetB64JSON string `env:"PUB_PRIV_ENCRYPT_KEY_SET_B64_JSON" envDefault:"eyJrZXlzIjpbXX0="`
	// PrivateSigningKeys will hold private keys for signing JWTs
	PubPrivSigningKeySetB64JSON string `env:"PUB_PRIV_SIGNING_KEY_SET_B64_JSON" envDefault:"eyJrZXlzIjpbXX0="`
	// DefaultSignKeyID defines the default key id used for signing
	DefaultSignKeyID string `env:"DEFAULT_SIGN_KEY_ID" envDefault:"key1"`
	// DefaultEncryptKeyID defines the default encryption id used for encryption
	DefaultEncryptKeyID string `env:"DEFAULT_ENCRYPT_KEY_ID" envDefault:"key1"`
	// DefaultRedirectTo is the URL to redirect to if there is no state cookie
	DefaultRedirectTo *url.URL `env:"DEFAULT_REDIRECT_TO" envDefault:"http://localhost:8080/healthz"`
	// SessionCookiePath is the name he access token cookie will get.
	SessionCookieName string `env:"SESSION_COOKIE_NAME" envDefault:"cl_session"`
	// StateCookieName is the Name for the state nonce cookie will be set
	StateCookieName string `env:"AUTH_NONCE_COOKIE_NAME" envDefault:"cl_auth_state"`
	// AccessTokenCookiePath is the name he access token cookie will get.
	AccessTokenCookieName string `env:"ACCESS_TOKEN_COOKIE_NAME" envDefault:"cl_access_token"`
	// instructs the browser on how long the session cookie should be stored.
	SessionCookieMaxAgeSeconds int `env:"SESSION_COOKIE_MAX_AGE_SECONDS" envDefault:"34560000"`
	// instructs the browser on how long the access token cookie should be stored.
	AccessTokenCookieMaxAgeSeconds int `env:"ACCESS_TOKEN_COOKIE_MAX_AGE_SECONDS" envDefault:"34560000"`
	// allow username/password auth only for certain (system) users.
	BasicAuthWhitelist []string `env:"BASIC_AUTH_WHITELIST"`
}

// Keys hold our own private keys, and the WorkOS public keys.
type Keys struct {
	cfg    Config
	workos struct {
		public jwk.Set
	}
	signing struct {
		private jwk.Set
		public  jwk.Set
	}
	encrypt struct {
		private jwk.Set
		public  jwk.Set
	}
}

// NewKeys creates a new Keys instance.
func NewKeys(cfg Config) (*Keys, error) {
	keys := &Keys{cfg: cfg}

	// decode and parse the signing keys
	{
		dec, err := base64.StdEncoding.DecodeString(keys.cfg.PubPrivSigningKeySetB64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signing keys as base64: %w", err)
		}

		keys.signing.private, err = jwk.Parse(dec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signing encoded key set: %w", err)
		}

		keys.signing.public, err = jwk.PublicSetOf(keys.signing.private)
		if err != nil {
			return nil, fmt.Errorf("failed to get signing public key set: %w", err)
		}
	}

	// decode and parse the encrypt keys
	{
		dec, err := base64.StdEncoding.DecodeString(keys.cfg.PubPrivEncryptKeySetB64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encrypt keys as base64: %w", err)
		}

		keys.encrypt.private, err = jwk.Parse(dec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse encoded encrypt key set: %w", err)
		}

		keys.encrypt.public, err = jwk.PublicSetOf(keys.encrypt.private)
		if err != nil {
			return nil, fmt.Errorf("failed to get encrypt public key set: %w", err)
		}
	}

	return keys, nil
}

// start initializes the handler for async setup work.
func (keys *Keys) start(ctx context.Context) (err error) {
	keys.workos.public, err = jwk.Fetch(ctx, keys.cfg.JWKEndpoint+keys.cfg.MainClientID)
	if err != nil {
		return fmt.Errorf("failed to fetch WorkOS public keys: %w", err)
	}

	return nil
}
