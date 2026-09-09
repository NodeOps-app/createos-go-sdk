package structs

// HostStatus is the scheduling state of a worker host.
type HostStatus string

// Supported host scheduling states.
const (
	HostStatusActive   HostStatus = "active"
	HostStatusDraining HostStatus = "draining"
	HostStatusDead     HostStatus = "dead"
)

// SandboxStatus is the lifecycle state of a sandbox.
type SandboxStatus string

// Supported sandbox lifecycle states.
const (
	SandboxStatusCreating   SandboxStatus = "creating"
	SandboxStatusRunning    SandboxStatus = "running"
	SandboxStatusPausing    SandboxStatus = "pausing"
	SandboxStatusPaused     SandboxStatus = "paused"
	SandboxStatusResuming   SandboxStatus = "resuming"
	SandboxStatusForking    SandboxStatus = "forking"
	SandboxStatusError      SandboxStatus = "error"
	SandboxStatusDestroying SandboxStatus = "destroying"
	SandboxStatusDestroyed  SandboxStatus = "destroyed"
	SandboxStatusFailed     SandboxStatus = "failed"
)

// RetryReason identifies why an HTTP request is retried.
type RetryReason string

// Reasons reported when the transport retries a request.
const (
	RetryReasonNetwork   RetryReason = "network"
	RetryReasonStatus    RetryReason = "status"
	RetryReasonRateLimit RetryReason = "rate-limit"
)

// ExecStreamEventType identifies a command stream event.
type ExecStreamEventType string

// Event types emitted by CommandStream.
const (
	ExecStreamEventStdout    ExecStreamEventType = "stdout"
	ExecStreamEventStderr    ExecStreamEventType = "stderr"
	ExecStreamEventExit      ExecStreamEventType = "exit"
	ExecStreamEventError     ExecStreamEventType = "error"
	ExecStreamEventHeartbeat ExecStreamEventType = "heartbeat"
)

// ManagedProcessKind identifies whether a managed process uses pipes or a PTY.
type ManagedProcessKind string

// Supported managed process execution modes.
const (
	ManagedProcessKindProcess ManagedProcessKind = "process"
	ManagedProcessKindPTY     ManagedProcessKind = "pty"
)

// ManagedProcessState is the lifecycle state of a managed process.
type ManagedProcessState string

// Supported managed process lifecycle states.
const (
	ManagedProcessStateStarting    ManagedProcessState = "starting"
	ManagedProcessStateRunning     ManagedProcessState = "running"
	ManagedProcessStateTerminating ManagedProcessState = "terminating"
	ManagedProcessStateExited      ManagedProcessState = "exited"
	ManagedProcessStateFailed      ManagedProcessState = "failed"
)

// ManagedProcessSignal is a signal accepted by the managed process API.
type ManagedProcessSignal string

// Signals accepted by ProcessesService.Signal.
const (
	SignalHangup       ManagedProcessSignal = "SIGHUP"
	SignalInterrupt    ManagedProcessSignal = "SIGINT"
	SignalQuit         ManagedProcessSignal = "SIGQUIT"
	SignalKill         ManagedProcessSignal = "SIGKILL"
	SignalTerminate    ManagedProcessSignal = "SIGTERM"
	SignalUserDefined1 ManagedProcessSignal = "SIGUSR1"
	SignalUserDefined2 ManagedProcessSignal = "SIGUSR2"
	SignalWindowChange ManagedProcessSignal = "SIGWINCH"
)

// ManagedProcessStream identifies an output stream.
type ManagedProcessStream string

// Output streams emitted by a managed process connection.
const (
	ManagedProcessStreamStdout ManagedProcessStream = "stdout"
	ManagedProcessStreamStderr ManagedProcessStream = "stderr"
	ManagedProcessStreamPTY    ManagedProcessStream = "pty"
)

// ManagedProcessConnectEventType identifies a process connection event.
type ManagedProcessConnectEventType string

// Event types emitted by ProcessStream.
const (
	ManagedProcessConnectEventData      ManagedProcessConnectEventType = "data"
	ManagedProcessConnectEventExit      ManagedProcessConnectEventType = "exit"
	ManagedProcessConnectEventHeartbeat ManagedProcessConnectEventType = "heartbeat"
	ManagedProcessConnectEventError     ManagedProcessConnectEventType = "error"
)

// ComputerScreenID identifies one of the eight supported desktop screens.
type ComputerScreenID string

// Screen identifiers accepted by desktop operations.
const (
	ComputerScreen0 ComputerScreenID = "screen-0"
	ComputerScreen1 ComputerScreenID = "screen-1"
	ComputerScreen2 ComputerScreenID = "screen-2"
	ComputerScreen3 ComputerScreenID = "screen-3"
	ComputerScreen4 ComputerScreenID = "screen-4"
	ComputerScreen5 ComputerScreenID = "screen-5"
	ComputerScreen6 ComputerScreenID = "screen-6"
	ComputerScreen7 ComputerScreenID = "screen-7"
)

// ComputerMouseButton identifies a mouse button.
type ComputerMouseButton string

// Mouse buttons accepted by mouse operations.
const (
	ComputerMouseButtonLeft   ComputerMouseButton = "left"
	ComputerMouseButtonMiddle ComputerMouseButton = "middle"
	ComputerMouseButtonRight  ComputerMouseButton = "right"
)

// ComputerScrollDirection identifies a vertical scroll direction.
type ComputerScrollDirection string

// Vertical directions accepted by MouseService.Scroll.
const (
	ComputerScrollUp   ComputerScrollDirection = "up"
	ComputerScrollDown ComputerScrollDirection = "down"
)

// TemplateStatus is the build state of a template.
type TemplateStatus string

// Supported template build states.
const (
	TemplateStatusPending  TemplateStatus = "pending"
	TemplateStatusBuilding TemplateStatus = "building"
	TemplateStatusReady    TemplateStatus = "ready"
	TemplateStatusFailed   TemplateStatus = "failed"
)

// TemplateInclude identifies optional template detail fields.
type TemplateInclude string

// TemplateIncludeDockerfile requests the source Dockerfile in template details.
const TemplateIncludeDockerfile TemplateInclude = "dockerfile"

// DiskKind identifies a registered disk backend.
type DiskKind string

// DiskKindS3 identifies an S3-compatible persistent disk.
const DiskKindS3 DiskKind = "s3"

// DiskMountStatus is the state of a sandbox disk mount.
type DiskMountStatus string

// Supported states of a sandbox disk mount.
const (
	DiskMountStatusPending    DiskMountStatus = "pending"
	DiskMountStatusMounted    DiskMountStatus = "mounted"
	DiskMountStatusError      DiskMountStatus = "error"
	DiskMountStatusUnmounting DiskMountStatus = "unmounting"
)

// ManagedProcessWaitScope selects which process boundary to wait for.
type ManagedProcessWaitScope string

// Process boundaries accepted by ProcessesService.Wait.
const (
	ManagedProcessWaitScopeLeader ManagedProcessWaitScope = "leader"
	ManagedProcessWaitScopeTree   ManagedProcessWaitScope = "tree"
)
