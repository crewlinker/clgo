package clservev2

import (
	"context"
	"fmt"
	"net/http"
)

// Middleware functions wrap each other to create unilateral functionality.
type Middleware[C context.Context] func(Handler[C]) Handler[C]

// Use takes the inner handler h and wraps it with middleware. The order of wrapping
// follows the order in which the middlewares are provided here. The "most left" middleware
// provided is the inner-most middleware.
func Use[C context.Context](h Handler[C], m ...Middleware[C]) Handler[C] {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	for i := range m {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

// Errorer middleware will reset the buffered response, and return a server error.
func Errorer[C context.Context]() Middleware[C] {
	return func(next Handler[C]) Handler[C] {
		return HandlerFunc[C](func(c C, w ResponseWriter, r *http.Request) error {
			err := next.ServeHTTP(c, w, r)
			if err != nil {
				w.Reset()
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return nil
		})
	}
}

// Recover middleware. It will recover any panics and turn it into an error.
func Recoverer[C context.Context]() Middleware[C] {
	return func(next Handler[C]) Handler[C] {
		return HandlerFunc[C](func(c C, w ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("recovered: %v", e) //nolint:goerr113
				}
			}()

			return next.ServeHTTP(c, w, r)
		})
	}
}
