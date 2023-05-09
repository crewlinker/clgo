// Package clserve provides buffered HTTP response serving to support error handling
package clserve

import (
	"context"
	"errors"
	"net/http"
)

// Handle takes a handler function with a customizable context that is able to return an error. To support
// this the response is buffered until the handler is done. If an error occurs the buffer is discarded and
// a full replacement response can be formulated. The underlying buffer is re-used between requests for
// improved performance.
func Handle[C context.Context](
	hdlf func(C, http.ResponseWriter, *http.Request) error, os ...Option[C],
) http.HandlerFunc {
	opts := applyOptions(os)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		bresp := NewBufferResponse(resp, opts.bufLimit)
		defer bresp.Free()

		ctx, err := opts.ctxBuilder(req)
		if err != nil {
			opts.errHandler(resp, req, err)

			return
		}

		if opts.panicHandler != nil {
			defer func() {
				if v := recover(); v != nil {
					opts.panicHandler(ctx, resp, req, v, opts.ctxErrHandler)
				}
			}()
		}

		if err := hdlf(ctx, bresp, req); err != nil {
			if resetErr := bresp.Reset(); resetErr != nil {
				opts.ctxErrHandler(ctx, resp, req, errors.Join(err, resetErr))

				return
			}

			opts.ctxErrHandler(ctx, resp, req, err)

			return
		}

		if err := bresp.ImplicitFlush(); err != nil {
			opts.ctxErrHandler(ctx, resp, req, err)

			return
		}
	})
}
