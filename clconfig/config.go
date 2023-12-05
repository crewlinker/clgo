// Package clconfig provides re-usable configuration utilities
package clconfig

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"go.uber.org/fx"
)

// EnvConfigurer returns a function that parses environment variables into a configuration struct T. If the
// prefix is provided it will set a prefix for the underlying environment parser.
func EnvConfigurer[T any](prefix ...string) func(o env.Options) (T, error) {
	return func(envo env.Options) (T, error) {
		var cfg T

		if len(prefix) > 0 {
			envo.Prefix = prefix[0]
		}

		if err := env.ParseWithOptions(&cfg, envo); err != nil {
			return cfg, fmt.Errorf("failed to parse environment: %w", err)
		}

		return cfg, nil
	}
}

// Provide configuration T as an fx dependency that parses the environment with an optional prefix.
func Provide[T any](prefix ...string) fx.Option {
	return fx.Provide(fx.Annotate(
		EnvConfigurer[T](prefix...),
		fx.ParamTags(`optional:"true"`)))
}
