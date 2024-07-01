// Package clworkos provides auth-flow middleware and endpoint using WorkOS.
package clworkos

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/advdv/bhttp"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clworkos/clworkosmock"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handler provides WorkOS auth-flow functionality.
type Handler struct {
	cfg    Config
	logs   *zap.Logger
	engine *Engine

	http.Handler
}

// default http response buffer linit: 2MiB.
const bufferLimit = 2 * 1024 * 1024

// C describes typed context values.
type C struct{}

// New creates a new Handler with the provided configuration and logger.
func New(cfg Config, logs *zap.Logger, engine *Engine) *Handler {
	mux := bhttp.NewServeMux[C](bhttp.WithBufferLimit(bufferLimit))
	hdlr := &Handler{
		cfg:     cfg,
		logs:    logs,
		Handler: mux,
		engine:  engine,
	}

	mux.BUse(hdlr.errorMiddleware())
	mux.BHandle("/sign-in", hdlr.handleSignIn())
	mux.BHandle("/sign-up", hdlr.handleSignUp())
	mux.BHandle("/sign-out", hdlr.handleSignOut())
	mux.BHandle("/callback", hdlr.handleCallback())

	return hdlr
}

// handleSignUp handles the sign-up flow.
func (h *Handler) handleSignUp() bhttp.HandlerFunc[C] {
	return func(ctx *bhttp.Context[C], w bhttp.ResponseWriter, r *http.Request) error {
		loc, err := h.engine.StartAuthenticationFlow(ctx, w, r, "sign-up")
		if err != nil {
			return fmt.Errorf("failed to start sign-up flow: %w", err)
		}

		http.Redirect(w, r, loc.String(), http.StatusFound)

		return nil
	}
}

// handleSignIn handles the sign-in flow.
func (h *Handler) handleSignIn() bhttp.HandlerFunc[C] {
	return func(ctx *bhttp.Context[C], w bhttp.ResponseWriter, r *http.Request) error {
		loc, err := h.engine.StartAuthenticationFlow(ctx, w, r, "sign-in")
		if err != nil {
			return fmt.Errorf("failed to start sign-in flow: %w", err)
		}

		http.Redirect(w, r, loc.String(), http.StatusFound)

		return nil
	}
}

// handleCallback handles the callback from WorkOS.
func (h *Handler) handleCallback() bhttp.HandlerFunc[C] {
	return func(ctx *bhttp.Context[C], w bhttp.ResponseWriter, r *http.Request) error {
		loc, err := h.engine.HandleSignInCallback(ctx, w, r)
		if err != nil {
			return fmt.Errorf("failed to handle sign-in callback: %w", err)
		}

		http.Redirect(w, r, loc.String(), http.StatusFound)

		return nil
	}
}

// handleSignOut handles the sign-in flow.
func (h *Handler) handleSignOut() bhttp.HandlerFunc[C] {
	return func(ctx *bhttp.Context[C], w bhttp.ResponseWriter, r *http.Request) error {
		loc, err := h.engine.StartSignOutFlow(ctx, w, r)
		if err != nil {
			return fmt.Errorf("failed to start sign-out flow: %w", err)
		}

		http.Redirect(w, r, loc.String(), http.StatusFound)

		return nil
	}
}

// moduleName for naming conventions.
const moduleName = "clworkos"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the rpc implementations
		fx.Provide(fx.Annotate(New), fx.Annotate(NewLister)),
		// provide the real user management client
		fx.Provide(fx.Annotate(NewUserManagement, fx.As(new(UserManagement)))),
		// provide the real user management client
		fx.Provide(fx.Annotate(NewOrganizations, fx.As(new(Organizations)))),
		// provide the engine
		fx.Provide(fx.Annotate(NewEngine, fx.ParamTags(``, ``, `optional:"true"`))),
		// provide the keys
		fx.Provide(fx.Annotate(NewKeys, fx.OnStart(func(ctx context.Context, k *Keys) error { return k.start(ctx) }))),
		// provide time.Now as the wall-clock time
		fx.Supply(fx.Annotate(jwt.ClockFunc(time.Now), fx.As(new(Clock)))),
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}

// TestProvide provides the WorkOS handler with well-known (public) testing keys. It will also Mock
// the WorkOS API client and put the wall-clock on a fixed point in time.
func TestProvide(tb testing.TB, clockAt int64) fx.Option {
	tb.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		//nolint:lll
		fmt.Fprintf(w, `{"keys":[{"alg":"RS256","kty":"RSA","use":"sig","n":"stfFLll3xBHX_4PAy4w082fV8SKoo2OomSfwv0tq9dZUZXsRtGBFztisK1CnyERnKx7Vr65JJI2s26EW4DlKZ8JqCEHLZEur1K8bEPr4A8H3Jq0iitlOfsdZgpi2EwWzzJxnHvqL-Mgy-l2eADmcunnttLM-xQzzZ3K_fLmlw6ztIINoTZQ_2VhiCK1DxkSZK3r9I5MhzVWrTcj5lajGjcHdnNpKFXL6X8CI7WOuj7f5qW52ibw0SDhb_dFxEI21Mdy4wN6nS2smNNhSz-Y1sSLYkWbOfC0ubNYBUJcgTu-V8fNK8eZz4AUnSRha4klhvbTlnbY2myLY4ybzGB5tuw","e":"AQAB","kid":"sso_oidc_key_pair_01HJT8QD5WB9WENVX0A8A36QAM","x5c":["MIICwDCCAaigAwIBAgIUOTxVEeNf5Y2v1VXFYX4HfN+ssCcwDQYJKoZIhvcNAQEFBQAwGjEYMBYGA1UEAxMPYXV0aC53b3Jrb3MuY29tMB4XDTIzMTIyOTA3NDgyM1oXDTI4MTIyOTA3NDgyM1owGjEYMBYGA1UEAxMPYXV0aC53b3Jrb3MuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAstfFLll3xBHX/4PAy4w082fV8SKoo2OomSfwv0tq9dZUZXsRtGBFztisK1CnyERnKx7Vr65JJI2s26EW4DlKZ8JqCEHLZEur1K8bEPr4A8H3Jq0iitlOfsdZgpi2EwWzzJxnHvqL+Mgy+l2eADmcunnttLM+xQzzZ3K/fLmlw6ztIINoTZQ/2VhiCK1DxkSZK3r9I5MhzVWrTcj5lajGjcHdnNpKFXL6X8CI7WOuj7f5qW52ibw0SDhb/dFxEI21Mdy4wN6nS2smNNhSz+Y1sSLYkWbOfC0ubNYBUJcgTu+V8fNK8eZz4AUnSRha4klhvbTlnbY2myLY4ybzGB5tuwIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQCjL4NTDUk5NBz8McMMfpqpRsYdFct2GEeIryktLxpB9Zl/50FN2rdsxmYxpp3E2HpTai3Zoiq1Lak7K8SPaJCbFcrj5UtfFgsiJyak6WLmUbMIWoLCvtFAfz8IkrX8/WV9MCFMMgdmoP2h2WCMp9f4qgQqhvM/99p9YQF0MJKeQgy8tz+LUIlNKhguVOhuGHFQb2OJInxhmK3BjuW4hU07b7ADKNPX1a5MVaZLTz/b9Z1EMEaAhQT3lGdzaYuqdZVDL/voMwcvm5V8HrX7U/8g0bIKFaRldvXjRTZW9EpbaPzAb9H4G8pWxariZMuO3YmQ6Nv/tg8N2RjjYxeRRUq9"],"x5t#S256":"RGCESoezgxWB3mn9fS7wiW9tG_RX6VDgAGDLT11j0cY"}]}`)
	}))

	umMock := clworkosmock.NewMockUserManagement(tb)
	orgMock := clworkosmock.NewMockOrganizations(tb)

	if clockAt == 0 {
		clockAt = time.Now().Unix()
	}

	return fx.Options(
		Provide(),
		// supply our mock implementation
		fx.Supply(umMock, orgMock),
		// replace the real user management client with the mock
		fx.Decorate(func(UserManagement) UserManagement { return umMock }),
		// replace the real orgs client with the mock
		fx.Decorate(func(Organizations) Organizations { return orgMock }),
		// fix the wall-clock at the given time
		fx.Decorate(func(Clock) Clock { return jwt.ClockFunc(func() time.Time { return time.Unix(clockAt, 0) }) }),
		// setup config that allows for testing
		fx.Decorate(func(c Config) Config {
			//nolint:lll
			c.PubPrivSigningKeySetB64JSON = "ewogICAgImtleXMiOiBbCiAgICAgICAgewogICAgICAgICAgICAia3R5IjogIkVDIiwKICAgICAgICAgICAgImQiOiAiUWowRGtZZnFFNGpKcUI3aVBoQ25zUUpldDJwbzMwMTRPWEFYek9oWlNkcyIsCiAgICAgICAgICAgICJ1c2UiOiAic2lnIiwKICAgICAgICAgICAgImNydiI6ICJQLTI1NiIsCiAgICAgICAgICAgICJraWQiOiAia2V5MSIsCiAgICAgICAgICAgICJ4IjogIjZXb1lqdFB2MUVieVBwcXpkaG41c1RjeXhIbkRTNmhnb3kxYUo2aVpWQWMiLAogICAgICAgICAgICAieSI6ICJmUUdvVW5kdVJOVFB6QzNLblJsdjh3Y3JnaGY5YzFCSDdCZERtNUVFV0c4IiwKICAgICAgICAgICAgImFsZyI6ICJFUzI1NiIKICAgICAgICB9CiAgICBdCn0K"
			//nolint:lll
			c.PubPrivEncryptKeySetB64JSON = "ewogICAgImtleXMiOiBbCiAgICAgICAgewogICAgICAgICAgICAia3R5IjogIkVDIiwKICAgICAgICAgICAgImQiOiAiNERXcURtaWZkcXN1M0FKWF9rY1pZdER3QTF5cERfWFkyNHN2REFxdlY0ayIsCiAgICAgICAgICAgICJ1c2UiOiAiZW5jIiwKICAgICAgICAgICAgImNydiI6ICJQLTI1NiIsCiAgICAgICAgICAgICJraWQiOiAia2V5MSIsCiAgICAgICAgICAgICJ4IjogIkxhUUZfTmxkWXRNTVJUWjl0QmM5SFB3SkRJQTUxVkNNREdiUXlVeFRMLTgiLAogICAgICAgICAgICAieSI6ICI3M1BLMVk2VktCS185X1ltMVdZUHlvZmYwSnM1dDdUaUxJU1ZEV0NFanJvIiwKICAgICAgICAgICAgImFsZyI6ICJFQ0RILUVTK0ExMjhLVyIKICAgICAgICB9CiAgICBdCn0K"
			c.JWKEndpoint = srv.URL
			c.ShowErrorMessagesToClient = true
			c.RedirectToAllowedHosts = []string{"localhost", "*.foo.com", "*--atsdash.netlify.app"}

			c.BasicAuthWhitelist = []string{"admin+system1@crewlinker.com"}

			return c
		}),
	)
}
