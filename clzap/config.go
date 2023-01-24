package clzap

import (
	"strings"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

// Config configures the logging package
type Config struct {
	// Level configures the the minium logging level that will be captured
	Level zapcore.Level `env:"LEVEL" envDefault:"info"`
	// Outputs configures the zap outputs that will be opened for logging
	Outputs []string `env:"OUTPUTS" envDefault:"stderr"`
}

// parseConfig from environment
func parseConfig(o env.Options) (c Config, err error) {
	o.Prefix = strings.ToUpper(moduleName) + "_"
	return c, env.Parse(&c, o)
}
