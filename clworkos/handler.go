// Package clworkos provides auth-flow middleware and endpoint using WorkOS.
package clworkos

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clserve"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the package components.
type Config struct {
	// WorkOS API key.
	APIKey string `env:"API_KEY"`
	// WorkOS main client ID.
	MainClientID string `env:"MAIN_CLIENT_ID"`
	// ShowServerErrors will show server errors to the client, should only be visible in development.
	ShowServerErrors bool `env:"SHOW_SERVER_ERRORS" envDefault:"false"`
	// JWKEndpoint is the endpoint for fetching the public key set for verifying the access key.
	JWKEndpoint string `env:"JWK_ENDPOINT" envDefault:"https://api.workos.com/sso/jwks/"`
	// PubPrivSigningKeySetB64JSON will hold private keys for JWE encryption
	PubPrivEncryptKeySetB64JSON string `env:"PUB_PRIV_ENCRYPT_KEY_SET_B64_JSON" envDefault:"eyJrZXlzIjpbXX0="`
}

// Handler provides WorkOS auth-flow functionality.
type Handler struct {
	cfg    Config
	logs   *zap.Logger
	users  *usermanagement.Client
	engine *Engine
	keys   struct {
		workos struct {
			public jwk.Set
		}
		encrypt struct {
			private jwk.Set
			public  jwk.Set
		}
	}

	http.Handler
}

// default http response buffer linit: 2MiB.
const bufferLimit = 2 * 1024 * 1024

// New creates a new Handler with the provided configuration and logger.
func New(cfg Config, logs *zap.Logger, engine *Engine) *Handler {
	mux := http.NewServeMux()
	hdlr := &Handler{
		cfg:     cfg,
		logs:    logs,
		Handler: mux,
		users:   usermanagement.NewClient(cfg.APIKey),
		engine:  engine,
	}

	serveOpts := []clserve.Option[context.Context]{
		clserve.WithBufferLimit[context.Context](bufferLimit),
		clserve.WithErrorHandling[context.Context](hdlr.handleError),
	}

	mux.Handle("/sign-in", clserve.Handle(hdlr.handleSignInStart(), serveOpts...))
	// mux.Handle("/callback", clserve.Handle(hdlr.callbackFromProvider(), serveOpts...))
	// mux.Handle("/logout", clserve.Handle(hdlr.logout(), serveOpts...))

	return hdlr
}

// start initializes the handler for async setup work.
func (h *Handler) start(ctx context.Context) (err error) {
	h.keys.workos.public, err = jwk.Fetch(ctx, h.cfg.JWKEndpoint+h.cfg.MainClientID)
	if err != nil {
		return fmt.Errorf("failed to fetch public keys: %w", err)
	}

	h.logs.Info("fetched WorkOS JWKS", zap.Int("num_of_keys", h.keys.workos.public.Len()))

	// decode and parse the encrypt keys
	{
		dec, err := base64.StdEncoding.DecodeString(h.cfg.PubPrivEncryptKeySetB64JSON)
		if err != nil {
			return fmt.Errorf("failed to decode encrypt keys as base64: %w", err)
		}

		h.keys.encrypt.private, err = jwk.Parse(dec)
		if err != nil {
			return fmt.Errorf("failed to parse encoded encrypt key set: %w", err)
		}

		h.keys.encrypt.public, err = jwk.PublicSetOf(h.keys.encrypt.private)
		if err != nil {
			return fmt.Errorf("failed to get encrypt public key set: %w", err)
		}

		h.logs.Info("read JWE encryption keys", zap.Int("num_of_keys", h.keys.encrypt.private.Len()))
	}

	return nil
}

// moduleName for naming conventions.
const moduleName = "webwos"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the rpc implementations
		fx.Provide(fx.Annotate(New,
			fx.OnStart(func(ctx context.Context, h *Handler) error { return h.start(ctx) }),
		)),
		// provide the real user management client
		fx.Provide(fx.Annotate(NewUserManagement, fx.As(new(UserManagement)))),
		// provide the engine
		fx.Provide(NewEngine),

		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}

// TestProvide provides the WorkOS handler with well-known (public) testing keys.
func TestProvide() fx.Option {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		//nolint:lll
		fmt.Fprintf(w, `{"keys":[{"alg":"RS256","kty":"RSA","use":"sig","n":"stfFLll3xBHX_4PAy4w082fV8SKoo2OomSfwv0tq9dZUZXsRtGBFztisK1CnyERnKx7Vr65JJI2s26EW4DlKZ8JqCEHLZEur1K8bEPr4A8H3Jq0iitlOfsdZgpi2EwWzzJxnHvqL-Mgy-l2eADmcunnttLM-xQzzZ3K_fLmlw6ztIINoTZQ_2VhiCK1DxkSZK3r9I5MhzVWrTcj5lajGjcHdnNpKFXL6X8CI7WOuj7f5qW52ibw0SDhb_dFxEI21Mdy4wN6nS2smNNhSz-Y1sSLYkWbOfC0ubNYBUJcgTu-V8fNK8eZz4AUnSRha4klhvbTlnbY2myLY4ybzGB5tuw","e":"AQAB","kid":"sso_oidc_key_pair_01HJT8QD5WB9WENVX0A8A36QAM","x5c":["MIICwDCCAaigAwIBAgIUOTxVEeNf5Y2v1VXFYX4HfN+ssCcwDQYJKoZIhvcNAQEFBQAwGjEYMBYGA1UEAxMPYXV0aC53b3Jrb3MuY29tMB4XDTIzMTIyOTA3NDgyM1oXDTI4MTIyOTA3NDgyM1owGjEYMBYGA1UEAxMPYXV0aC53b3Jrb3MuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAstfFLll3xBHX/4PAy4w082fV8SKoo2OomSfwv0tq9dZUZXsRtGBFztisK1CnyERnKx7Vr65JJI2s26EW4DlKZ8JqCEHLZEur1K8bEPr4A8H3Jq0iitlOfsdZgpi2EwWzzJxnHvqL+Mgy+l2eADmcunnttLM+xQzzZ3K/fLmlw6ztIINoTZQ/2VhiCK1DxkSZK3r9I5MhzVWrTcj5lajGjcHdnNpKFXL6X8CI7WOuj7f5qW52ibw0SDhb/dFxEI21Mdy4wN6nS2smNNhSz+Y1sSLYkWbOfC0ubNYBUJcgTu+V8fNK8eZz4AUnSRha4klhvbTlnbY2myLY4ybzGB5tuwIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQCjL4NTDUk5NBz8McMMfpqpRsYdFct2GEeIryktLxpB9Zl/50FN2rdsxmYxpp3E2HpTai3Zoiq1Lak7K8SPaJCbFcrj5UtfFgsiJyak6WLmUbMIWoLCvtFAfz8IkrX8/WV9MCFMMgdmoP2h2WCMp9f4qgQqhvM/99p9YQF0MJKeQgy8tz+LUIlNKhguVOhuGHFQb2OJInxhmK3BjuW4hU07b7ADKNPX1a5MVaZLTz/b9Z1EMEaAhQT3lGdzaYuqdZVDL/voMwcvm5V8HrX7U/8g0bIKFaRldvXjRTZW9EpbaPzAb9H4G8pWxariZMuO3YmQ6Nv/tg8N2RjjYxeRRUq9"],"x5t#S256":"RGCESoezgxWB3mn9fS7wiW9tG_RX6VDgAGDLT11j0cY"}]}`)
	}))

	return fx.Options(
		Provide(),
		fx.Decorate(func(c Config) Config {
			c.JWKEndpoint = srv.URL

			return c
		}),
	)
}
