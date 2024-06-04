package clworkos

import (
	"context"
	"fmt"

	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"go.uber.org/zap"
)

// limit pages to this many items.
const workOSListLimit = 100

// Lister provides WorkOS list functionality.
type Lister struct {
	cfg   Config
	logs  *zap.Logger
	users UserManagement
	orgs  Organizations
}

// NewLister inits the WorkOS lister.
func NewLister(cfg Config, logs *zap.Logger, users UserManagement, orgs Organizations) *Lister {
	return &Lister{
		cfg:   cfg,
		logs:  logs,
		users: users,
		orgs:  orgs,
	}
}

// ListAllUsers returns all users in WorkOS.
func (ls *Lister) ListAllUsers(ctx context.Context) (users []usermanagement.User, err error) {
	for cursor := ""; ; {
		page, err := ls.users.ListUsers(ctx, usermanagement.ListUsersOpts{After: cursor, Limit: workOSListLimit})
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		users = append(users, page.Data...)

		if page.ListMetadata.After == "" {
			break
		}

		cursor = page.ListMetadata.After
	}

	return
}

// ListAllOrganizations returns all users in WorkOS.
func (ls *Lister) ListAllOrganizations(ctx context.Context) (users []organizations.Organization, err error) {
	for cursor := ""; ; {
		page, err := ls.orgs.ListOrganizations(ctx, organizations.ListOrganizationsOpts{
			After: cursor, Limit: workOSListLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list organizations: %w", err)
		}

		users = append(users, page.Data...)

		if page.ListMetadata.After == "" {
			break
		}

		cursor = page.ListMetadata.After
	}

	return
}
