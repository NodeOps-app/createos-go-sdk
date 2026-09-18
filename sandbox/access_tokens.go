package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// WithAccessToken returns a new handle to this sandbox that authenticates
// runtime operations with its delegated token. The original handle keeps its
// owner credential. Management methods on the delegated handle are rejected
// by the server; use the original handle for token management.
func (i *Instance) WithAccessToken(token string) (*Instance, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("sandbox access token must not be empty")
	}
	return newInstance(i.transport.WithAPIKey(token), i.snapshot()), nil
}

// CreateAccessToken creates a delegated token for this sandbox. The token is
// returned only once. Use an owner's normal API key to call this method.
// If a token is already enabled, the server returns HTTP 409; use
// RotateAccessToken to replace it.
func (i *Instance) CreateAccessToken(ctx context.Context) (structs.SandboxAccessTokenCreateResponse, error) {
	var response structs.SandboxAccessTokenCreateResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/access-token"), transport.RequestOptions{}, &response)
	if err != nil {
		return response, fmt.Errorf("create access token for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// GetAccessToken returns the delegated token's state and redacted hint.
// It never returns the plaintext token. Use an owner's normal API key.
func (i *Instance) GetAccessToken(ctx context.Context) (structs.SandboxAccessTokenMetadata, error) {
	var response structs.SandboxAccessTokenMetadata
	err := i.transport.Do(ctx, http.MethodGet, i.path("/access-token"), transport.RequestOptions{}, &response)
	if err != nil {
		return response, fmt.Errorf("get access token for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// RotateAccessToken replaces the current delegated token and returns its new
// plaintext value once. The server returns HTTP 404 if no token exists.
// Use an owner's normal API key.
func (i *Instance) RotateAccessToken(ctx context.Context) (structs.SandboxAccessTokenCreateResponse, error) {
	var response structs.SandboxAccessTokenCreateResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/access-token/rotate"), transport.RequestOptions{}, &response)
	if err != nil {
		return response, fmt.Errorf("rotate access token for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// DisableAccessToken revokes the current delegated token. It succeeds even if
// no token exists. Revocation is immediate in the home region and propagates
// asynchronously to peer regions. Use an owner's normal API key.
func (i *Instance) DisableAccessToken(ctx context.Context) (structs.SandboxAccessTokenMetadata, error) {
	var response structs.SandboxAccessTokenMetadata
	err := i.transport.Do(ctx, http.MethodDelete, i.path("/access-token"), transport.RequestOptions{}, &response)
	if err != nil {
		return response, fmt.Errorf("disable access token for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}
