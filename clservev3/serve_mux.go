package clservev3

import "net/http"

// ServeMux is an extension to the standard http.ServeMux. It supports handling requests with a
// buffered response for error returns, typed context values and named routes.
type ServeMux[V any] struct {
	reverser    *Reverser
	middlewares struct {
		standard []StdMiddleware
		buffered []Middleware[V]
	}
	options []Option
	mux     *http.ServeMux
}

func NewServeMux[V any](opts ...Option) *ServeMux[V] {
	return &ServeMux[V]{
		reverser: NewReverser(),
		options:  opts,
		mux:      http.NewServeMux(),
	}
}

// Reverse a route 'name' with values for each parameter.
func (m *ServeMux[V]) Reverse(name string, vals ...string) (string, error) {
	return m.reverser.Reverse(name, vals...)
}

// Use will add a standard http middleware triggered for both buffered and unbuffered request handling.
func (m *ServeMux[V]) Use(mw ...StdMiddleware) {
	m.middlewares.standard = append(m.middlewares.standard, mw...)
}

// BUse will add a middleware ONLY for any buffered http handling, that is handlers setup using BHandle or BHandleFunc.
func (m *ServeMux[V]) BUse(mw ...Middleware[V]) {
	m.middlewares.buffered = append(m.middlewares.buffered, mw...)
}

// BHandle will invoke 'handler' with a buffered response for the named route and pattern.
func (m *ServeMux[V]) BHandle(name, pattern string, handler Handler[V]) {
	m.Handle(name, pattern, Serve(Use(handler, m.middlewares.buffered...), m.options...))
}

// BHandleFunc will invoke a handler func with a buffered response.
func (m *ServeMux[V]) BHandleFunc(name, pattern string, handler HandlerFunc[V]) {
	m.BHandle(name, pattern, handler)
}

// Handle will invoke 'handler' with a buffered response for the named route and pattern.
func (m *ServeMux[V]) Handle(name, pattern string, handler http.Handler) {
	m.mux.Handle(m.reverser.Named(name, pattern), UseStd(handler, m.middlewares.standard...))
}

// Handle will invoke 'handler' with a buffered response for the named route and pattern.
func (m *ServeMux[V]) HandleFunc(name, pattern string, handler http.HandlerFunc) {
	m.Handle(name, pattern, handler)
}

// ServeHTTP maxes the mux implement http.Handler.
func (m ServeMux[V]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}
