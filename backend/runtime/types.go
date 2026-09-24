package runtime

import (
	"time"
)

// ContainerStatus represents the lifecycle state of a container
type ContainerStatus string

const (
	StatusCreated   ContainerStatus = "created"
	StatusRunning   ContainerStatus = "running"
	StatusPaused    ContainerStatus = "paused"
	StatusStopped   ContainerStatus = "stopped"
	StatusOomKilled ContainerStatus = "oom_killed"
)

// LogEntry represents a timestamped stdout/stderr log
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Stream    string `json:"stream"` // stdout | stderr | system
	Message   string `json:"message"`
}

// ProcessInfo represents a process in the container's PID namespace
type ProcessInfo struct {
	HostPID      int     `json:"hostPid"`
	ContainerPID int     `json:"containerPid"`
	User         string  `json:"user"`
	CPU          float64 `json:"cpu"`
	MemoryMB     float64 `json:"memoryMb"`
	Command      string  `json:"command"`
}

// NamespaceInfo tracks Linux namespaces isolation metadata
type NamespaceInfo struct {
	PIDInode   string `json:"pidInode"`   // CLONE_NEWPID
	UTSInode   string `json:"utsInode"`   // CLONE_NEWUTS
	MNTInode   string `json:"mntInode"`   // CLONE_NEWNS
	NETInode   string `json:"netInode"`   // CLONE_NEWNET
	IPCInode   string `json:"ipcInode"`   // CLONE_NEWIPC
	USERInode  string `json:"userInode"`  // CLONE_NEWUSER
	CGROUPInode string `json:"cgroupInode"` // CLONE_NEWCGROUP
	Hostname   string `json:"hostname"`
	VirtualIP  string `json:"virtualIp"`
	VethPair   string `json:"vethPair"`
}

// CgroupLimits tracks resource constraints and statistics
type CgroupLimits struct {
	CgroupPath       string  `json:"cgroupPath"` // /sys/fs/cgroup/minijail/<id>
	MemoryMaxBytes   int64   `json:"memoryMaxBytes"`
	MemoryUsageBytes int64   `json:"memoryUsageBytes"`
	MemoryMaxUsage   int64   `json:"memoryMaxUsage"`
	CpuQuotaUs       int64   `json:"cpuQuotaUs"`  // e.g. 50000
	CpuPeriodUs      int64   `json:"cpuPeriodUs"` // 100000
	CpuShares        int     `json:"cpuShares"`   // 1024
	CpuPercent       float64 `json:"cpuPercent"`
	ThrottlePeriods  int64   `json:"throttlePeriods"`
	ThrottledTimeUs  int64   `json:"throttledTimeUs"`
	OomKillEvents    int     `json:"oomKillEvents"`
	OomScoreAdj      int     `json:"oomScoreAdj"`
}

// RootfsOverlay tracks the copy-on-write overlay filesystem layers
type RootfsOverlay struct {
	BaseImage     string            `json:"baseImage"`
	LowerDir      string            `json:"lowerDir"`  // Read-only layer
	UpperDir      string            `json:"upperDir"`  // Read-write layer
	WorkDir       string            `json:"workDir"`   // Overlay internal work
	MergedDir     string            `json:"mergedDir"` // Pivot_root destination
	ModifiedFiles []OverlayFileDiff `json:"modifiedFiles"`
}

// OverlayFileDiff tracks Copy-On-Write changes in UpperDir
type OverlayFileDiff struct {
	Path      string `json:"path"`
	Action    string `json:"action"` // ADDED, MODIFIED, DELETED (whiteout)
	SizeBytes int64  `json:"sizeBytes"`
	Timestamp string `json:"timestamp"`
}

// Container represents an active or stopped container instance
type Container struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	Command      string          `json:"command"`
	Status       ContainerStatus `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	HostPID      int             `json:"hostPid"`
	ContainerPID int             `json:"containerPid"`
	PortBindings []string        `json:"portBindings"`
	Env          []string        `json:"env"`
	Namespaces   NamespaceInfo   `json:"namespaces"`
	Cgroups      CgroupLimits    `json:"cgroups"`
	Rootfs       RootfsOverlay   `json:"rootfs"`
	Processes    []ProcessInfo   `json:"processes"`
	Logs         []LogEntry      `json:"logs"`
}

// CreateContainerRequest is the spawner payload
type CreateContainerRequest struct {
	Name           string   `json:"name"`
	Image          string   `json:"image"`
	Command        string   `json:"command"`
	MemoryLimitMB  int64    `json:"memoryLimitMb"`
	CpuQuotaCores  float64  `json:"cpuQuotaCores"` // e.g. 0.5 cores
	EnablePidNs    bool     `json:"enablePidNs"`
	EnableUtsNs    bool     `json:"enableUtsNs"`
	EnableNetNs    bool     `json:"enableNetNs"`
	EnableMntNs    bool     `json:"enableMntNs"`
	Hostname       string   `json:"hostname"`
	PortBindings   []string `json:"portBindings"`
	Env            []string `json:"env"`
}

// ExecRequest is for executing commands inside container
type ExecRequest struct {
	Command string `json:"command"`
}

// ExecResponse returns stdout/stderr and exit code
type ExecResponse struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// CliCommandRequest executes raw docker commands
type CliCommandRequest struct {
	Command string `json:"command"`
}

// CliCommandResponse returns output of CLI command
type CliCommandResponse struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exitCode"`
}
