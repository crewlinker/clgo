package clserve

import (
	"errors"
	"fmt"
	"net/http"
)

type codedError struct {
	status int
	err    error
}

func (e codedError) Unwrap() error {
	return e.err
}

func (e codedError) Error() string {
	return fmt.Sprintf("%d: %v", e.status, e.err)
}

// NewError inits a new coded error.
func NewError(status int, err error) error {
	return codedError{status: status, err: err}
}

// Errorf formats an error with a http status code.
func Errorf(status int, format string, vals ...any) error {
	return NewError(status, fmt.Errorf(format, vals...)) //nolint:goerr113
}

// StandardErrorHandler implements an error handler with common behaviour. It allow configuring if real
// error messages are printed and it sets the status code for errors created with NewError, or Errorf.
// The errf has the same signature as http.Error and it's common to pass it here.
func StandarErrorHandler(
	showRealMessages bool,
	errf func(w http.ResponseWriter, msg string, code int),
) NoContextErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var (
			cerr codedError
			msg  string
		)

		ok, status := errors.As(err, &cerr), http.StatusInternalServerError
		if !ok {
			msg = err.Error()
		} else {
			msg = cerr.Unwrap().Error()
			status = cerr.status
		}

		if !showRealMessages {
			msg = http.StatusText(status)
		}

		errf(w, msg, status)
	}
}
