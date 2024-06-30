package clservev3

// Middleware functions wrap each other to create unilateral functionality.
type Middleware[V any] func(Handler[V]) Handler[V]

// Use takes the inner handler h and wraps it with middleware. The order of wrapping
// follows the order in which the middlewares are provided here. The "most left" middleware
// provided is the inner-most middleware and the "most right" middleware is the outer-most middleware.
func Use[V any](h Handler[V], m ...Middleware[V]) Handler[V] {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	for i := range m {
		wrapped = m[i](wrapped)
	}

	return wrapped
}
