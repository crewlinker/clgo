package clservev2

import (
	"context"
	"net/http"
)

// ResponseWriter implements the http.ResponseWriter but the underlying bytes are buffered. This allows
// middleware to reset the writer and formulate a completely new response.
type ResponseWriter interface {
	http.ResponseWriter
	Reset()
}

// HandlerFunc mirrors the http.Handler but with a generic context type and error return.
type Handler[C context.Context] interface {
	ServeHTTP(ctx C, w ResponseWriter, r *http.Request) error
}

// HandlerFunc mirrors the http.HandlerFunc but with a generic context type and error return.
type HandlerFunc[C context.Context] func(C, ResponseWriter, *http.Request) error

func (f HandlerFunc[C]) ServeHTTP(ctx C, w ResponseWriter, r *http.Request) error {
	return f(ctx, w, r)
}

// ServeFunc takes a handler func and then calls [Serve].
func ServeFunc[C context.Context](
	hdlr HandlerFunc[C], os ...Option,
) http.Handler {
	return Serve[C](hdlr, os...)
}

// Serve takes a handler with a customizable context that is able to return an error. To support
// this the response is buffered until the handler is done. If an error occurs the buffer is discarded and
// a full replacement response can be formulated. The underlying buffer is re-used between requests for
// improved performance.
func Serve[C context.Context](
	hdlr Handler[C], os ...Option,
) http.Handler {
	opts := applyOptions(os)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		bresp := NewBufferResponse(resp, opts.bufLimit)
		defer bresp.Free()

		var initialCtx C

		if err := hdlr.ServeHTTP(initialCtx, bresp, req); err != nil {
			opts.logger.LogUnhandledServeError(err)

			return
		}

		if err := bresp.ImplicitFlush(); err != nil {
			opts.logger.LogImplicitFlushError(err)

			return
		}
	})
}
