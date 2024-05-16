package clworkos

import (
	"context"
	"errors"
	"net/http"

	"github.com/crewlinker/clgo/clserve"
	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/zap"
)

// scope context values to this package.
type ctxKey string

// WithIdentity stores the identity in the request context.
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, ctxKey("identity"), identity)
}

// IdentityFromContext retrieves the identity from the request context.
func IdentityFromContext(ctx context.Context) Identity {
	v, _ := ctx.Value(ctxKey("identity")).(Identity)

	return v
}

// Authenticate provides the authentication middleware. It will set an identity in the request context based on
// the access token. If the access token is expired it will try to refresh the token ad-hoc.
func (h Handler) Authenticate() clserve.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idn, err := h.engine.ContinueSession(r.Context(), w, r)
			if err != nil && !errors.Is(err, ErrNoAuthentication) {
				clzap.Log(r.Context(), h.logs).Warn("middleware failed to continue session", zap.Error(err))
			}

			if idn.IsValid {
				clzap.Log(r.Context(), h.logs).Info("authenticated identity",
					zap.String("session_id", idn.SessionID),
					zap.String("role", idn.Role),
					zap.String("impersonator_email", idn.Impersonator.Email),
					zap.Time("expires_at", idn.ExpiresAt),
					zap.String("organization_id", idn.OrganizationID),
					zap.String("user_id", idn.UserID))
			}

			next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), idn)))
		})
	}
}
