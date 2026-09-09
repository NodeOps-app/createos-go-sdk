package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// ListShapes returns the available CPU, memory, and disk sizing presets.
func (c *Client) ListShapes(ctx context.Context) ([]structs.Shape, error) {
	shapes, err := fetchAll[structs.Shape](ctx, c.transport, "/v1/shapes", nil, 0, "shapes", transport.RequestOptions{SkipAuth: true})
	if err != nil {
		return nil, fmt.Errorf("list sandbox shapes: %w", err)
	}
	return shapes, nil
}

// ListRootFileSystems returns the built-in root filesystem catalog.
func (c *Client) ListRootFileSystems(ctx context.Context) (*structs.RootFSData, error) {
	var result structs.RootFSData
	if err := c.transport.Do(ctx, http.MethodGet, "/v1/rootfs", transport.RequestOptions{SkipAuth: true}, &result); err != nil {
		return nil, fmt.Errorf("list root file systems: %w", err)
	}
	return &result, nil
}

// ListHosts returns the public worker-host projection. The control plane
// requires administrator credentials for this endpoint.
func (c *Client) ListHosts(ctx context.Context) ([]structs.HostPublic, error) {
	hosts, err := fetchAll[structs.HostPublic](ctx, c.transport, "/v1/hosts", nil, 0, "")
	if err != nil {
		return nil, fmt.Errorf("list sandbox hosts: %w", err)
	}
	return hosts, nil
}
