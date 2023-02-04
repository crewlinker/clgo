package clserve

import (
	"context"
	"errors"
	"net/http"
)

var (
	// ErrContextBuilderRequired is thrown when the context cannot be cased from context.Context
	ErrContextBuilderRequired = errors.New("context builder required")
)

// Option allows for customization of the (buffered) handling logic.
type Option[C context.Context] func(*opts[C])

// WithBufferLimit allows limiting the buffered writer (if it buffered). This can protect against
// buffered response writers taking up too much memory per response.
func WithBufferLimit[C context.Context](v int) Option[C] {
	return func(o *opts[C]) { o.bufLimit = v }
}

// WithContextBuilder customizes how a request is turned into a context for the handlers.
func WithContextBuilder[C context.Context](f ContextBuilderFunc[C]) Option[C] {
	return func(o *opts[C]) { o.ctxBuilder = f }
}

// WithErrorHandling allows for custom logic to handle the error returned from the Handler. In case the
// buffer's reset fails that error is joined with the original error and together are passed in as 'err'.
// In case the flushing fails (i.e: writing data to the underlying ResponseWriter) it is also passed in
// but it is unlikely to be useable for writing so that case is mosly usefull for logging purposes.
// When context building failed the first context argument 'c' might be nil.
func WithErrorHandling[C context.Context](f ErrorHandlerFunc[C]) Option[C] {
	return func(o *opts[C]) { o.errHandler = f }
}

// ErrorHandlerFunc can be configured across handlers to standardize how handling errors are handled
type ErrorHandlerFunc[C context.Context] func(c C, w http.ResponseWriter, r *http.Request, err error)

// ContextBuilderFunc describes the signature of a context builder
type ContextBuilderFunc[C context.Context] func(r *http.Request) (C, error)

// opts for the handling
type opts[C context.Context] struct {
	ctxBuilder ContextBuilderFunc[C]
	errHandler ErrorHandlerFunc[C]
	bufLimit   int
}

// defaultErrHandler simply returns an internal server error when an error occured
func defaultErrHandler[C context.Context](c C, w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// defeaultCtxBuilder builds a context from the request
func defaultCtxBuilder[C context.Context](r *http.Request) (C, error) {
	c, ok := r.Context().(C)
	if !ok {
		// in case we cannot just cast from the regular context.Context to our custom context
		// the user must provide a builder, else it panics.
		panic(ErrContextBuilderRequired)
	}

	return c, nil
}

// applyOptions applies options and sets sensible default
func applyOptions[C context.Context](opts []Option[C]) (o opts[C]) {
	o.bufLimit = -1
	o.errHandler = defaultErrHandler[C]
	o.ctxBuilder = defaultCtxBuilder[C]
	for _, opt := range opts {
		opt(&o)
	}
	return
}
