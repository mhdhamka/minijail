package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// Engine is the central Golang Container Runtime
type Engine struct {
	mu         sync.RWMutex
	containers map[string]*Container
	cgroupCtrl *CgroupController
	stopChan   chan struct{}
}

// NewEngine creates and seeds a new container runtime engine
func NewEngine() *Engine {
	engine := &Engine{
		containers: make(map[string]*Container),
		cgroupCtrl: NewCgroupController(),
		stopChan:   make(chan struct{}),
	}

	// Seed 3 realistic initial containers for immediate demonstration
	engine.seedInitialContainers()

	// Start background metric and log tick worker
	go engine.backgroundWorker()

	return engine
}

// GenerateID produces a 64-char container hex ID
func GenerateID() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RandomHostPID returns a realistic host PID (e.g. 10000 - 65535)
func RandomHostPID() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(50000))
	return int(n.Int64()) + 12000
}

func (e *Engine) seedInitialContainers() {
	// Container 1: alpine-web
	c1ID := "c7a8f912e34b1509d845e67041ab34fe20194851239857ab90e812d46e10f391"
	c1StartTime := time.Now().Add(-25 * time.Minute)
	c1HostPID := RandomHostPID()
	c1 := &Container{
		ID:           c1ID,
		Name:         "alpine-web",
		Image:        "alpine:3.19",
		Command:      "sh -c 'while true; do echo \"[HTTP 200] GET /api/v1/health - 12ms\"; sleep 4; done'",
		Status:       StatusRunning,
		CreatedAt:    c1StartTime.Add(-2 * time.Minute),
		StartedAt:    &c1StartTime,
		HostPID:      c1HostPID,
		ContainerPID: 1,
		PortBindings: []string{"8080:80/tcp"},
		Env:          []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "ENV=production"},
		Namespaces:   GenerateNamespaceInodes(c1ID),
		Cgroups:      e.cgroupCtrl.InitLimits(c1ID, 64, 0.5),
		Rootfs:       InitRootfs(c1ID, "alpine:3.19"),
		Processes: []ProcessInfo{
			{HostPID: c1HostPID, ContainerPID: 1, User: "root", CPU: 2.1, MemoryMB: 18.4, Command: "sh -c while true; do echo [HTTP 200]..."},
			{HostPID: c1HostPID + 1, ContainerPID: 8, User: "root", CPU: 0.1, MemoryMB: 4.2, Command: "sleep 4"},
		},
		Logs: []LogEntry{
			{Timestamp: c1StartTime.Format("15:04:05"), Stream: "system", Message: "Mounted OverlayFS lowerdir=/var/lib/minijail/images/alpine:3.19/rootfs,upperdir=/var/lib/minijail/containers/c7a8f912/diff"},
			{Timestamp: c1StartTime.Add(time.Second).Format("15:04:05"), Stream: "system", Message: "pivot_root into /var/lib/minijail/containers/c7a8f912/merged; old root unmounted with MNT_DETACH"},
			{Timestamp: c1StartTime.Add(2 * time.Second).Format("15:04:05"), Stream: "system", Message: "Linux cgroups v2 applied: memory.max=64MB, cpu.max=50000/100000 (0.50 cores)"},
			{Timestamp: c1StartTime.Add(3 * time.Second).Format("15:04:05"), Stream: "stdout", Message: "Server listening on 0.0.0.0:80 (veth_c7a8 -> eth0, IP: 172.18.0.2)"},
			{Timestamp: time.Now().Add(-10 * time.Second).Format("15:04:05"), Stream: "stdout", Message: "[HTTP 200] GET /api/v1/health - 12ms"},
		},
	}
	c1.Namespaces.Hostname = "alpine-web"

	// Container 2: worker-daemon
	c2ID := "9f14b2d8e0917234aa6b71f021c38941da782910e54361ab8920194bcde5912a"
	c2StartTime := time.Now().Add(-12 * time.Minute)
	c2HostPID := RandomHostPID()
	c2 := &Container{
		ID:           c2ID,
		Name:         "worker-daemon",
		Image:        "busybox:1.36",
		Command:      "sh -c 'while true; do echo \"[Job-Queue] Processed payload chunk\"; sleep 8; done'",
		Status:       StatusRunning,
		CreatedAt:    c2StartTime.Add(-1 * time.Minute),
		StartedAt:    &c2StartTime,
		HostPID:      c2HostPID,
		ContainerPID: 1,
		PortBindings: []string{},
		Env:          []string{"PATH=/bin:/sbin", "WORKER_THREADS=4"},
		Namespaces:   GenerateNamespaceInodes(c2ID),
		Cgroups:      e.cgroupCtrl.InitLimits(c2ID, 128, 1.0),
		Rootfs:       InitRootfs(c2ID, "busybox:1.36"),
		Processes: []ProcessInfo{
			{HostPID: c2HostPID, ContainerPID: 1, User: "root", CPU: 1.4, MemoryMB: 12.1, Command: "sh -c while true..."},
		},
		Logs: []LogEntry{
			{Timestamp: c2StartTime.Format("15:04:05"), Stream: "system", Message: "Namespaces cloned: CLONE_NEWPID | CLONE_NEWUTS | CLONE_NEWNS | CLONE_NEWNET"},
			{Timestamp: c2StartTime.Add(time.Second).Format("15:04:05"), Stream: "stdout", Message: "Worker initialized. Subscribed to internal event bus."},
			{Timestamp: time.Now().Add(-20 * time.Second).Format("15:04:05"), Stream: "stdout", Message: "[Job-Queue] Processed payload chunk batch #148"},
		},
	}
	c2.Namespaces.Hostname = "worker-daemon"

	// Container 3: redis-cache (stopped)
	c3ID := "e34b1509d845e67041ab34fe20194851239857ab90e812d46e10f391c7a8f912"
	c3CreatedAt := time.Now().Add(-45 * time.Minute)
	c3 := &Container{
		ID:           c3ID,
		Name:         "redis-cache",
		Image:        "alpine:3.19",
		Command:      "redis-server --protected-mode no",
		Status:       StatusStopped,
		CreatedAt:    c3CreatedAt,
		HostPID:      0,
		ContainerPID: 0,
		PortBindings: []string{"6379:6379/tcp"},
		Env:          []string{"ALLOW_EMPTY_PASSWORD=yes"},
		Namespaces:   GenerateNamespaceInodes(c3ID),
		Cgroups:      e.cgroupCtrl.InitLimits(c3ID, 256, 1.5),
		Rootfs:       InitRootfs(c3ID, "alpine:3.19"),
		Processes:    []ProcessInfo{},
		Logs: []LogEntry{
			{Timestamp: c3CreatedAt.Format("15:04:05"), Stream: "system", Message: "Container created with PID/UTS/MNT namespaces"},
			{Timestamp: c3CreatedAt.Add(2 * time.Minute).Format("15:04:05"), Stream: "system", Message: "Received SIGTERM (stop requested). Graceful shutdown complete."},
		},
	}
	c3.Namespaces.Hostname = "redis-cache"

	e.containers[c1.ID] = c1
	e.containers[c2.ID] = c2
	e.containers[c3.ID] = c3
}

// List returns all registered containers
func (e *Engine) List() []*Container {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Container, 0, len(e.containers))
	for _, c := range e.containers {
		result = append(result, c)
	}
	return result
}

// Get finds a container by exact ID or short ID prefix or Name
func (e *Engine) Get(idOrName string) (*Container, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for id, c := range e.containers {
		if id == idOrName || strings.HasPrefix(id, idOrName) || c.Name == idOrName {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no such container: %s", idOrName)
}

// Create spawns a new container record
func (e *Engine) Create(req CreateContainerRequest) (*Container, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if req.Image == "" {
		req.Image = "alpine:3.19"
	}
	if req.Command == "" {
		req.Command = "sh"
	}
	if req.MemoryLimitMB <= 0 {
		req.MemoryLimitMB = 128
	}
	if req.CpuQuotaCores <= 0 {
		req.CpuQuotaCores = 1.0
	}

	id := GenerateID()
	name := req.Name
	if name == "" {
		name = fmt.Sprintf("container-%s", id[:8])
	}

	// Check name uniqueness
	for _, c := range e.containers {
		if c.Name == name {
			return nil, fmt.Errorf("conflict: container name '%s' is already in use", name)
		}
	}

	ns := GenerateNamespaceInodes(id)
	if req.Hostname != "" {
		ns.Hostname = req.Hostname
	} else {
		ns.Hostname = name
	}

	now := time.Now()
	cont := &Container{
		ID:           id,
		Name:         name,
		Image:        req.Image,
		Command:      req.Command,
		Status:       StatusCreated,
		CreatedAt:    now,
		PortBindings: req.PortBindings,
		Env:          req.Env,
		Namespaces:   ns,
		Cgroups:      e.cgroupCtrl.InitLimits(id, req.MemoryLimitMB, req.CpuQuotaCores),
		Rootfs:       InitRootfs(id, req.Image),
		Processes:    []ProcessInfo{},
		Logs: []LogEntry{
			{Timestamp: now.Format("15:04:05"), Stream: "system", Message: fmt.Sprintf("Created container %s with image %s", id[:12], req.Image)},
			{Timestamp: now.Format("15:04:05"), Stream: "system", Message: fmt.Sprintf("Cgroups v2 assigned: memory.max=%dMB, cpu.max=%.2f cores", req.MemoryLimitMB, req.CpuQuotaCores)},
		},
	}

	e.containers[id] = cont
	return cont, nil
}

// Start boots a created or stopped container
func (e *Engine) Start(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status == StatusRunning {
		return errors.New("container is already running")
	}

	now := time.Now()
	hostPID := RandomHostPID()
	c.HostPID = hostPID
	c.ContainerPID = 1
	c.Status = StatusRunning
	c.StartedAt = &now

	// Reset memory usage to base
	c.Cgroups.MemoryUsageBytes = 14 * 1024 * 1024
	c.Cgroups.CpuPercent = 3.5

	c.Processes = []ProcessInfo{
		{HostPID: hostPID, ContainerPID: 1, User: "root", CPU: 3.5, MemoryMB: 14.0, Command: c.Command},
	}

	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "system",
		Message:   fmt.Sprintf("Kernel clone invoked: host PID=%d mapped to container PID=1 in %s", hostPID, c.Namespaces.PIDInode),
	})
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "system",
		Message:   fmt.Sprintf("Veth pair established: %s (IP: %s)", c.Namespaces.VethPair, c.Namespaces.VirtualIP),
	})
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "stdout",
		Message:   fmt.Sprintf("Process started: %s", c.Command),
	})

	return nil
}

// Stop gracefully terminates a container
func (e *Engine) Stop(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status != StatusRunning && c.Status != StatusPaused {
		return errors.New("container is not running")
	}

	now := time.Now()
	c.Status = StatusStopped
	c.HostPID = 0
	c.ContainerPID = 0
	c.Processes = []ProcessInfo{}
	c.Cgroups.CpuPercent = 0
	c.Cgroups.MemoryUsageBytes = 0

	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "system",
		Message:   "Sent SIGTERM to container PID 1. Gracefully stopped.",
	})

	return nil
}

// Pause suspends process execution via cgroups freezer
func (e *Engine) Pause(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status != StatusRunning {
		return errors.New("can only pause running containers")
	}

	c.Status = StatusPaused
	c.Cgroups.CpuPercent = 0
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Stream:    "system",
		Message:   "cgroup.freeze = 1 applied: all threads in container suspended",
	})
	return nil
}

// Unpause resumes container execution
func (e *Engine) Unpause(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status != StatusPaused {
		return errors.New("container is not paused")
	}

	c.Status = StatusRunning
	c.Cgroups.CpuPercent = 2.5
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Stream:    "system",
		Message:   "cgroup.freeze = 0 applied: threads unfrozen",
	})
	return nil
}

// Kill terminates container immediately with SIGKILL (exit 137)
func (e *Engine) Kill(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	c.Status = StatusStopped
	c.HostPID = 0
	c.ContainerPID = 0
	c.Processes = []ProcessInfo{}
	c.Cgroups.CpuPercent = 0
	c.Cgroups.MemoryUsageBytes = 0

	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "system",
		Message:   "Sent SIGKILL to container process tree. Process terminated immediately (exit 137).",
	})
	return nil
}

// Remove deletes container record
func (e *Engine) Remove(idOrName string) error {
	c, err := e.Get(idOrName)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status == StatusRunning {
		return errors.New("cannot remove a running container. Stop or kill it first")
	}

	delete(e.containers, c.ID)
	return nil
}

// Exec simulates executing a shell command within the container's isolated namespaces
func (e *Engine) Exec(idOrName, command string) (*ExecResponse, error) {
	c, err := e.Get(idOrName)
	if err != nil {
		return nil, err
	}

	if c.Status != StatusRunning {
		return nil, errors.New("container is not running")
	}

	trimmed := strings.TrimSpace(command)
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return &ExecResponse{ExitCode: 0, Stdout: ""}, nil
	}

	baseCmd := parts[0]
	var stdout, stderr string
	exitCode := 0

	switch baseCmd {
	case "uname":
		stdout = "Linux " + c.Namespaces.Hostname + " 4.19.0-minijail #1 SMP x86_64 GNU/Linux\n"
	case "hostname":
		stdout = c.Namespaces.Hostname + "\n"
	case "id", "whoami":
		stdout = "uid=0(root) gid=0(root) groups=0(root),1(bin),2(daemon),10(wheel)\n"
	case "uptime":
		stdout = fmt.Sprintf(" 15:42:10 up 25 min,  0 users,  load average: 0.12, 0.08, 0.04\n")
	case "ip", "ifconfig":
		stdout = fmt.Sprintf("1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN\n    inet 127.0.0.1/8 scope host lo\n2: eth0@if14: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 state UP\n    inet %s scope global eth0\n    link/ether 02:42:ac:12:00:02 brd ff:ff:ff:ff:ff:ff\n", c.Namespaces.VirtualIP)
	case "cat":
		if len(parts) > 1 {
			target := parts[1]
			switch target {
			case "/etc/os-release", "/etc/alpine-release":
				if strings.Contains(c.Image, "alpine") {
					stdout = "NAME=\"Alpine Linux\"\nID=alpine\nVERSION_ID=3.19.1\nPRETTY_NAME=\"Alpine Linux v3.19\"\nHOME_URL=\"https://alpinelinux.org/\"\n"
				} else if strings.Contains(c.Image, "ubuntu") {
					stdout = "NAME=\"Ubuntu\"\nVERSION=\"22.04.4 LTS (Jammy Jellyfish)\"\nID=ubuntu\nPRETTY_NAME=\"Ubuntu 22.04.4 LTS\"\n"
				} else {
					stdout = "BusyBox v1.36.1 multi-call binary.\n"
				}
			case "/etc/hostname":
				stdout = c.Namespaces.Hostname + "\n"
			case "/etc/hosts":
				stdout = fmt.Sprintf("127.0.0.1\tlocalhost\n::1\tlocalhost ip6-localhost ip6-loopback\n%s\t%s\n", strings.Split(c.Namespaces.VirtualIP, "/")[0], c.Namespaces.Hostname)
			case "/etc/resolv.conf":
				stdout = "nameserver 127.0.0.11\noptions ndots:0\n"
			case "/proc/1/cmdline":
				stdout = c.Command + "\n"
			case "/sys/fs/cgroup/memory.max":
				stdout = fmt.Sprintf("%d\n", c.Cgroups.MemoryMaxBytes)
			default:
				stdout = fmt.Sprintf("cat: %s: No such file or directory\n", target)
				exitCode = 1
			}
		} else {
			stdout = ""
		}
	case "ps":
		stdout = "PID   USER     TIME  COMMAND\n"
		for _, p := range c.Processes {
			stdout += fmt.Sprintf("%-5d %-8s 0:00 %s\n", p.ContainerPID, p.User, p.Command)
		}
	case "env":
		stdout = strings.Join(c.Env, "\n") + "\n"
	case "touch":
		if len(parts) > 1 {
			filePath := parts[1]
			c.Rootfs.AddFileDiff(filePath, "CREATED", int64(len(filePath)*8))
			stdout = ""
		}
	case "ls":
		stdout = "bin   dev   etc   home  lib   media mnt   opt   proc  root  run   sbin  srv   sys   tmp   usr   var\n"
	case "echo":
		if len(parts) > 1 {
			stdout = strings.Join(parts[1:], " ") + "\n"
		}
	case "df":
		stdout = "Filesystem           1K-blocks      Used Available Use% Mounted on\noverlay               61255492   4194304  53937188   7% /\ntmpfs                    65536         0     65536   0% /dev\nshm                      65536         0     65536   0% /dev/shm\n"
	default:
		stdout = fmt.Sprintf("%s: command executed inside container PID 1 namespace\n", command)
	}

	// Append to container log
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Stream:    "stdout",
		Message:   fmt.Sprintf("$ %s -> %s", command, strings.TrimSpace(stdout)),
	})

	return &ExecResponse{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// StressMemory simulates an aggressive memory allocation to test cgroups OOM enforcement
func (e *Engine) StressMemory(idOrName string, deltaMB int64) (*Container, bool, error) {
	c, err := e.Get(idOrName)
	if err != nil {
		return nil, false, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status != StatusRunning {
		return nil, false, errors.New("container must be running to stress memory")
	}

	oomTriggered := e.cgroupCtrl.AddMemory(&c.Cgroups, deltaMB)

	now := time.Now()
	if oomTriggered {
		c.Status = StatusOomKilled
		c.HostPID = 0
		c.ContainerPID = 0
		c.Processes = []ProcessInfo{}
		c.Cgroups.CpuPercent = 0

		c.Logs = append(c.Logs, LogEntry{
			Timestamp: now.Format("15:04:05"),
			Stream:    "stderr",
			Message:   fmt.Sprintf("[cgroups v2] memory.max (%d MB) exceeded!", c.Cgroups.MemoryMaxBytes/(1024*1024)),
		})
		c.Logs = append(c.Logs, LogEntry{
			Timestamp: now.Format("15:04:05"),
			Stream:    "system",
			Message:   "[kernel OOM-Killer] Memory cgroup out of memory: Killed process 1 (sh) total-vm:131072kB, anon-rss:65536kB, file-rss:0kB",
		})
		c.Logs = append(c.Logs, LogEntry{
			Timestamp: now.Format("15:04:05"),
			Stream:    "system",
			Message:   "Container exited with code 137 (Fatal error signal 9 - SIGKILL due to OOM)",
		})
	} else {
		c.Logs = append(c.Logs, LogEntry{
			Timestamp: now.Format("15:04:05"),
			Stream:    "stdout",
			Message:   fmt.Sprintf("Allocated +%dMB in anonymous RSS. Total usage: %dMB / %dMB", deltaMB, c.Cgroups.MemoryUsageBytes/(1024*1024), c.Cgroups.MemoryMaxBytes/(1024*1024)),
		})
	}

	return c, oomTriggered, nil
}

// StressCPU triggers high CPU load to test CFS quota throttling
func (e *Engine) StressCPU(idOrName string, durationSec int) (*Container, error) {
	c, err := e.Get(idOrName)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if c.Status != StatusRunning {
		return nil, errors.New("container must be running to stress CPU")
	}

	now := time.Now()
	e.cgroupCtrl.UpdateUsage(&c.Cgroups, true, true)

	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "stdout",
		Message:   fmt.Sprintf("Spawned CPU stress worker threads. Pegged at CFS quota (%.2f%%)", (float64(c.Cgroups.CpuQuotaUs)/float64(c.Cgroups.CpuPeriodUs))*100),
	})
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: now.Format("15:04:05"),
		Stream:    "system",
		Message:   fmt.Sprintf("[cgroups cpu.stat] nr_throttled=%d throttled_usec=%d (CFS scheduler actively enforced quota)", c.Cgroups.ThrottlePeriods, c.Cgroups.ThrottledTimeUs),
	})

	return c, nil
}

func (e *Engine) backgroundWorker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.mu.Lock()
			for _, c := range e.containers {
				if c.Status == StatusRunning {
					e.cgroupCtrl.UpdateUsage(&c.Cgroups, true, false)

					// Periodically add an informative log for demo containers
					if len(c.Logs) < 100 && time.Now().Unix()%15 == 0 {
						c.Logs = append(c.Logs, LogEntry{
							Timestamp: time.Now().Format("15:04:05"),
							Stream:    "stdout",
							Message:   fmt.Sprintf("Heartbeat tick: %s alive [CPU: %.1f%%, RAM: %.1fMiB]", c.Name, c.Cgroups.CpuPercent, float64(c.Cgroups.MemoryUsageBytes)/(1024*1024)),
						})
					}
				}
			}
			e.mu.Unlock()
		case <-e.stopChan:
			return
		}
	}
}
