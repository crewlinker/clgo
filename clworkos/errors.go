package clworkos

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/zap"
)

// ErrRedirectToNotProvided is returned when the redirect_to query parameter is missing.
var ErrRedirectToNotProvided = errors.New("missing redirect_to query parameter")

// ErrCallbackCodeNotProvided is returned when the code query parameter is missing.
var ErrCallbackCodeNotProvided = errors.New("missing code query parameter")

// ErrStateNonceMismatch is returned when the nonce from the query does not match the nonce from the state cookie.
var ErrStateNonceMismatch = errors.New("state nonce mismatch")

// ErrNoAuthentication is returned when no authentication is present in the request.
var ErrNoAuthentication = errors.New("no authentication")

// ErrNoAccessTokenForSignOut is returned when no access token is present for sign out.
var ErrNoAccessTokenForSignOut = errors.New("no credentials to sign-out with")

// InvalidInputError can wrap any error to mark it as being invalid input.
type InvalidInputError struct{ error }

func (e InvalidInputError) isInvalidInput() {}

// InputErrorf is a helper function to create a new InvalidInputError.
func InputErrorf(format string, args ...interface{}) InvalidInputError {
	return InvalidInputError{fmt.Errorf(format, args...)} //nolint:goerr113
}

// RedirectToNotAllowedError is returned when the redirect URL is not allowed.
type RedirectToNotAllowedError struct {
	actual  string
	allowed []string
}

func (e RedirectToNotAllowedError) isInvalidInput() {}
func (e RedirectToNotAllowedError) Error() string {
	return fmt.Sprintf("redirect URL is not allowed: %q, allowed hosts: %v", e.actual, e.allowed)
}

// KeyNotFoundError is returned when a key with a specific id is not found.
type KeyNotFoundError struct {
	id string
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("key with id %q not found", e.id)
}

// handleErrors will make sure that any errors that are caught will be returned to the client.
func (h Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	clzap.Log(r.Context(), h.logs).Error("error while serving request", zap.Error(err))

	// try to be smart about the status code
	status := http.StatusInternalServerError
	if IsBadRequestError(err) {
		status = http.StatusBadRequest
	}

	// only show server errors in environment where it is safe. Errors from our outh
	// provider might leak too much information.
	message := http.StatusText(status)
	if h.cfg.ShowErrorMessagesToClient {
		message = err.Error()
	}

	http.Error(w, message, status)
}

// WorkOSCallbackError is an error returned from the WorkOS callback logic.
type WorkOSCallbackError struct {
	code        string
	description string
}

func (e WorkOSCallbackError) Error() string {
	return fmt.Sprintf("callback with error from WorkOS: %s: %s", e.code, e.description)
}

// IsBadRequestError will return true if the error is an error that describes bad input from the user. this is useful
// to provide the client with more information about what went wrong.
func IsBadRequestError(err error) bool {
	var isInvalidInput interface{ isInvalidInput() }
	if errors.As(err, &isInvalidInput) {
		return true
	}

	switch {
	case errors.Is(err, ErrRedirectToNotProvided),
		errors.Is(err, ErrNoAccessTokenForSignOut),
		errors.Is(err, ErrCallbackCodeNotProvided),
		errors.Is(err, ErrStateNonceMismatch):
		return true

	default:
		return false
	}
}
