package clauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	"go.uber.org/zap"
)

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

	dec, err := base64.URLEncoding.DecodeString(cfg.AuthnPubPrivKeySetB64JSON)
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
	sk, ok := a.keys.private.LookupKeyID(a.cfg.AuthnDefaultSignKeyID)
	if !ok {
		return nil, fmt.Errorf("no known secret key with key id '%s'", a.cfg.AuthnDefaultSignKeyID)
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
