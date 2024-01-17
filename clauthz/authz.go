// Package clauthz provides Authorization (AuthZ) functionality.
package clauthz

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"

	_ "embed"
)

// Config configures the package.
type Config struct {
	// id for the system that is unning OPA.
	OPASystemID string `env:"OPA_SYSTEM_ID" envDefault:"auth"`
}

//go:embed opa.yml
var cfg []byte

// Authz provides authn and authz functionality. It includes a simple web server that
// serves our policy bundle on a random port on localhost.
type Authz struct {
	cfg  Config
	bsrv BundleServer
	logs *zap.Logger
	opa  *sdk.OPA
	opaw *zapio.Writer
}

// NewAuthz inits the auth service.
func NewAuthz(cfg Config, logs *zap.Logger, bsrv BundleServer) (a *Authz, err error) {
	return &Authz{bsrv: bsrv, cfg: cfg, logs: logs.Named("authz"), opaw: &zapio.Writer{Log: logs.Named("opa")}}, nil
}

// Start the auth service.
func (a *Authz) Start(ctx context.Context) (err error) {
	// pass ologs to zap
	ologs := logging.New()
	ologs.SetOutput(a.opaw)

	// setup OPA with a config with the service_url replaced
	a.opa, err = sdk.New(ctx, sdk.Options{
		ID:            a.cfg.OPASystemID,
		Config:        bytes.NewReader(bytes.ReplaceAll(cfg, []byte(`$SERVICE_URL$`), []byte(a.bsrv.URL()))),
		Logger:        ologs,
		ConsoleLogger: ologs,
	})
	if err != nil {
		return fmt.Errorf("failed to init opa: %w", err)
	}

	return nil
}

// Stop the auth service.
func (a *Authz) Stop(ctx context.Context) (err error) {
	if err := a.opaw.Close(); err != nil {
		return fmt.Errorf("failed to close zap writer: %w", err)
	}

	return
}

// IsAuthorized the user for a given setup.
func (a *Authz) IsAuthorized(ctx context.Context, inp any) (bool, error) {
	res, err := a.opa.Decision(ctx, sdk.DecisionOptions{
		Path:  "/authz/allow",
		Input: inp,
	})
	if err != nil {
		return false, fmt.Errorf("failed to decide: %w", err)
	}

	allow, ok := res.Result.(bool)
	if !ok {
		return false, fmt.Errorf("decision did not return bool, but: %T", res.Result)
	}

	return allow, nil
}

// moduleName for consistent component naming.
const moduleName = "clauthz"

// AllowAll policy always returns allow, for testing.
func AllowAll() map[string]string {
	return map[string]string{
		"main.rego": `
			package authz
			import rego.v1		
			default allow := true
		`,
	}
}

// Provide the auth components as an fx dependency.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the webos webhooks client
		fx.Provide(fx.Annotate(NewAuthz,
			fx.OnStart(func(ctx context.Context, a *Authz) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *Authz) error { return a.Stop(ctx) }),
		)),
	)
}

// BundleProvide provides a bundle server.
func BundleProvide(bfs fs.FS) fx.Option {
	return fx.Options(
		fx.Supply(BundleFS{FS: bfs}),
		fx.Provide(fx.Annotate(NewFSBundles,
			fx.OnStart(func(ctx context.Context, a *FSBundles) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *FSBundles) error { return a.Stop(ctx) }),
		)),
		fx.Provide(func(bs *FSBundles) BundleServer { return bs }),
	)
}

// TestProvide provides authn authz dependencies that are easy to use in
// tests.
func TestProvide(policies map[string]string) fx.Option {
	return fx.Options(
		Provide(),

		// supply the policies
		fx.Supply(MockBundle(policies)),

		// provide a bundle server that is easy to use in tests.
		fx.Provide(fx.Annotate(NewMockBundles,
			fx.OnStart(func(ctx context.Context, a *MockBundles) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *MockBundles) error { return a.Stop(ctx) }),
		)),

		// provide as bundle server
		fx.Provide(func(bs *MockBundles) BundleServer { return bs }),
	)
}
