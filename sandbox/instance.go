package sandbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/internal/polling"
	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const defaultWaitTimeout = 2 * time.Minute

// Instance is a stateful handle to one sandbox.
type Instance struct {
	transport *transport.Client

	mu   sync.RWMutex
	data structs.SandboxResponse

	files     *FilesService
	processes *ProcessesService
	computer  *ComputerService
}

func newInstance(client *transport.Client, data structs.SandboxResponse) *Instance {
	instance := &Instance{transport: client, data: data}
	instance.files = &FilesService{instance: instance}
	instance.processes = &ProcessesService{instance: instance}
	instance.computer = newComputerService(instance)
	return instance
}

// ID returns the sandbox identifier.
func (i *Instance) ID() string { return i.snapshot().ID }

// Name returns the user-visible sandbox name.
func (i *Instance) Name() string {
	name := i.snapshot().Name
	if name == nil {
		return ""
	}
	return *name
}

// Status returns the last observed lifecycle status.
func (i *Instance) Status() structs.SandboxStatus { return i.snapshot().Status }

// IPAddress returns the last observed private IP address.
func (i *Instance) IPAddress() string {
	address := i.snapshot().IPAddress
	if address == nil {
		return ""
	}
	return *address
}

// Data returns a copy of the last observed server projection.
func (i *Instance) Data() structs.SandboxResponse { return i.snapshot() }

// Files returns file-transfer operations for this sandbox.
func (i *Instance) Files() *FilesService { return i.files }

// Processes returns managed process and PTY operations for this sandbox.
func (i *Instance) Processes() *ProcessesService { return i.processes }

// Computer returns desktop computer-use operations for this sandbox.
func (i *Instance) Computer() *ComputerService { return i.computer }

func (i *Instance) snapshot() structs.SandboxResponse {
	i.mu.RLock()
	defer i.mu.RUnlock()
	data := i.data
	data.EgressRules = slices.Clone(i.data.EgressRules)
	data.EnvironmentVariables = slices.Clone(i.data.EnvironmentVariables)
	data.SSHPublicKeys = slices.Clone(i.data.SSHPublicKeys)
	return data
}

func (i *Instance) update(data structs.SandboxResponse) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.data = data
}

func (i *Instance) path(suffix string) string {
	return "/v1/sandboxes/" + transport.EncodePath(i.ID()) + suffix
}

// Refresh reloads the sandbox projection.
func (i *Instance) Refresh(ctx context.Context) error {
	var response structs.SandboxResponse
	if err := i.transport.Do(ctx, http.MethodGet, i.path(""), transport.RequestOptions{}, &response); err != nil {
		return fmt.Errorf("refresh sandbox %q: %w", i.ID(), err)
	}
	i.update(response)
	return nil
}

// Pause snapshots and pauses the sandbox.
func (i *Instance) Pause(ctx context.Context) error {
	return i.applyLifecycle(ctx, http.MethodPost, "/pause", "pause")
}

// Resume restores a paused sandbox.
func (i *Instance) Resume(ctx context.Context) error {
	return i.applyLifecycle(ctx, http.MethodPost, "/resume", "resume")
}

func (i *Instance) applyLifecycle(ctx context.Context, method, suffix, operation string) error {
	var response structs.SandboxResponse
	if err := i.transport.Do(ctx, method, i.path(suffix), transport.RequestOptions{}, &response); err != nil {
		return fmt.Errorf("%s sandbox %q: %w", operation, i.ID(), err)
	}
	i.update(response)
	return nil
}

// Fork creates an independent sandbox from this sandbox.
func (i *Instance) Fork(ctx context.Context, request structs.ForkSandboxRequest) (*Instance, error) {
	var response structs.SandboxResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/fork"), transport.RequestOptions{Body: request}, &response)
	if err != nil {
		return nil, fmt.Errorf("fork sandbox %q: %w", i.ID(), err)
	}
	return newInstance(i.transport, response), nil
}

// Destroy starts destruction of the sandbox.
func (i *Instance) Destroy(ctx context.Context) error {
	var response structs.DestroyedResponse
	if err := i.transport.Do(ctx, http.MethodDelete, i.path(""), transport.RequestOptions{}, &response); err != nil {
		return fmt.Errorf("destroy sandbox %q: %w", i.ID(), err)
	}
	data := i.snapshot()
	data.Status = response.Status
	i.update(data)
	return nil
}

// Resize grows the sandbox overlay disk to diskMiB.
func (i *Instance) Resize(ctx context.Context, diskMiB int64) (structs.ResizeSandboxResponse, error) {
	var response structs.ResizeSandboxResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/resize"), transport.RequestOptions{
		Body: structs.ResizeSandboxRequest{DiskMiB: diskMiB},
	}, &response)
	if err != nil {
		return structs.ResizeSandboxResponse{}, fmt.Errorf("resize sandbox %q: %w", i.ID(), err)
	}
	data := i.snapshot()
	data.DiskMiB = response.DiskMiB
	i.update(data)
	return response, nil
}

// SetIngress enables or disables public ingress.
func (i *Instance) SetIngress(ctx context.Context, enabled bool) error {
	var response structs.SandboxResponse
	err := i.transport.Do(ctx, http.MethodPatch, i.path(""), transport.RequestOptions{
		Body: structs.PatchSandboxRequest{IngressEnabled: &enabled},
	}, &response)
	if err != nil {
		return fmt.Errorf("set ingress for sandbox %q: %w", i.ID(), err)
	}
	i.update(response)
	return nil
}

// SetAutoPause sets the idle timeout. A nil duration disables auto-pause.
func (i *Instance) SetAutoPause(ctx context.Context, timeout *time.Duration) error {
	request := structs.PatchSandboxRequest{DisableAutoPause: timeout == nil}
	if timeout != nil {
		if *timeout < time.Minute || *timeout > 24*time.Hour || *timeout%time.Second != 0 {
			return errors.New("auto-pause timeout must be a whole number of seconds between 1 minute and 24 hours")
		}
		seconds := int(timeout.Seconds())
		request.AutoPauseAfterSeconds = &seconds
	}
	var response structs.SandboxResponse
	if err := i.transport.Do(ctx, http.MethodPatch, i.path(""), transport.RequestOptions{Body: request}, &response); err != nil {
		return fmt.Errorf("set auto-pause for sandbox %q: %w", i.ID(), err)
	}
	i.update(response)
	return nil
}

// AddSSHPublicKeys adds OpenSSH public keys to a running sandbox.
func (i *Instance) AddSSHPublicKeys(ctx context.Context, keys []string) (int, error) {
	var response structs.AddSSHPublicKeysResponse
	err := i.transport.Do(ctx, http.MethodPost, i.path("/ssh-pubkeys"), transport.RequestOptions{
		Body: structs.AddSSHPublicKeysRequest{Keys: keys},
	}, &response)
	if err != nil {
		return 0, fmt.Errorf("add SSH public keys to sandbox %q: %w", i.ID(), err)
	}
	return response.Count, nil
}

// WaitUntilRunning waits until the sandbox is running.
func (i *Instance) WaitUntilRunning(ctx context.Context, options structs.WaitOptions) error {
	return i.waitFor(ctx, options, structs.SandboxStatusRunning, map[structs.SandboxStatus]bool{
		structs.SandboxStatusError: true, structs.SandboxStatusFailed: true,
		structs.SandboxStatusDestroying: true, structs.SandboxStatusDestroyed: true,
	})
}

// WaitUntilPaused waits until the sandbox is paused.
func (i *Instance) WaitUntilPaused(ctx context.Context, options structs.WaitOptions) error {
	return i.waitFor(ctx, options, structs.SandboxStatusPaused, map[structs.SandboxStatus]bool{
		structs.SandboxStatusError: true, structs.SandboxStatusFailed: true,
		structs.SandboxStatusDestroying: true, structs.SandboxStatusDestroyed: true,
	})
}

// WaitUntilDestroyed waits until the sandbox is fully destroyed.
func (i *Instance) WaitUntilDestroyed(ctx context.Context, options structs.WaitOptions) error {
	return i.waitFor(ctx, options, structs.SandboxStatusDestroyed, map[structs.SandboxStatus]bool{
		structs.SandboxStatusError: true, structs.SandboxStatusFailed: true,
	})
}

func (i *Instance) waitFor(ctx context.Context, options structs.WaitOptions, desired structs.SandboxStatus, terminal map[structs.SandboxStatus]bool) error {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = defaultWaitTimeout
	}
	err := polling.Wait(ctx, polling.Options{Timeout: timeout}, func(ctx context.Context) (bool, error) {
		requestOptions := transportOptions(options.Request)
		var response structs.SandboxResponse
		if err := i.transport.Do(ctx, http.MethodGet, i.path(""), requestOptions, &response); err != nil {
			return false, err
		}
		i.update(response)
		if terminal[response.Status] {
			return false, fmt.Errorf("sandbox %q entered terminal state %q", response.ID, response.Status)
		}
		return response.Status == desired, nil
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, polling.ErrTimeout) {
		return fmt.Errorf("wait for sandbox %q to become %q: %w", i.ID(), desired, ErrTimeout)
	}
	return fmt.Errorf("wait for sandbox %q to become %q: %w", i.ID(), desired, err)
}
