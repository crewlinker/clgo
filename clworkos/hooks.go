package clworkos

import "context"

// Hooks can be optionally provided to the engine to allow custom
// behavior at various points in the authentication flow.
type Hooks interface {
	AuthenticateWithCodeDidSucceed(ctx context.Context, idn Identity) error
}

var _ Hooks = noOpHooks{}

// noOpHooks is a no-op implementation of Hooks that is the default value if not is provided.
type noOpHooks struct{}

func (noOpHooks) AuthenticateWithCodeDidSucceed(ctx context.Context, idn Identity) error { return nil }
