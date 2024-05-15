package clworkos

import (
	"net/http"

	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/zap"
)

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
