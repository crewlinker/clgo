// Package clbuildinfo provides the official AWS SDK (v2)
package clbuildinfo

import (
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures this package.
type Config struct{}

// Info provides build-time information to the rest of the application.
type Info struct {
	cfg Config

	version string
}

// New initializes the build info component.
func New(cfg Config, version string) *Info {
	return &Info{
		cfg:     cfg,
		version: version,
	}
}

// Version as determined at build time.
func (in Info) Version() string {
	return in.version
}

// moduleName for naming conventions.
const moduleName = "clbuildinfo"

// Provide configures the DI for providng database connectivity.
func Provide(version string) fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the actual aws config
		fx.Supply(fx.Annotate(version, fx.ResultTags(`name:"version"`))),
		// provide the build info
		fx.Provide(fx.Annotate(New, fx.ParamTags(``, `name:"version"`))),
	)
}

// TestProvide provides di for testing where no specific version is required to be provided.
func TestProvide() fx.Option {
	return Provide("v0.0.0-test")
}
