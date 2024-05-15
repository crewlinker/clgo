package clworkos

import (
	"context"
	"net/url"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// UserManagement interface describes the interface with the WorkOS User Management API.
type UserManagement interface {
	GetAuthorizationURL(
		opts usermanagement.GetAuthorizationURLOpts) (*url.URL, error)
	GetLogoutURL(
		opts usermanagement.GetLogoutURLOpts) (*url.URL, error)
	AuthenticateWithCode(
		ctx context.Context,
		opts usermanagement.AuthenticateWithCodeOpts) (usermanagement.AuthenticateResponse, error)
	AuthenticateWithRefreshToken(
		ctx context.Context,
		opts usermanagement.AuthenticateWithRefreshTokenOpts) (usermanagement.RefreshAuthenticationResponse, error)
}

// NewUserManagement creates a new UserManagement implementation with the provided configuration.
func NewUserManagement(cfg Config) *usermanagement.Client {
	return usermanagement.NewClient(cfg.APIKey)
}

// Engine implements the core business logic for WorkOS-powered authentication.
type Engine struct{}

// NewEngine creates a new Engine with the provided UserManagement implementation.
func NewEngine(um UserManagement) *Engine {
	return &Engine{}
}

// StartSignInFlow starts the sign-in flow.
func (e *Engine) StartSignInFlow(context.Context) error {
	return nil
}
