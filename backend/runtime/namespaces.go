package runtime

import (
	"fmt"
	"math/rand"
)

// GenerateNamespaceInodes generates simulated or system Linux namespace inodes
func GenerateNamespaceInodes(containerID string) NamespaceInfo {
	baseInode := 4026532000 + int64(rand.Intn(10000))
	shortID := containerID
	if len(shortID) > 6 {
		shortID = shortID[:6]
	}

	ipOctet := rand.Intn(250) + 2
	virtualIP := fmt.Sprintf("172.18.0.%d/16", ipOctet)
	vethPair := fmt.Sprintf("veth_%s <-> eth0", shortID)

	return NamespaceInfo{
		PIDInode:    fmt.Sprintf("pid:[%d]", baseInode+1),
		UTSInode:    fmt.Sprintf("uts:[%d]", baseInode+2),
		MNTInode:    fmt.Sprintf("mnt:[%d]", baseInode+3),
		NETInode:    fmt.Sprintf("net:[%d]", baseInode+4),
		IPCInode:    fmt.Sprintf("ipc:[%d]", baseInode+5),
		USERInode:   fmt.Sprintf("user:[%d]", baseInode+6),
		CGROUPInode: fmt.Sprintf("cgroup:[%d]", baseInode+7),
		Hostname:    fmt.Sprintf("container-%s", shortID),
		VirtualIP:   virtualIP,
		VethPair:    vethPair,
	}
}

// KernelPrimitiveGuide provides deep architectural documentation of Go kernel bindings
type KernelPrimitiveGuide struct {
	Name        string `json:"name"`
	LinuxFlag   string `json:"linuxFlag"`
	SyscallFile string `json:"syscallFile"`
	GoCode      string `json:"goCode"`
	Explanation string `json:"explanation"`
}

// GetKernelPrimitivesData returns educational explanations and Go code snippets
func GetKernelPrimitivesData() []KernelPrimitiveGuide {
	return []KernelPrimitiveGuide{
		{
			Name:        "PID Namespace (Process Isolation)",
			LinuxFlag:   "syscall.CLONE_NEWPID",
			SyscallFile: "/proc/[pid]/ns/pid",
			GoCode: `cmd := exec.Command("/proc/self/exe", "child", command)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
}
// The first spawned process inside becomes PID 1
// Ps inside only sees this child tree!`,
			Explanation: "Isolates the process ID space. The container's primary process becomes PID 1 (init) inside its own virtual tree, preventing it from seeing or signaling any processes running on the host OS or in other containers.",
		},
		{
			Name:        "UTS Namespace (Hostname & Domain)",
			LinuxFlag:   "syscall.CLONE_NEWUTS",
			SyscallFile: "/proc/[pid]/ns/uts",
			GoCode: `cmd.SysProcAttr.Cloneflags |= syscall.CLONE_NEWUTS
// Inside child handler:
if err := syscall.Sethostname([]byte(containerHostname)); err != nil {
    log.Fatalf("failed to set container hostname: %v", err)
}`,
			Explanation: "Isolates system identifiers including hostname and NIS domain. Setting the hostname inside the container modifies only its local UTS namespace without changing the host system's identity.",
		},
		{
			Name:        "Mount Namespace & Pivot Root (Filesystem)",
			LinuxFlag:   "syscall.CLONE_NEWNS",
			SyscallFile: "/proc/[pid]/ns/mnt",
			GoCode: `// 1. Remount root as private to avoid host propagation
syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
// 2. Bind mount target rootfs
syscall.Mount(newRoot, newRoot, "bind", syscall.MS_BIND|syscall.MS_REC, "")
// 3. Pivot root swap
os.MkdirAll(filepath.Join(newRoot, ".oldroot"), 0700)
syscall.PivotRoot(newRoot, filepath.Join(newRoot, ".oldroot"))
os.Chdir("/")
syscall.Unmount("/.oldroot", syscall.MNT_DETACH)
os.Remove("/.oldroot")`,
			Explanation: "Isolates the filesystem mount table. Together with pivot_root, it detaches the host filesystem and pivots the container root to an isolated OverlayFS directory, making escape to parent directories impossible.",
		},
		{
			Name:        "Network Namespace (Veth & IP)",
			LinuxFlag:   "syscall.CLONE_NEWNET",
			SyscallFile: "/proc/[pid]/ns/net",
			GoCode: `// Go runtime utilizes netlink to create virtual ethernet pairs
// Host side: vethX plugged into bridge docker0
// Container side: peer moved into container netns and renamed to eth0
netlink.LinkAdd(&netlink.Veth{
    LinkAttrs: netlink.LinkAttrs{Name: "veth_c1"},
    PeerName:  "eth0",
})
netlink.LinkSetNsPid(peerLink, containerHostPid)`,
			Explanation: "Provides an isolated network stack including network device interfaces, IPv4/IPv6 addresses, IP routing tables, firewall rules (/proc/net and iptables), and sockets.",
		},
		{
			Name:        "Control Groups (cgroups v2 Resource Throttling)",
			LinuxFlag:   "/sys/fs/cgroup/",
			SyscallFile: "/sys/fs/cgroup/container/<id>/memory.max",
			GoCode: `cgroupPath := "/sys/fs/cgroup/minijail/" + containerID
os.MkdirAll(cgroupPath, 0755)
// Memory limit: 64MB
os.WriteFile(cgroupPath+"/memory.max", []byte("67108864"), 0644)
// CPU quota: 50ms every 100ms (0.5 core)
os.WriteFile(cgroupPath+"/cpu.max", []byte("50000 100000"), 0644)
// Add container PID to cgroup tasks
os.WriteFile(cgroupPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)`,
			Explanation: "Linux cgroups v2 restricts the hardware consumption (CPU time quotas via CFS scheduler, Memory limits, and PID forks) of a process hierarchy and triggers the kernel OOM Killer if thresholds are exceeded.",
		},
	}
}
