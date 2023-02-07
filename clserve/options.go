package clserve

import (
	"context"
	"errors"
	"fmt"
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

// WithContextBuilder customizes how a request is turned into a context for the handlers. The builder func
// is not provided a ResponseWriter to force the user to not write anything as the buffer will not be reset
// when an error is returned.
func WithContextBuilder[C context.Context](f ContextBuilderFunc[C]) Option[C] {
	return func(o *opts[C]) { o.ctxBuilder = f }
}

// WithContextErrorHandling allows for custom logic to handle the error returned from the Handler in case
// there is a context available. By default the the handle just calls the non-context ErrorHandler. It
// allows full customization of the error response (including headers and status code) but the error might
// be due to a failture to reset the buffer or because writing to the underlying ResponseWriter has failed.
// In those cases it might not be possible to guarantee a complete error response can be returned.
func WithContextErrorHandling[C context.Context](f ErrorHandlerFunc[C]) Option[C] {
	return func(o *opts[C]) {
		o.ctxErrHandler = f
	}
}

// WithErrorHandling allows for custom error handling in case there is no context available. This
// happens when the context builder has failed for some reason.
func WithErrorHandling[C context.Context](f NoContextErrorHandlerFunc) Option[C] {
	return func(o *opts[C]) {
		o.errHandler = f
	}
}

// WithPanicHandler configures how exeptions in the handling function are caught. By default, the recovered
// panic is send to the error handler but it can be customized for example to provide custom logging.
// Additionally the panic handler can be set to nil to disable recovering of panics altogether. This can be
// done because it is also common to have middleware recover any panics, which has the advantage being able
// to catch panics in the middleware chain as well.
//
// Panics in the context builder are not caught by this logic and should be taken care of in the context
// builder itself. But panics in error handling with a context are caught by this handler.
func WithPanicHandler[C context.Context](f PanicHandlerFunc[C]) Option[C] {
	return func(o *opts[C]) { o.panicHandler = f }
}

// ErrorHandlerFunc can be configured across handlers to standardize how handling errors are handled
type ErrorHandlerFunc[C context.Context] func(c C, w http.ResponseWriter, r *http.Request, err error)

// NoContextErrorHandlerFunc can be configured across handlers to standardize how handling errors in case
// context building has failed.
type NoContextErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// PanicHandlerFunc can be configured across handlers to standardize how panics are handled
type PanicHandlerFunc[C context.Context] func(c C, w http.ResponseWriter, r *http.Request, v any, errh ErrorHandlerFunc[C])

// ContextBuilderFunc describes the signature of a context builder
type ContextBuilderFunc[C context.Context] func(r *http.Request) (C, error)

// opts for the handling
type opts[C context.Context] struct {
	ctxBuilder    ContextBuilderFunc[C]
	errHandler    NoContextErrorHandlerFunc
	ctxErrHandler ErrorHandlerFunc[C]
	panicHandler  PanicHandlerFunc[C]
	bufLimit      int
}

// defaultErrHandler simply returns an internal server error when an error occured
func defaultErrHandler(w http.ResponseWriter, r *http.Request, err error) {
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

// defaultPanicHandler simply defers the panic to the error handler
func defaultPanicHandler[C context.Context](
	c C,
	w http.ResponseWriter,
	r *http.Request,
	v any,
	errh ErrorHandlerFunc[C],
) {
	errh(c, w, r, fmt.Errorf("%v", v))
}

// applyOptions applies options and sets sensible default
func applyOptions[C context.Context](opts []Option[C]) (o opts[C]) {
	o.bufLimit = -1
	o.errHandler = defaultErrHandler
	o.ctxErrHandler = func(c C, w http.ResponseWriter, r *http.Request, err error) {
		o.errHandler(w, r, err) // by default, just call the no-context variant
	}
	o.ctxBuilder = defaultCtxBuilder[C]
	o.panicHandler = defaultPanicHandler[C]
	for _, opt := range opts {
		opt(&o)
	}
	return
}
