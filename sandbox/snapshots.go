package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// SnapshotsService manages account-level reusable sandbox snapshots.
type SnapshotsService struct{ transport *transport.Client }

// List returns reusable snapshots owned by the caller.
func (s *SnapshotsService) List(ctx context.Context, options structs.PaginationOptions) ([]structs.SnapshotView, error) {
	snapshots, err := fetchAll[structs.SnapshotView](ctx, s.transport, "/v1/snapshots", paginationQuery(options.Offset), options.Limit, "snapshots")
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	return snapshots, nil
}

// Get returns an owned reusable snapshot by ID.
func (s *SnapshotsService) Get(ctx context.Context, snapshotID string) (*structs.SnapshotView, error) {
	var result structs.SnapshotView
	path := "/v1/snapshots/" + transport.EncodePath(snapshotID)
	if err := s.transport.Do(ctx, http.MethodGet, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("get snapshot %q: %w", snapshotID, err)
	}
	return &result, nil
}

// Delete removes an owned ready or failed snapshot from API listings.
func (s *SnapshotsService) Delete(ctx context.Context, snapshotID string) (*structs.SnapshotDeletedResponse, error) {
	var result structs.SnapshotDeletedResponse
	path := "/v1/snapshots/" + transport.EncodePath(snapshotID)
	if err := s.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("delete snapshot %q: %w", snapshotID, err)
	}
	return &result, nil
}

// CreateSnapshot starts a reusable snapshot capture from this running sandbox.
func (i *Instance) CreateSnapshot(ctx context.Context, request structs.CreateSnapshotRequest) (*structs.SnapshotView, error) {
	var result structs.SnapshotView
	if err := i.transport.Do(ctx, http.MethodPost, i.path("/snapshots"), transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("create snapshot from sandbox %q: %w", i.ID(), err)
	}
	return &result, nil
}
