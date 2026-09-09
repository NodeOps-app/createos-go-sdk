package structs

import "time"

// JSendEnvelope is the generic wrapper used by every JSON API response.
// Data is the endpoint-specific success or failure payload.
type JSendEnvelope[T any] struct {
	Status  string `json:"status"`
	Data    T      `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// PaginationMetadata describes one server-side collection page.
type PaginationMetadata struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

// PaginatedResponse contains one page of collection data.
type PaginatedResponse[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// Shape describes a sandbox CPU, memory, and disk sizing preset.
type Shape struct {
	ID             string `json:"id"`
	VCPU           int    `json:"vcpu"`
	MemoryMiB      int    `json:"mem_mib"`
	DefaultDiskMiB int64  `json:"default_disk_mib"`
	CPUQuotaPct    int    `json:"cpu_quota_pct,omitempty"`
}

// RootFSEntry describes a built-in root filesystem image.
type RootFSEntry struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Deprecated  bool    `json:"deprecated,omitempty"`
	Successor   *string `json:"successor,omitempty"`
}

// RootFSData is the built-in root filesystem catalog.
type RootFSData struct {
	RootFileSystems []string      `json:"rootfs"`
	Default         string        `json:"default"`
	Entries         []RootFSEntry `json:"entries,omitempty"`
}

// HostPublic describes a worker host visible to an administrator.
type HostPublic struct {
	ID              string     `json:"id"`
	Status          HostStatus `json:"status"`
	FreeMemoryMiB   int        `json:"free_mib"`
	SandboxCount    int        `json:"vm_count"`
	RootFileSystems []string   `json:"rootfses,omitempty"`
}

// CreateSandboxResponse records resolved placement and boot timing.
type CreateSandboxResponse struct {
	ID                  string        `json:"id"`
	Status              SandboxStatus `json:"status"`
	Name                *string       `json:"name,omitempty"`
	IPAddress           string        `json:"ip"`
	Shape               string        `json:"shape"`
	RootFS              *string       `json:"rootfs,omitempty"`
	VCPU                int           `json:"vcpu"`
	MemoryMiB           int           `json:"mem_mib"`
	DiskMiB             int64         `json:"disk_mib"`
	SpawnMilliseconds   float64       `json:"spawn_ms"`
	EgressRules         []string      `json:"egress"`
	BandwidthQuotaBytes int64         `json:"bandwidth_quota_bytes"`
	IngressURLTemplate  string        `json:"ingress_url_template,omitempty"`
}

// SandboxResponse is the full server-side projection of a sandbox.
type SandboxResponse struct {
	ID                    string        `json:"id"`
	Status                SandboxStatus `json:"status"`
	IPAddress             *string       `json:"ip,omitempty"`
	VCPU                  int           `json:"vcpu"`
	MemoryMiB             int           `json:"mem_mib"`
	DiskMiB               int64         `json:"disk_mib"`
	CreatedAt             time.Time     `json:"created_at"`
	IngressEnabled        bool          `json:"ingress_enabled"`
	IngressURLTemplate    string        `json:"ingress_url_template,omitempty"`
	Name                  *string       `json:"name,omitempty"`
	RunningAt             *time.Time    `json:"running_at,omitempty"`
	DestroyedAt           *time.Time    `json:"destroyed_at,omitempty"`
	SpawnMilliseconds     float64       `json:"spawn_ms,omitempty"`
	Shape                 string        `json:"shape,omitempty"`
	RootFS                *string       `json:"rootfs,omitempty"`
	Region                string        `json:"region,omitempty"`
	EgressRules           []string      `json:"egress,omitempty"`
	EnvironmentVariables  []string      `json:"envs,omitempty"`
	SSHPublicKeys         []string      `json:"ssh_pubkeys,omitempty"`
	CreatedBy             string        `json:"created_by,omitempty"`
	BandwidthIngressBytes int64         `json:"bandwidth_ingress_bytes,omitempty"`
	PausedAt              *time.Time    `json:"paused_at,omitempty"`
	LastResumedAt         *time.Time    `json:"last_resumed_at,omitempty"`
	ForkedFrom            *string       `json:"forked_from,omitempty"`
	AutoPauseAfterSeconds *int          `json:"auto_pause_after_seconds,omitempty"`
}

// AddSSHPublicKeysResponse reports the resulting key count.
type AddSSHPublicKeysResponse struct {
	Count int `json:"count"`
}

// DestroyedResponse reports a sandbox destruction result.
type DestroyedResponse struct {
	ID     string        `json:"id"`
	Status SandboxStatus `json:"status"`
}

// CommandResult contains buffered command output.
type CommandResult struct {
	StandardOutput string `json:"stdout"`
	StandardError  string `json:"stderr"`
	ExitCode       int    `json:"exit_code"`
	ErrorMessage   string `json:"error,omitempty"`
}

// RunCommandResponse contains a command result and timing.
type RunCommandResponse struct {
	Result                CommandResult `json:"result"`
	ExecutionMilliseconds float64       `json:"exec_ms"`
}

// CommandStreamFrame is a raw command NDJSON frame.
type CommandStreamFrame struct {
	StandardOutput string `json:"stdout,omitempty"`
	StandardError  string `json:"stderr,omitempty"`
	ExitCode       *int   `json:"exit_code,omitempty"`
	ErrorMessage   string `json:"error,omitempty"`
	Heartbeat      bool   `json:"hb,omitempty"`
}

// CommandStreamEvent is a decoded command stream event.
type CommandStreamEvent struct {
	Type         ExecStreamEventType `json:"type"`
	Data         string              `json:"data,omitempty"`
	ExitCode     *int                `json:"exitCode,omitempty"`
	ErrorMessage string              `json:"message,omitempty"`
}

// ManagedProcessOutputWindow describes retained output sequence bounds.
type ManagedProcessOutputWindow struct {
	OldestSequence int64 `json:"oldest_seq"`
	NewestSequence int64 `json:"newest_seq"`
	Bytes          int64 `json:"bytes"`
}

// ManagedProcessForeground describes the foreground command in a PTY.
type ManagedProcessForeground struct {
	ProcessID int      `json:"pid,omitempty"`
	Command   string   `json:"cmd"`
	Arguments []string `json:"args,omitempty"`
}

// ManagedProcess is the server projection of a persistent process.
type ManagedProcess struct {
	ProcessID        string                     `json:"process_id"`
	Kind             ManagedProcessKind         `json:"kind"`
	PID              int                        `json:"pid"`
	State            ManagedProcessState        `json:"state"`
	LeaderExited     bool                       `json:"leader_exited"`
	TreeExited       bool                       `json:"tree_exited"`
	CreatedAt        time.Time                  `json:"created_at"`
	FinishedAt       *time.Time                 `json:"finished_at,omitempty"`
	ExitCode         *int                       `json:"exit_code,omitempty"`
	Signal           *string                    `json:"signal,omitempty"`
	Command          string                     `json:"cmd,omitempty"`
	Arguments        []string                   `json:"args,omitempty"`
	WorkingDirectory string                     `json:"cwd,omitempty"`
	Foreground       *ManagedProcessForeground  `json:"foreground,omitempty"`
	Output           ManagedProcessOutputWindow `json:"output"`
}

// ManagedProcessListResponse contains managed processes in a sandbox.
type ManagedProcessListResponse struct {
	Processes []ManagedProcess `json:"processes"`
}

// ManagedProcessConnectFrame is a raw process connection NDJSON frame.
type ManagedProcessConnectFrame struct {
	Type                    ManagedProcessConnectEventType `json:"type"`
	Sequence                int64                          `json:"seq,omitempty"`
	Stream                  ManagedProcessStream           `json:"stream,omitempty"`
	DataBase64              string                         `json:"data_base64,omitempty"`
	ExitCode                *int                           `json:"exit_code,omitempty"`
	Signal                  *string                        `json:"signal,omitempty"`
	ErrorMessage            string                         `json:"error,omitempty"`
	OldestAvailableSequence int64                          `json:"oldest_available_seq,omitempty"`
}

// ManagedProcessConnectEvent is a decoded process connection event.
type ManagedProcessConnectEvent struct {
	Type                    ManagedProcessConnectEventType `json:"type"`
	Sequence                int64                          `json:"seq,omitempty"`
	Stream                  ManagedProcessStream           `json:"stream,omitempty"`
	Data                    []byte                         `json:"data,omitempty"`
	ExitCode                *int                           `json:"exitCode,omitempty"`
	Signal                  *string                        `json:"signal,omitempty"`
	ErrorMessage            string                         `json:"message,omitempty"`
	OldestAvailableSequence int64                          `json:"oldestAvailableSeq,omitempty"`
}

// ManagedProcessInputResponse acknowledges process input.
type ManagedProcessInputResponse struct {
	InputSequence int64 `json:"input_seq"`
}

// ComputerPoint is a coordinate on a desktop screen.
type ComputerPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ComputerScreenGeometry describes desktop screen dimensions.
type ComputerScreenGeometry struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ComputerClipboard contains desktop clipboard text.
type ComputerClipboard struct {
	Text string `json:"text"`
}

// ComputerWindow describes a visible desktop window.
type ComputerWindow struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

// ComputerWindowGeometry describes a window's position and size.
type ComputerWindowGeometry struct {
	ID     string `json:"id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Screen int    `json:"screen"`
}

// ComputerScreen describes a configured desktop screen.
type ComputerScreen struct {
	ScreenID  ComputerScreenID `json:"screen_id"`
	Display   string           `json:"display"`
	Width     int              `json:"width"`
	Height    int              `json:"height"`
	VNCPort   int              `json:"vnc_port"`
	NoVNCPort int              `json:"novnc_port"`
}

// ComputerScreenConnection contains temporary noVNC connection details.
type ComputerScreenConnection struct {
	ScreenID  ComputerScreenID `json:"screen_id"`
	Port      int              `json:"port"`
	Path      string           `json:"path"`
	Token     string           `json:"token"`
	ExpiresAt time.Time        `json:"expires_at"`
	URL       string           `json:"url,omitempty"`
}

// EgressView contains the current sandbox egress allowlist.
type EgressView struct {
	ID          string   `json:"id"`
	EgressRules []string `json:"egress"`
}

// BandwidthView contains sandbox quota and usage counters.
type BandwidthView struct {
	ID             string `json:"id"`
	QuotaBytes     int64  `json:"quota_bytes"`
	UsedBytes      int64  `json:"used_bytes"`
	IngressBytes   int64  `json:"ingress_bytes"`
	RemainingBytes int64  `json:"remaining_bytes"`
	Capped         bool   `json:"capped"`
}

// ResizeSandboxResponse reports the sandbox disk size after a grow.
type ResizeSandboxResponse struct {
	ID      string `json:"id"`
	DiskMiB int64  `json:"disk_mib"`
}

// IDResponse is the canonical acknowledgement for a mutation of one resource.
type IDResponse struct {
	ID string `json:"id"`
}

// WhoAmIStatsView contains sandbox counts for the calling identity.
type WhoAmIStatsView struct {
	Running int `json:"running"`
	Paused  int `json:"paused"`
	Other   int `json:"other"`
	Total   int `json:"total"`
}

// WhoAmIView describes the authenticated identity.
type WhoAmIView struct {
	UserID string          `json:"user_id"`
	Stats  WhoAmIStatsView `json:"stats"`
}

// TemplateView describes a custom root filesystem template.
type TemplateView struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Base          string         `json:"base"`
	Status        TemplateStatus `json:"status"`
	Ext4SizeBytes int64          `json:"ext4_size_bytes"`
	CreatedAt     time.Time      `json:"created_at"`
	BuiltAt       *time.Time     `json:"built_at,omitempty"`
	Dockerfile    string         `json:"dockerfile,omitempty"`
}

// TemplateLogEvent is one template build log frame.
type TemplateLogEvent struct {
	Timestamp *time.Time             `json:"ts,omitempty"`
	Level     string                 `json:"level,omitempty"`
	Line      string                 `json:"line,omitempty"`
	Attempt   int                    `json:"attempt,omitempty"`
	Final     bool                   `json:"final,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

// DiskView describes a registered persistent disk without credentials.
type DiskView struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Kind      DiskKind   `json:"kind"`
	Config    DiskConfig `json:"config"`
	CreatedAt time.Time  `json:"created_at"`
}

// SandboxDiskView describes one disk attached to a sandbox.
type SandboxDiskView struct {
	DiskID      string          `json:"disk_id"`
	Name        string          `json:"name"`
	Kind        DiskKind        `json:"kind"`
	Config      DiskConfig      `json:"config"`
	MountPath   string          `json:"mount_path"`
	SubPath     string          `json:"sub_path,omitempty"`
	MountStatus DiskMountStatus `json:"mount_status"`
	MountError  string          `json:"mount_error,omitempty"`
}

// DiskDeletedResponse reports whether a disk was deleted.
type DiskDeletedResponse struct {
	Deleted bool `json:"deleted"`
}

// DiskDetachedResponse reports whether a disk was detached.
type DiskDetachedResponse struct {
	Detached bool `json:"detached"`
}

// NetworkMember describes a sandbox attached to an overlay network.
type NetworkMember struct {
	SandboxID string `json:"sandbox_id"`
	Status    string `json:"status"`
	IPAddress string `json:"ip,omitempty"`
	Name      string `json:"name,omitempty"`
}

// Network describes an overlay network.
type Network struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	CreatedAt   time.Time       `json:"created_at"`
	MemberCount int             `json:"member_count,omitempty"`
	Members     []NetworkMember `json:"members,omitempty"`
}

// OKResponse is a generic successful acknowledgement.
type OKResponse struct {
	OK bool `json:"ok"`
}

// HealthzResponse is the control-plane liveness result.
type HealthzResponse struct {
	Up bool `json:"up"`
}

// ReadyzResponse is the control-plane readiness result.
type ReadyzResponse struct {
	Ready                          bool   `json:"ready"`
	Reason                         string `json:"reason,omitempty"`
	SchedulerLastOKMillisecondsAgo int64  `json:"scheduler_last_ok_ms_ago,omitempty"`
}
