package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CliInterpreter processes raw docker commands against the runtime engine
type CliInterpreter struct {
	engine *Engine
}

// NewCliInterpreter creates a CLI parser
func NewCliInterpreter(engine *Engine) *CliInterpreter {
	return &CliInterpreter{engine: engine}
}

// Execute parses and runs a docker CLI command string
func (c *CliInterpreter) Execute(cmdLine string) CliCommandResponse {
	trimmed := strings.TrimSpace(cmdLine)
	if trimmed == "" {
		return CliCommandResponse{Output: "", ExitCode: 0}
	}

	parts := strings.Fields(trimmed)
	if parts[0] != "docker" && parts[0] != "minijail" {
		return CliCommandResponse{
			Error:    fmt.Sprintf("command not found: %s. Supported root commands: 'docker', 'minijail'", parts[0]),
			ExitCode: 127,
		}
	}

	if len(parts) == 1 {
		return CliCommandResponse{Output: c.helpText(), ExitCode: 0}
	}

	subCmd := parts[1]
	args := parts[2:]

	switch subCmd {
	case "ps":
		all := false
		for _, arg := range args {
			if arg == "-a" || arg == "--all" {
				all = true
			}
		}
		return CliCommandResponse{Output: c.cmdPs(all), ExitCode: 0}

	case "run":
		return c.cmdRun(args)

	case "stop":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker stop requires at least 1 container ID/name", ExitCode: 1}
		}
		err := c.engine.Stop(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		return CliCommandResponse{Output: args[0], ExitCode: 0}

	case "start":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker start requires at least 1 container ID/name", ExitCode: 1}
		}
		err := c.engine.Start(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		return CliCommandResponse{Output: args[0], ExitCode: 0}

	case "restart":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker restart requires at least 1 container ID/name", ExitCode: 1}
		}
		_ = c.engine.Stop(args[0])
		err := c.engine.Start(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		return CliCommandResponse{Output: args[0], ExitCode: 0}

	case "kill":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker kill requires at least 1 container ID/name", ExitCode: 1}
		}
		err := c.engine.Kill(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		return CliCommandResponse{Output: args[0], ExitCode: 0}

	case "rm":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker rm requires at least 1 container ID/name", ExitCode: 1}
		}
		err := c.engine.Remove(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		return CliCommandResponse{Output: args[0], ExitCode: 0}

	case "logs":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker logs requires at least 1 container ID/name", ExitCode: 1}
		}
		cObj, err := c.engine.Get(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		var lines []string
		for _, log := range cObj.Logs {
			lines = append(lines, fmt.Sprintf("[%s] (%s) %s", log.Timestamp, log.Stream, log.Message))
		}
		return CliCommandResponse{Output: strings.Join(lines, "\n"), ExitCode: 0}

	case "stats":
		return CliCommandResponse{Output: c.cmdStats(), ExitCode: 0}

	case "top":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker top requires at least 1 container ID/name", ExitCode: 1}
		}
		cObj, err := c.engine.Get(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		output := "UID   PID   PPID  C  STIME  TTY   TIME      CMD\n"
		for _, p := range cObj.Processes {
			output += fmt.Sprintf("%-5s %-5d %-5d %-2.0f %-6s %-5s %-9s %s\n",
				p.User, p.HostPID, 1, p.CPU, "00:00", "?", "00:00:01", p.Command)
		}
		return CliCommandResponse{Output: output, ExitCode: 0}

	case "inspect":
		if len(args) == 0 {
			return CliCommandResponse{Error: "docker inspect requires at least 1 container ID/name", ExitCode: 1}
		}
		cObj, err := c.engine.Get(args[0])
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		data, _ := json.MarshalIndent(cObj, "", "  ")
		return CliCommandResponse{Output: string(data), ExitCode: 0}

	case "exec":
		if len(args) < 2 {
			return CliCommandResponse{Error: "usage: docker exec <container> <command>", ExitCode: 1}
		}
		target := args[0]
		// skip flags like -it or -d
		cmdIndex := 1
		for i := 1; i < len(args); i++ {
			if strings.HasPrefix(args[i], "-") {
				cmdIndex = i + 1
			} else {
				break
			}
		}
		if cmdIndex >= len(args) {
			return CliCommandResponse{Error: "no command specified for exec", ExitCode: 1}
		}
		execCmd := strings.Join(args[cmdIndex:], " ")
		res, err := c.engine.Exec(target, execCmd)
		if err != nil {
			return CliCommandResponse{Error: err.Error(), ExitCode: 1}
		}
		out := res.Stdout
		if res.Stderr != "" {
			out += "\n" + res.Stderr
		}
		return CliCommandResponse{Output: out, ExitCode: res.ExitCode}

	case "images":
		output := "REPOSITORY   TAG       IMAGE ID       SIZE\n" +
			"alpine       3.19      9a807d4b9261   7.38MB\n" +
			"busybox      1.36      e97db6746ef7   4.26MB\n" +
			"ubuntu       22.04     5921820980f7   77.8MB\n" +
			"nginx        alpine    f801646274b5   23.4MB\n"
		return CliCommandResponse{Output: output, ExitCode: 0}

	case "version":
		output := "Client: Minijail-Go Engine / Docker CLI\n" +
			" Version:           26.0.0-sim\n" +
			" API version:       1.45 (Go runtime)\n" +
			" Go version:        go1.22.5\n" +
			" Git commit:        9a807d4\n" +
			" OS/Arch:           linux/amd64\n" +
			" Experimental:      true\n" +
			" Kernel Primitives: Linux Namespaces (PID, UTS, MNT, NET, IPC, USER), cgroups v2, OverlayFS\n"
		return CliCommandResponse{Output: output, ExitCode: 0}

	case "help", "--help", "-h":
		return CliCommandResponse{Output: c.helpText(), ExitCode: 0}

	default:
		return CliCommandResponse{
			Error:    fmt.Sprintf("docker: '%s' is not a docker command. See 'docker --help'", subCmd),
			ExitCode: 1,
		}
	}
}

func (c *CliInterpreter) cmdPs(all bool) string {
	containers := c.engine.List()
	output := fmt.Sprintf("%-12s %-16s %-20s %-12s %-16s %-16s\n",
		"CONTAINER ID", "IMAGE", "COMMAND", "STATUS", "PORTS", "NAMES")

	count := 0
	for _, cont := range containers {
		if !all && cont.Status != StatusRunning {
			continue
		}
		shortCmd := cont.Command
		if len(shortCmd) > 18 {
			shortCmd = shortCmd[:15] + "..."
		}
		ports := strings.Join(cont.PortBindings, ", ")
		if ports == "" {
			ports = "-"
		}
		statusStr := string(cont.Status)
		if cont.Status == StatusRunning {
			statusStr = "Up (Active)"
		} else if cont.Status == StatusOomKilled {
			statusStr = "Exited (137) OOM"
		} else if cont.Status == StatusStopped {
			statusStr = "Exited (0)"
		}

		output += fmt.Sprintf("%-12s %-16s %-20s %-12s %-16s %-16s\n",
			cont.ID[:12], cont.Image, fmt.Sprintf("\"%s\"", shortCmd), statusStr, ports, cont.Name)
		count++
	}
	if count == 0 {
		output += "(no active containers. Use 'docker ps -a' to see all)\n"
	}
	return output
}

func (c *CliInterpreter) cmdStats() string {
	containers := c.engine.List()
	output := fmt.Sprintf("%-12s %-16s %-10s %-22s %-10s %-14s %-10s\n",
		"CONTAINER ID", "NAME", "CPU %", "MEM USAGE / LIMIT", "MEM %", "THROTTLED", "PIDS")

	for _, cont := range containers {
		if cont.Status != StatusRunning {
			continue
		}
		memUsedMB := float64(cont.Cgroups.MemoryUsageBytes) / (1024 * 1024)
		memLimitMB := float64(cont.Cgroups.MemoryMaxBytes) / (1024 * 1024)
		memPercent := (memUsedMB / memLimitMB) * 100

		memStr := fmt.Sprintf("%.1fMiB / %.1fMiB", memUsedMB, memLimitMB)
		output += fmt.Sprintf("%-12s %-16s %-10.2f %-22s %-10.1f %-14d %-10d\n",
			cont.ID[:12], cont.Name, cont.Cgroups.CpuPercent, memStr, memPercent,
			cont.Cgroups.ThrottlePeriods, len(cont.Processes))
	}
	return output
}

func (c *CliInterpreter) cmdRun(args []string) CliCommandResponse {
	req := CreateContainerRequest{
		Image:         "alpine:3.19",
		Command:       "sh",
		MemoryLimitMB: 128,
		CpuQuotaCores: 1.0,
		EnablePidNs:   true,
		EnableUtsNs:   true,
		EnableNetNs:   true,
		EnableMntNs:   true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--detach":
			// detach flag
		case (arg == "--name" || arg == "-n") && i+1 < len(args):
			req.Name = args[i+1]
			i++
		case (arg == "-m" || arg == "--memory") && i+1 < len(args):
			memStr := strings.ToLower(args[i+1])
			if strings.HasSuffix(memStr, "m") {
				val, _ := strconv.ParseInt(strings.TrimSuffix(memStr, "m"), 10, 64)
				req.MemoryLimitMB = val
			} else if strings.HasSuffix(memStr, "g") {
				val, _ := strconv.ParseInt(strings.TrimSuffix(memStr, "g"), 10, 64)
				req.MemoryLimitMB = val * 1024
			}
			i++
		case (arg == "--cpus" || arg == "-c") && i+1 < len(args):
			cpus, _ := strconv.ParseFloat(args[i+1], 64)
			req.CpuQuotaCores = cpus
			i++
		case (arg == "-p" || arg == "--publish") && i+1 < len(args):
			req.PortBindings = append(req.PortBindings, args[i+1])
			i++
		case (arg == "-h" || arg == "--hostname") && i+1 < len(args):
			req.Hostname = args[i+1]
			i++
		case (arg == "-e" || arg == "--env") && i+1 < len(args):
			req.Env = append(req.Env, args[i+1])
			i++
		default:
			if !strings.HasPrefix(arg, "-") {
				if req.Image == "alpine:3.19" && arg != "alpine:3.19" {
					req.Image = arg
				} else {
					req.Command = strings.Join(args[i:], " ")
					break
				}
			}
		}
	}

	newContainer, err := c.engine.Create(req)
	if err != nil {
		return CliCommandResponse{Error: err.Error(), ExitCode: 1}
	}
	_ = c.engine.Start(newContainer.ID)

	return CliCommandResponse{Output: newContainer.ID, ExitCode: 0}
}

func (c *CliInterpreter) helpText() string {
	return `Minijail-Go / Docker CLI Simulator
Available Commands:
  run      Run a command in a new container (e.g. docker run -d --name web -m 64m alpine:3.19 sh)
  ps       List active containers (use 'docker ps -a' for all)
  stop     Stop one or more running containers
  start    Start one or more stopped containers
  restart  Restart a container
  kill     Kill one or more running containers (SIGKILL)
  rm       Remove one or more containers
  logs     Fetch the logs of a container
  stats    Display a live stream of container resource usage statistics
  top      Display the running processes of a container (host vs isolated PID)
  exec     Run a command in a running container (e.g. docker exec web uname -a)
  inspect  Return low-level information on Docker objects
  images   List images
  version  Show the Docker version information`
}
