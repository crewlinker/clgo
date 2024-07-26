package clworkos

import (
	"context"
)

// Hooks can be optionally provided to the engine to allow custom
// behavior at various points in the authentication flow.
type Hooks interface {
	AuthenticateWithCodeDidSucceed(
		ctx context.Context, clientID string, accessToken, refreshToken string,
	) (string, Session, error)
}

var _ Hooks = NoOpHooks{}

// NoOpHooks is a no-op implementation of Hooks that is the default value if not is provided. This is exported so
// implementations can embed it and only override the methods they care about.
type NoOpHooks struct{}

func (NoOpHooks) AuthenticateWithCodeDidSucceed(
	ctx context.Context,
	clientID string,
	accessToken, refreshToken string,
) (string, Session, error) {
	return accessToken, Session{RefreshToken: refreshToken}, nil
}
