package clserve

import "net/http"

// Middleware the de-facto type for middleware in the Go ecosystem.
type Middleware func(http.Handler) http.Handler

// Use turns a slice of middleware into wrapped calls. The left-most middleware
// will become the other middleware. 'h' will be come the inner handler.
func Use(h http.Handler, m ...Middleware) http.Handler {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}
