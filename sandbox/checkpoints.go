package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// CheckpointsService manages local checkpoints inside one running sandbox.
type CheckpointsService struct{ instance *Instance }

// Create captures the current file history of an eligible absolute directory.
func (s *CheckpointsService) Create(ctx context.Context, request structs.CreateCheckpointRequest) (*structs.Checkpoint, error) {
	var result structs.Checkpoint
	if err := s.instance.transport.Do(ctx, http.MethodPost, s.instance.path("/checkpoints"), transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("create checkpoint in sandbox %q: %w", s.instance.ID(), err)
	}
	return &result, nil
}

// List returns checkpoints oldest first.
func (s *CheckpointsService) List(ctx context.Context) ([]structs.Checkpoint, error) {
	var result []structs.Checkpoint
	if err := s.instance.transport.Do(ctx, http.MethodGet, s.instance.path("/checkpoints"), transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("list checkpoints in sandbox %q: %w", s.instance.ID(), err)
	}
	return result, nil
}

// Restore switches the checkpointed directory back to its recorded view and path.
func (s *CheckpointsService) Restore(ctx context.Context, checkpointID string) (*structs.CheckpointRestoreResponse, error) {
	var result structs.CheckpointRestoreResponse
	path := s.instance.path("/checkpoints/" + transport.EncodePath(checkpointID) + "/restore")
	if err := s.instance.transport.Do(ctx, http.MethodPost, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("restore checkpoint %q in sandbox %q: %w", checkpointID, s.instance.ID(), err)
	}
	return &result, nil
}

// Delete removes one checkpoint record while retaining layers still in use.
func (s *CheckpointsService) Delete(ctx context.Context, checkpointID string) error {
	path := s.instance.path("/checkpoints/" + transport.EncodePath(checkpointID))
	var result struct{}
	if err := s.instance.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return fmt.Errorf("delete checkpoint %q in sandbox %q: %w", checkpointID, s.instance.ID(), err)
	}
	return nil
}
