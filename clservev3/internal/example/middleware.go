// Package example implements example middleware in an outside package.
package example

import (
	"log/slog"
	"net/http"

	"github.com/crewlinker/clgo/clservev3"
)

// values types the context that needs to be passed. It forces implementations to implement a method
// that allows setting the logger.
type values[V any] interface {
	WithLogger(logs *slog.Logger) V
}

// ctxKey type scopes middlware values.
type ctxKey string

// Middleware provides an example for middleware that adds a logger to the context.
func Middleware[V values[V]](logs *slog.Logger) clservev3.Middleware[V] {
	return func(n clservev3.Handler[V]) clservev3.Handler[V] {
		return clservev3.HandlerFunc[V](func(c *clservev3.Context[V], w clservev3.ResponseWriter, r *http.Request) error {
			logs := logs.With(slog.String("method", r.Method))

			c.V = c.V.WithLogger(logs)                  // set on the typed values of the context
			c, r = c.WithValue(r, ctxKey("slog"), logs) // set on the untyped values of the context

			return n.ServeBHTTP(c, w, r)
		})
	}
}
