package clserve

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var (
	// ErrBufferFull is returned when the write call will cause the buffer to be filled beyond its limit.
	ErrBufferFull = errors.New("buffer is full")

	// ErrAlreadyFlushed is returned when the buffer is reset even though a handler already flushed the data
	// to the underlying writer.
	ErrAlreadyFlushed = errors.New("already flushed")
)

// ResponseBuffer is a http.ResponseWriter implementation that buffers writes up to configurable amount of
// bytes. This allows the implementation of handlers that can error halfway and return a
// completely different response instead of what was written before the error occurred.
type ResponseBuffer struct {
	resp              http.ResponseWriter
	buf               bytes.Buffer
	limit             int
	status            int
	headerFlushed     bool
	bodyFlushed       bool
	unflushableHeader http.Header
}

// responseBufferPool allows us to reuse some ResponseBuffer objects to
// conserve system resources.
var responseBufferPool = sync.Pool{ //nolint:gochecknoglobals
	New: func() interface{} { return new(ResponseBuffer) },
}

// NewBufferResponse inits a buffered response writer. It has a configurable limit after
// which the writing will return an error. This is to protect unchecked handlers from claiming
// too much memory. Limit can be set to -1 to disable this check.
func NewBufferResponse(resp http.ResponseWriter, limit int) *ResponseBuffer {
	w, _ := responseBufferPool.Get().(*ResponseBuffer)
	w.resp = resp
	w.limit = limit
	w.status = http.StatusOK

	return w
}

// Free resets all members of the ResponseBuffer and puts it back in the sync pool to
// allow it to be re-used for a possible next initilization. It should be called after
// the handling has completed and the buffer should not be used after.
func (w *ResponseBuffer) Free() {
	w.buf.Reset()
	w.resp = nil
	w.limit = 0
	w.status = 0
	w.headerFlushed = false
	w.bodyFlushed = false
	w.unflushableHeader = nil
	responseBufferPool.Put(w)
}

// WriteHeader will cause headers to be flushed to the underlying writer while calling WriteHeader
// on the underlying writer with the given status code.
func (w *ResponseBuffer) WriteHeader(statusCode int) {
	if w.headerFlushed {
		return // cannot set if header was already flushed
	}

	w.status = statusCode
	w.markHeaderAsFlushed()
}

// Header allows users to modify the headers (and trailers) sent to the client. The headers are not
// actually flushed to the underlying writer until a write or flush is being triggered.
func (w *ResponseBuffer) Header() http.Header {
	if w.headerFlushed {
		// to emulate the behaviour of the stdlib response writer we return a header that will never be
		// flushed. The extra allocation that this causes only influences bad usage of the response writer.
		if w.unflushableHeader == nil {
			w.unflushableHeader = make(http.Header)
		}

		return w.unflushableHeader
	}

	return w.resp.Header()
}

// Reset provides the differentiating feature from a regular ResponseWriter: it allows changing the
// response completely even if some data has been written already. This behaviour cannot be guaranteed
// if flush has been called explicitly. In that case it will return ErrAlreadyFlushed.
func (w *ResponseBuffer) Reset() error {
	if w.bodyFlushed {
		return errAlreadyFlushed()
	}

	for k := range w.resp.Header() {
		w.resp.Header().Del(k)
	}

	w.headerFlushed = false
	w.status = http.StatusOK
	w.buf.Reset()

	return nil
}

// markHeaderAsFlushed will mark the headers are being flushed to emulate the stdlib response writer
// behaviour.
func (w *ResponseBuffer) markHeaderAsFlushed() {
	w.headerFlushed = true
}

// Write appends the contents of p to the buffered response, growing the internal buffer as needed. If
// the write will cause the buffer be larger then the configure limit it will return ErrBufferFull.
func (w *ResponseBuffer) Write(buf []byte) (int, error) {
	if w.limit >= 0 && w.buf.Len()+len(buf) > w.limit {
		return 0, errBufferFull()
	}

	w.markHeaderAsFlushed()

	n, err := w.buf.Write(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to write underlying response: %w", err)
	}

	return n, nil
}

// ImplicitFlush flushes data to the underlying writer without calling .Flush on it by proxy. This is provided
// separately from FlushError to allow for emulating the original ResponseWriter behaviour more correctly.
func (w *ResponseBuffer) ImplicitFlush() error {
	w.markHeaderAsFlushed()
	w.resp.WriteHeader(w.status)

	_, err := w.buf.WriteTo(w.resp)
	if err != nil {
		return fmt.Errorf("failed to write underlying: %w", err)
	}

	w.bodyFlushed = true

	return nil
}

// FlushError any buffered bytes to the underlying response writer and resets the buffer. After flush has been
// called the response data should be considered sent and in-transport to the client.
func (w *ResponseBuffer) FlushError() error {
	if err := w.ImplicitFlush(); err != nil {
		return err
	}

	// calling flush on the underlying writer to make it explicit
	if err := http.NewResponseController(w.resp).Flush(); err != nil { //nolint:bodyclose
		return fmt.Errorf("failed to flush underlying: %w", err)
	}

	return nil
}

// Unwrap returns the underlying response writer. This is expected by the http.ResponseController to
// allow it to call appropriate optional interface implementations.
func (w *ResponseBuffer) Unwrap() http.ResponseWriter {
	return w.resp
}

// errBufferFull returns an error that Is ErrBufferFull but is not == to it.
func errBufferFull() error { return fmt.Errorf("%w", ErrBufferFull) }

// errAlreadyFlushed returns an error that Is ErrAlreadyFlushed but is not == to it.
func errAlreadyFlushed() error { return fmt.Errorf("%w", ErrAlreadyFlushed) }
