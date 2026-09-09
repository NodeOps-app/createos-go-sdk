package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// DisksService manages account-level registered disks.
type DisksService struct{ transport *transport.Client }

// List returns disks registered by the caller.
func (s *DisksService) List(ctx context.Context, options structs.PaginationOptions) ([]structs.DiskView, error) {
	disks, err := fetchAll[structs.DiskView](ctx, s.transport, "/v1/disks", paginationQuery(options.Offset), options.Limit, "disks")
	if err != nil {
		return nil, fmt.Errorf("list disks: %w", err)
	}
	return disks, nil
}

// Create registers an S3-backed disk.
func (s *DisksService) Create(ctx context.Context, request structs.DiskCreateRequest) (*structs.DiskView, error) {
	var result structs.DiskView
	if err := s.transport.Do(ctx, http.MethodPost, "/v1/disks", transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("create disk: %w", err)
	}
	return &result, nil
}

// Get returns a disk by ID or user-scoped name.
func (s *DisksService) Get(ctx context.Context, diskIDOrName string) (*structs.DiskView, error) {
	var result structs.DiskView
	path := "/v1/disks/" + transport.EncodePath(diskIDOrName)
	if err := s.transport.Do(ctx, http.MethodGet, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("get disk %q: %w", diskIDOrName, err)
	}
	return &result, nil
}

// Delete removes a disk registration without modifying bucket contents.
func (s *DisksService) Delete(ctx context.Context, diskIDOrName string) (*structs.DiskDeletedResponse, error) {
	var result structs.DiskDeletedResponse
	path := "/v1/disks/" + transport.EncodePath(diskIDOrName)
	if err := s.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("delete disk %q: %w", diskIDOrName, err)
	}
	return &result, nil
}

// RotateCredentials replaces the stored credentials for a disk.
func (s *DisksService) RotateCredentials(ctx context.Context, diskIDOrName string, credentials structs.DiskCredentials) (*structs.DiskView, error) {
	request := structs.RotateDiskCredentialsRequest{Credentials: credentials}
	var result structs.DiskView
	path := "/v1/disks/" + transport.EncodePath(diskIDOrName)
	if err := s.transport.Do(ctx, http.MethodPatch, path, transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("rotate credentials for disk %q: %w", diskIDOrName, err)
	}
	return &result, nil
}

// ListDisks returns disks attached to this sandbox.
func (i *Instance) ListDisks(ctx context.Context, options structs.PaginationOptions) ([]structs.SandboxDiskView, error) {
	disks, err := fetchAll[structs.SandboxDiskView](ctx, i.transport, i.path("/disks"), paginationQuery(options.Offset), options.Limit, "disks")
	if err != nil {
		return nil, fmt.Errorf("list disks attached to sandbox %q: %w", i.ID(), err)
	}
	return disks, nil
}

// AttachDisk mounts a registered disk into this running sandbox.
func (i *Instance) AttachDisk(ctx context.Context, options structs.AttachDiskOptions) error {
	request := structs.DiskAttachment(options)
	var result structs.IDResponse
	if err := i.transport.Do(ctx, http.MethodPost, i.path("/disks"), transport.RequestOptions{Body: request}, &result); err != nil {
		return fmt.Errorf("attach disk %q to sandbox %q: %w", options.DiskID, i.ID(), err)
	}
	if result.ID != i.ID() {
		return fmt.Errorf("attach disk %q to sandbox %q: server acknowledged sandbox %q", options.DiskID, i.ID(), result.ID)
	}
	return nil
}

// DetachDisk unmounts one disk attachment from this sandbox.
func (i *Instance) DetachDisk(ctx context.Context, options structs.DetachDiskOptions) (*structs.DiskDetachedResponse, error) {
	query := make(url.Values)
	query.Set("mount_path", options.MountPath)
	var result structs.DiskDetachedResponse
	path := i.path("/disks/" + transport.EncodePath(options.DiskID))
	if err := i.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{Query: query}, &result); err != nil {
		return nil, fmt.Errorf("detach disk %q from sandbox %q: %w", options.DiskID, i.ID(), err)
	}
	return &result, nil
}
