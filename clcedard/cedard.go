// Package clcedard provides components for the cedard authorization service.
package clcedard

import (
	"net/http"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the package.
type Config struct {
	// BaseURL configures the base url of the cedard service.
	BaseURL string `env:"BASE_URL" envDefault:"https://authz.crewlinker.com"`
	// JWTSigningSecret configures the secret for signing JWTs.
	JWTSigningSecret string `env:"JWT_SIGNING_SECRET" envDefault:"some-secret-for-testing"`
	// BackoffMaxElapsedTime configures the max elapsed time for the retry mechanism.
	BackoffMaxElapsedTime time.Duration `env:"BACKOFF_MAX_ELAPSED_TIME" envDefault:"3s"`
}

// moduleName standardizes the module name.
const moduleName = "clcedard"

// Client implements a client for the cedard authorization service.
type Client struct {
	cfg  Config
	logs *zap.Logger
	htcl *http.Client
}

// NewClient inits the client.
func NewClient(cfg Config, logs *zap.Logger, htcl *http.Client) *Client {
	return &Client{
		cfg:  cfg,
		logs: logs,
		htcl: htcl,
	}
}

// Provide dependencies.
func Provide() fx.Option {
	return fx.Module(moduleName,
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		fx.Provide(fx.Annotate(NewClient)),
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}
