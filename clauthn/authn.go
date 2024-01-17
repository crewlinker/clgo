// Package clauthn provides re-usable authentication (AuthN).
package clauthn

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the package.
type Config struct {
	// PrivateSigningKeys will hold private keys for signing JWTs
	PubPrivKeySetB64JSON string `env:"PUB_PRIV_KEY_SET_B64_JSON"`
	// DefaultSignKeyID defines the default key id used for signing
	DefaultSignKeyID string `env:"DEFAULT_SIGN_KEY_ID" envDefault:"key1"`
}

// Authn provides authentication.
type Authn struct {
	cfg   Config
	logs  *zap.Logger
	clock jwt.Clock
	keys  struct {
		private jwk.Set
		public  jwk.Set
	}
}

// NewAuthn inits the Authn service.
func NewAuthn(cfg Config, logs *zap.Logger, clock jwt.Clock) (*Authn, error) {
	authn := &Authn{cfg: cfg, logs: logs.Named("authn"), clock: clock}

	dec, err := base64.URLEncoding.DecodeString(cfg.PubPrivKeySetB64JSON)
	if err != nil {
		return nil, fmt.Errorf("faield to decode keys as base64 padding url encoding: %w", err)
	}

	authn.keys.private, err = jwk.Parse(dec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encoded key set: %w", err)
	}

	authn.keys.public, err = jwk.PublicSetOf(authn.keys.private)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key set: %w", err)
	}

	return authn, nil
}

// SignJWT sings a JWT using the key set.
func (a *Authn) SignJWT(ctx context.Context, tok openid.Token) ([]byte, error) {
	sk, ok := a.keys.private.LookupKeyID(a.cfg.DefaultSignKeyID)
	if !ok {
		return nil, fmt.Errorf("no known secret key with key id '%s'", a.cfg.DefaultSignKeyID)
	}

	b, err := jwt.Sign(tok, jwt.WithKey(sk.Algorithm(), sk))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return b, nil
}

// AuthenticateJWT by parsing and validating the inp as a JSON web token (JWT).
func (a *Authn) AuthenticateJWT(ctx context.Context, inp []byte) (openid.Token, error) {
	tok, err := jwt.Parse(inp,
		jwt.WithToken(openid.New()),
		jwt.WithKeySet(a.keys.public),
		jwt.WithClock(a.clock),
		jwt.WithAcceptableSkew(time.Hour*1),
		jwt.WithValidate(true),
		jwt.WithVerify(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate token: %w", err)
	}

	oidt, ok := tok.(openid.Token)
	if !ok {
		return nil, fmt.Errorf("parsed token coun't be cast to openid.Token")
	}

	return oidt, nil
}

// moduleName for consistent component naming.
const moduleName = "clauthn"

// Provide the auth components as an fx dependency.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),

		// provide authentication service
		fx.Provide(NewAuthn),
		// provide a wall clock
		fx.Supply(fx.Annotate(jwt.ClockFunc(time.Now), fx.As(new(jwt.Clock)))),
	)
}

// TestProvide provides authn authz dependencies that are easy to use in
// tests.
func TestProvide() fx.Option {
	return fx.Options(
		Provide(),

		// set the configuration to have a signing key just for testing: mkjwk.org
		fx.Decorate(func(cfg Config) Config {
			cfg.PubPrivKeySetB64JSON = base64.URLEncoding.EncodeToString([]byte(`{
				"keys": [
					{
						"kty": "EC",
						"d": "v4FOE04_pQ0syWn8BSxSS3Seyq8IDhouKrkeRasbTHNjwHpS7JVq_gE_fqg_YnCA",
						"use": "sig",
						"crv": "P-384",
						"kid": "key1",
						"x": "lFTVVpHTLC_Gy2KhR4OKCNs31f-ww3kfyqUmJkc45PLTyX29TC_wYXWRnX2bIgwc",
						"y": "2G7GemEjzfMoZdt48ryOq7JOHgL_cqXjbFeGcwsyj1g2T_ZeeYU7gO51SGsiiiPg",
						"alg": "ES384"
					}
				]
			}`))

			return cfg
		}),
	)
}
