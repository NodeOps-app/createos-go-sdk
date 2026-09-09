package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// NetworksService manages account-level overlay networks.
type NetworksService struct{ transport *transport.Client }

// List returns overlay networks owned by the caller.
func (s *NetworksService) List(ctx context.Context, options structs.PaginationOptions) ([]structs.Network, error) {
	networks, err := fetchAll[structs.Network](ctx, s.transport, "/v1/networks", paginationQuery(options.Offset), options.Limit, "")
	if err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}
	return networks, nil
}

// Create creates an overlay network.
func (s *NetworksService) Create(ctx context.Context, request structs.NetworkCreateRequest) (*structs.Network, error) {
	var result structs.Network
	if err := s.transport.Do(ctx, http.MethodPost, "/v1/networks", transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}
	return &result, nil
}

// Get returns an overlay network by ID.
func (s *NetworksService) Get(ctx context.Context, networkID string) (*structs.Network, error) {
	var result structs.Network
	path := "/v1/networks/" + transport.EncodePath(networkID)
	if err := s.transport.Do(ctx, http.MethodGet, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("get network %q: %w", networkID, err)
	}
	return &result, nil
}

// Delete removes an overlay network.
func (s *NetworksService) Delete(ctx context.Context, networkID string) error {
	var result structs.OKResponse
	path := "/v1/networks/" + transport.EncodePath(networkID)
	if err := s.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return fmt.Errorf("delete network %q: %w", networkID, err)
	}
	return nil
}

// AttachNetwork connects this sandbox to an existing overlay network.
func (i *Instance) AttachNetwork(ctx context.Context, networkID string) error {
	var result structs.OKResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/networks"), transport.RequestOptions{
		Body: structs.NetworkEntry{ID: networkID},
	}, &result)
	if err != nil {
		return fmt.Errorf("attach network %q to sandbox %q: %w", networkID, i.ID(), err)
	}
	return nil
}

// DetachNetwork disconnects this sandbox from an overlay network.
func (i *Instance) DetachNetwork(ctx context.Context, networkID string) error {
	var result structs.OKResponse
	path := i.path("/networks/" + transport.EncodePath(networkID))
	if err := i.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return fmt.Errorf("detach network %q from sandbox %q: %w", networkID, i.ID(), err)
	}
	return nil
}
