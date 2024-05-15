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

// ErrStateCookieNotPresentOrInvalid is returned when the state cookie is not present or invalid.
var ErrStateCookieNotPresentOrInvalid = errors.New("state cookie not present or invalid")

// ErrStateNonceMismatch is returned when the nonce from the query does not match the nonce from the state cookie.
var ErrStateNonceMismatch = errors.New("state nonce mismatch")

// RedirectToNotAllowedError is returned when the redirect URL is not allowed.
type RedirectToNotAllowedError struct {
	actual  string
	allowed []string
}

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

	// only show server errors in environment where it is safe. Errors from our outh
	// provider might leak too much information.
	message := http.StatusText(http.StatusInternalServerError)
	if h.cfg.ShowServerErrors {
		message = err.Error()
	}

	http.Error(w, message, http.StatusInternalServerError)
}

// WorkOSCallbackError is an error returned from the WorkOS callback logic.
type WorkOSCallbackError struct {
	code        string
	description string
}

func (e WorkOSCallbackError) Error() string {
	return fmt.Sprintf("callback with error from WorkOS: %s: %s", e.code, e.description)
}
