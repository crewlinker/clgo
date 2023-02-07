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
func Handle[C context.Context](f func(C, http.ResponseWriter, *http.Request) error, os ...Option[C]) http.HandlerFunc {
	opts := applyOptions(os)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bw := NewBufferResponse(w, opts.bufLimit)
		defer bw.Free()

		ctx, err := opts.ctxBuilder(r)
		if err != nil {
			opts.errHandler(w, r, err)
			return
		}

		if opts.panicHandler != nil {
			defer func() {
				if v := recover(); v != nil {
					opts.panicHandler(ctx, w, r, v, opts.ctxErrHandler)
				}
			}()
		}

		if err := f(ctx, bw, r); err != nil {
			if resetErr := bw.Reset(); resetErr != nil {
				opts.ctxErrHandler(ctx, w, r, errors.Join(err, resetErr))
				return
			}

			opts.ctxErrHandler(ctx, w, r, err)
			return
		}

		if err := bw.ImplicitFlush(); err != nil {
			opts.ctxErrHandler(ctx, w, r, err)
			return
		}
	})
}
