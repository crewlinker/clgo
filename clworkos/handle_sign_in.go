package clworkos

import (
	"context"
	"net/http"

	"github.com/crewlinker/clgo/clserve"
)

// handleSignInStart starts the sign-in flow by redirecting to WorkOS. It also set's a special state
// cookie to protect against CSRF attacks and to transport information (such as redirecting after).
func (h *Handler) handleSignInStart() clserve.HandlerFunc[context.Context] {
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) error {
		// // to protect against CSRF attacks, we generate a nonce and store it in a cookie. It is
		// // checked in the callback
		// nonce, err := h.addStateCookie(w, r.URL.Query().Get("redirect_to"))
		// if err != nil {
		// 	return fmt.Errorf("failed to generate nonce: %w", err)
		// }

		// // get the authorization URL from WorkOS
		// loc, err := h.users.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		// 	ClientID:    h.cfg.MainClientID,
		// 	RedirectURI: h.cfg.CallbackBaseURL + "/auth/wos/callback",
		// 	Provider:    "authkit",
		// 	State:       nonce,
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to get authorization URL: %w", err)
		// }

		// // write the response to the client (with the nonce cookie)
		// http.Redirect(w, r, loc.String(), http.StatusFound)

		return nil
	}
}
