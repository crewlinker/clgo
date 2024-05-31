package clworkos

import (
	"context"

	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// Hooks can be optionally provided to the engine to allow custom
// behavior at various points in the authentication flow.
type Hooks interface {
	AuthenticateWithCodeDidSucceed(
		ctx context.Context, idn Identity, usr usermanagement.User, org organizations.Organization) error
}

var _ Hooks = NoOpHooks{}

// NoOpHooks is a no-op implementation of Hooks that is the default value if not is provided. This is exported so
// implementations can embed it and only override the methods they care about.
type NoOpHooks struct{}

func (NoOpHooks) AuthenticateWithCodeDidSucceed(
	ctx context.Context, idn Identity, usr usermanagement.User, org organizations.Organization,
) error {
	return nil
}
