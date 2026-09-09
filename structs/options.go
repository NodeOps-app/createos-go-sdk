package structs

import "time"

// RetryOptions configures exponential-backoff retry behavior.
type RetryOptions struct {
	// MaxRetries is the number of attempts after the initial request. Zero
	// disables retries for this request.
	MaxRetries int `json:"-"`
	// BaseDelay is the initial exponential-backoff delay.
	BaseDelay time.Duration `json:"-"`
	// MaxDelay caps an individual exponential-backoff delay.
	MaxDelay time.Duration `json:"-"`
}

// RequestOptions contains per-request transport overrides.
type RequestOptions struct {
	// Headers are added to this request. They override headers with the same
	// name supplied by the SDK, except for authentication.
	Headers map[string]string `json:"-"`
	// Timeout overrides the client's default timeout when greater than zero.
	Timeout time.Duration `json:"-"`
	// Retry overrides the client's retry policy when non-nil.
	Retry *RetryOptions `json:"-"`
	// DisableRetry disables retries and takes precedence over Retry.
	DisableRetry bool `json:"-"`
}

// ListSandboxesOptions filters and limits sandbox listing.
type ListSandboxesOptions struct {
	RequestOptions
	// Limit is the maximum number of sandboxes returned. Zero returns all
	// available sandboxes by walking server-side pages.
	Limit int `json:"-"`
	// Status restricts results to one lifecycle state. The zero value does not
	// filter by status.
	Status SandboxStatus `json:"-"`
}

// ExecOptions configures one buffered or streaming command.
type ExecOptions struct {
	RequestOptions
	// StandardInput replaces RunCommandRequest.StandardInput when non-empty.
	StandardInput string `json:"-"`
	// EnvironmentVariables replaces RunCommandRequest.EnvironmentVariables
	// when non-nil.
	EnvironmentVariables map[string]string `json:"-"`
}

// ManagedProcessConnectOptions configures output replay.
type ManagedProcessConnectOptions struct {
	RequestOptions
	// AfterSequence replays output strictly after this sequence number. Zero
	// starts from the beginning of retained output.
	AfterSequence int64 `json:"-"`
}

// ManagedProcessWaitOptions configures process waiting.
type ManagedProcessWaitOptions struct {
	RequestOptions
	// Scope selects whether to wait for the leader or the complete process tree.
	Scope ManagedProcessWaitScope `json:"-"`
	// WaitTimeout is the server-side long-poll timeout. It is distinct from the
	// HTTP request timeout in RequestOptions.
	WaitTimeout time.Duration `json:"-"`
}

// ManagedProcessDeleteOptions configures graceful process deletion.
type ManagedProcessDeleteOptions struct {
	RequestOptions
	// GracePeriod is the time allowed for graceful termination before the
	// process tree is forcibly stopped.
	GracePeriod time.Duration `json:"-"`
}

// ComputerScreenOptions scopes a desktop operation to one screen.
type ComputerScreenOptions struct {
	RequestOptions
	// ScreenID selects a screen. The zero value lets the API select its default.
	ScreenID ComputerScreenID `json:"-"`
}

// ComputerScreenshotOptions configures screenshot capture and cropping.
type ComputerScreenshotOptions struct {
	ComputerScreenOptions
	// WindowID restricts capture to a window when non-empty.
	WindowID string `json:"-"`
	// X and Y optionally set the crop origin. Pointers distinguish an explicit
	// zero coordinate from an omitted coordinate.
	X *int `json:"-"`
	Y *int `json:"-"`
	// Width and Height set the crop size when greater than zero.
	Width  int `json:"-"`
	Height int `json:"-"`
}

// ComputerListWindowsOptions filters visible desktop windows.
type ComputerListWindowsOptions struct {
	ComputerScreenOptions
	// Application restricts results to windows belonging to this application.
	Application string `json:"-"`
}

// GetTemplateOptions configures a template detail request.
type GetTemplateOptions struct {
	RequestOptions
	// Include requests an optional field in the response.
	Include TemplateInclude `json:"-"`
}

// TemplateLogsOptions filters template build logs.
type TemplateLogsOptions struct {
	RequestOptions
	// Attempt selects a build attempt. Zero selects the current attempt.
	Attempt int `json:"-"`
}

// AttachDiskOptions identifies a disk and its guest mount point.
type AttachDiskOptions struct {
	// DiskID identifies the registered disk to attach.
	DiskID string `json:"-"`
	// MountPath is the absolute destination inside the sandbox.
	MountPath string `json:"-"`
	// SubPath optionally mounts only a path within the disk.
	SubPath string `json:"-"`
}

// DetachDiskOptions identifies one attached disk mount.
type DetachDiskOptions struct {
	// DiskID identifies the registered disk to detach.
	DiskID string `json:"-"`
	// MountPath identifies the mount when the same disk has multiple mounts.
	MountPath string `json:"-"`
}

// CreateSandboxOptions configures the create request transport.
type CreateSandboxOptions struct {
	RequestOptions
}

// WaitOptions configures sandbox status polling.
type WaitOptions struct {
	// Timeout limits status polling. Zero uses the operation's default timeout.
	Timeout time.Duration `json:"-"`
	// Request configures each status request made while polling.
	Request RequestOptions `json:"-"`
}

// PaginationOptions configures a paginated collection request.
type PaginationOptions struct {
	// Limit is the maximum number of items returned. Zero returns all available
	// items by walking server-side pages.
	Limit int `json:"-"`
	// Offset is the zero-based number of items to skip.
	Offset int `json:"-"`
}
