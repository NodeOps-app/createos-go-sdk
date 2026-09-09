package structs

// NetworkEntry references an overlay network at sandbox creation time.
type NetworkEntry struct {
	// ID is the overlay network identifier.
	ID string `json:"id"`
}

// CreateSandboxRequest is the body used to create a sandbox.
type CreateSandboxRequest struct {
	// Shape is the required sizing preset identifier.
	Shape string `json:"shape"`
	// RootFS is a built-in root filesystem name or custom template ID.
	RootFS string `json:"rootfs,omitempty"`
	// Name is an optional human-readable sandbox name.
	Name string `json:"name,omitempty"`
	// Networks are overlay networks attached during creation.
	Networks []NetworkEntry `json:"networks,omitempty"`
	// DiskMiB requests an overlay disk size in mebibytes.
	DiskMiB int64 `json:"disk_mib,omitempty"`
	// EgressRules is the outbound network allowlist.
	EgressRules []string `json:"egress,omitempty"`
	// EnvironmentVariables are injected into the sandbox environment.
	EnvironmentVariables map[string]string `json:"envs,omitempty"`
	// SSHPublicKeys are OpenSSH public keys authorized in the sandbox.
	SSHPublicKeys []string `json:"ssh_pubkeys,omitempty"`
	// HostID requests placement on a specific worker host.
	HostID string `json:"host_id,omitempty"`
	// NodeSelector requests placement on a worker matching these labels.
	NodeSelector map[string]string `json:"node_selector,omitempty"`
	// IngressEnabled enables public preview URLs for the sandbox.
	IngressEnabled bool `json:"ingress_enabled,omitempty"`
	// Disks are persistent disks mounted during creation.
	Disks []DiskAttachment `json:"disks,omitempty"`
	// Region requests placement in a specific region.
	Region string `json:"region,omitempty"`
	// AutoPauseAfterSeconds pauses an idle sandbox after this many seconds.
	AutoPauseAfterSeconds int `json:"auto_pause_after_seconds,omitempty"`
}

// ForkSandboxRequest contains optional overrides applied to a fork.
type ForkSandboxRequest struct {
	// StartPaused leaves the fork paused instead of starting it.
	StartPaused bool `json:"start_paused,omitempty"`
	// SSHPublicKeys replaces the inherited authorized keys when provided.
	SSHPublicKeys []string `json:"ssh_pubkeys,omitempty"`
	// EgressRules replaces the inherited outbound allowlist when provided.
	EgressRules []string `json:"egress,omitempty"`
	// IngressEnabled overrides inherited ingress. A pointer permits explicitly
	// disabling ingress with false.
	IngressEnabled *bool `json:"ingress_enabled,omitempty"`
	// EnvironmentVariables replaces inherited environment variables when provided.
	EnvironmentVariables map[string]string `json:"envs,omitempty"`
}

// PatchSandboxRequest contains mutable sandbox settings.
type PatchSandboxRequest struct {
	// IngressEnabled changes ingress when non-nil.
	IngressEnabled *bool `json:"ingress_enabled,omitempty"`
	// AutoPauseAfterSeconds changes the idle timeout when non-nil.
	AutoPauseAfterSeconds *int `json:"auto_pause_after_seconds,omitempty"`
	// DisableAutoPause removes the configured idle timeout.
	DisableAutoPause bool `json:"disable_auto_pause,omitempty"`
}

// AddSSHPublicKeysRequest adds OpenSSH public keys to a running sandbox.
type AddSSHPublicKeysRequest struct {
	Keys []string `json:"keys"`
}

// RunCommandRequest describes a command to execute inside a sandbox.
type RunCommandRequest struct {
	// Command is the executable name or absolute path.
	Command string `json:"cmd"`
	// Arguments are passed directly to Command without shell expansion.
	Arguments []string `json:"args,omitempty"`
	// StandardInput is written to the command before execution.
	StandardInput string `json:"stdin,omitempty"`
	// EnvironmentVariables are added to the command environment.
	EnvironmentVariables map[string]string `json:"env,omitempty"`
	// Stream requests an NDJSON response and is normally set by StreamCommand.
	Stream bool `json:"stream,omitempty"`
}

// PTYSize specifies terminal dimensions.
type PTYSize struct {
	// Rows is the terminal height in character cells.
	Rows int `json:"rows,omitempty"`
	// Cols is the terminal width in character cells.
	Cols int `json:"cols,omitempty"`
}

// ManagedProcessCreateRequest describes a persistent managed process.
type ManagedProcessCreateRequest struct {
	// Command is the executable name or absolute path.
	Command string `json:"cmd,omitempty"`
	// Arguments are passed directly to Command without shell expansion.
	Arguments []string `json:"args,omitempty"`
	// WorkingDirectory is the process's initial directory.
	WorkingDirectory string `json:"cwd,omitempty"`
	// EnvironmentVariables are added to the process environment.
	EnvironmentVariables map[string]string `json:"env,omitempty"`
	// PTY requests terminal-backed execution when non-nil; nil uses pipes.
	PTY *PTYSize `json:"pty,omitempty"`
}

// ManagedProcessInputRequest carries base64-encoded process input.
type ManagedProcessInputRequest struct {
	DataBase64 string `json:"data_base64"`
}

// ManagedProcessSignalRequest sends a signal to a managed process.
type ManagedProcessSignalRequest struct {
	Signal ManagedProcessSignal `json:"signal"`
}

// ComputerClickRequest describes a mouse click.
type ComputerClickRequest struct {
	// Button defaults to the primary mouse button when omitted.
	Button ComputerMouseButton `json:"button,omitempty"`
	// X and Y optionally move the cursor before clicking. Pointers distinguish
	// an explicit zero coordinate from an omitted coordinate.
	X *int `json:"x,omitempty"`
	Y *int `json:"y,omitempty"`
	// Count is the number of clicks; zero uses the API default.
	Count int `json:"count,omitempty"`
}

// ComputerScrollRequest describes a mouse scroll.
type ComputerScrollRequest struct {
	Direction ComputerScrollDirection `json:"direction,omitempty"`
	Amount    int                     `json:"amount,omitempty"`
}

// ComputerDragRequest describes a mouse drag.
type ComputerDragRequest struct {
	From ComputerPoint `json:"from"`
	To   ComputerPoint `json:"to"`
}

// ComputerButtonRequest describes a mouse button action.
type ComputerButtonRequest struct {
	Button ComputerMouseButton `json:"button,omitempty"`
}

// ComputerTypeRequest describes text typed through the keyboard.
type ComputerTypeRequest struct {
	// Text is typed exactly as supplied.
	Text string `json:"text"`
	// DelayInMS is the delay between characters in milliseconds.
	DelayInMS int `json:"delay_in_ms,omitempty"`
}

// ComputerPressRequest describes a key or key-combination press.
type ComputerPressRequest struct {
	Keys []string `json:"keys"`
}

// ComputerOpenRequest opens a URL or desktop target.
type ComputerOpenRequest struct {
	Target string `json:"target"`
}

// ComputerLaunchRequest launches an installed desktop application.
type ComputerLaunchRequest struct {
	Application string `json:"application"`
	URI         string `json:"uri,omitempty"`
}

// ComputerWindowMoveRequest moves a desktop window.
type ComputerWindowMoveRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ComputerWindowResizeRequest resizes a desktop window.
type ComputerWindowResizeRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ComputerCreateScreenRequest creates or resizes a desktop screen.
type ComputerCreateScreenRequest struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

// SetEgressRequest replaces the sandbox egress allowlist.
type SetEgressRequest struct {
	// EgressRules is the replacement allowlist. A pointer distinguishes an
	// explicitly empty allowlist from an omitted value.
	EgressRules *[]string `json:"egress,omitempty"`
}

// RechargeBandwidthRequest adds bytes to a sandbox bandwidth quota.
type RechargeBandwidthRequest struct {
	// AddBytes is the number of quota bytes to add.
	AddBytes int64 `json:"add_bytes"`
}

// ResizeSandboxRequest grows a sandbox overlay disk.
type ResizeSandboxRequest struct {
	// DiskMiB is the requested total overlay disk size in mebibytes.
	DiskMiB int64 `json:"disk_mib"`
}

// TemplateCreateRequest describes a Dockerfile template build.
type TemplateCreateRequest struct {
	// Name is the human-readable template name.
	Name string `json:"name"`
	// Dockerfile contains the Dockerfile source used for the build.
	Dockerfile string `json:"dockerfile"`
	// Base optionally overrides the base root filesystem.
	Base string `json:"base,omitempty"`
}

// DiskConfig contains non-secret S3 disk configuration.
type DiskConfig struct {
	Bucket       string `json:"bucket"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region,omitempty"`
	UsePathStyle bool   `json:"use_path_style,omitempty"`
}

// DiskCredentials contains write-only bucket credentials.
type DiskCredentials struct {
	// AccessKey is the S3-compatible access key ID.
	AccessKey string `json:"access_key"`
	// SecretKey is the S3-compatible secret access key.
	SecretKey string `json:"secret_key"`
}

// DiskCreateRequest registers a persistent disk.
type DiskCreateRequest struct {
	Name        string          `json:"name"`
	Kind        DiskKind        `json:"kind"`
	Config      DiskConfig      `json:"config"`
	Credentials DiskCredentials `json:"credentials"`
}

// RotateDiskCredentialsRequest replaces the write-only credentials of a disk.
type RotateDiskCredentialsRequest struct {
	Credentials DiskCredentials `json:"credentials"`
}

// DiskAttachment mounts a registered disk into a sandbox.
type DiskAttachment struct {
	// DiskID identifies the registered persistent disk.
	DiskID string `json:"disk_id"`
	// MountPath is the absolute destination inside the sandbox.
	MountPath string `json:"mount_path"`
	// SubPath optionally mounts only a path within the disk.
	SubPath string `json:"sub_path,omitempty"`
}

// NetworkCreateRequest creates an overlay network.
type NetworkCreateRequest struct {
	Name string `json:"name"`
}
