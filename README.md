<div align="center">

# Minijail-Go
> Low-Level Docker Container Runtime Simulator, Linux Kernel Primitives, and OverlayFS Playground

[![Vue 3](https://img.shields.io/badge/Frontend-Vue%203.5%20%7C%20Vite-4FC08D?style=for-the-badge&logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/Language-TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Golang](https://img.shields.io/badge/Runtime-Golang%20Engine-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Tailwind CSS](https://img.shields.io/badge/Styling-Tailwind%20CSS-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)](https://tailwindcss.com/)

<p align="center">
  <b>An authentic Docker Desktop simulator bridging low-level Linux kernel primitives—Namespaces, Cgroups v2, and OverlayFS Copy-on-Write—with a fully functional Golang container engine and rich interactive dashboard.</b>
</p>

[Overview](#overview) · [Key Features](#key-features) · [Architecture Animation](#architecture-animation) · [Tech Stack](#tech-stack--architecture) · [Getting Started](#getting-started) · [Under The Hood](#how-it-works-under-the-hood)

</div>

---

## Overview

**Minijail** is an interactive, browser-based container runtime engine and educational playground designed to demystify how containerization works under the hood. While tools like Docker abstract away system complexity behind a single command, Minijail opens up the black box. 

It pairs a high-performance **Golang runtime backend**—which simulates real Linux kernel primitives like namespaces (`PID`, `UTS`, `NET`, `MNT`), hierarchical resource limits (`cgroups v2`), and Copy-on-Write layers (`OverlayFS`)—with a sleek, production-grade **Vue 3 Docker Desktop interface**. Whether you want to monitor live CPU/memory metrics, trigger simulated kernel OOM (Out-Of-Memory) kills, or inspect isolated mount filesystems, Minijail gives you full runtime visibility.

---

## Architecture 

```text
       [ User / Browser Dashboard (Vue 3) ]
                     │
         REST API (HTTP /api/containers)
                     ▼
       ┌──────────────────────────────┐
       │   Minijail Golang Engine     │
       │  ┌────────────────────────┐  │
       │  │ Namespace Isolation    │  │──► [PID 1 / UTS / NET / MNT]
       │  ├────────────────────────┤  │
       │  │ Cgroups v2 Controller  │  │──► [CPU CFS / Memory Max Limits]
       │  ├────────────────────────┤  │
       │  │ OverlayFS Sandbox      │  │──► [LowerDir / UpperDir Merged]
       │  └────────────────────────┘  │
       └──────────────┬───────────────┘
                      │
            Kernel Simulation Loop
                      ▼
         [ Container Lifecycle Active ]

```

---

## Key Features

### 1. Authentic Docker Desktop UI

* **Dual View Modes**: Switch seamlessly between the classic **Docker Desktop Table View** and **Card Grid View**.
* **Interactive Container Management**: Start, Pause (`SIGSTOP`), Unpause (`SIGCONT`), Stop (`SIGTERM`), Kill (`SIGKILL`), and Delete containers with instant state updates.
* **Bulk Operations**: Select multiple containers to perform batch start, stop, or removal.
* **Status Indicators**: Real-time visual pulses for container states: `running` (emerald), `paused` (amber), `stopped` (slate), and `oom_killed` (rose).
* **Resource Meters**: Live sidebar dials tracking global CPU core CFS allocations and physical memory limits.

### 2. Deep Container Inspector (5 Tabs)

Click any container row or card to slide open the comprehensive container drawer:

* **Logs**: Color-coded, auto-scrolling stdout/stderr and kernel lifecycle events with live search filtering.
* **Exec (Terminal)**: Interactive in-container shell (`root@container:/#`) with pre-set diagnostic commands (`ps -ef`, `cat /proc/1/status`, `ip addr`, `cgroups info`).
* **Inspect**: Full JSON configuration matching `docker inspect`, including virtual network interfaces, namespace inodes, and cgroup filesystem mounts.
* **Files (OverlayFS)**: File system browser displaying lower read-only rootfs layers vs. upper read-write Copy-on-Write diffs with the ability to touch new files.
* **Stats & Cgroups**: Real-time meters for `memory.current` vs `memory.max` and CPU CFS quota usage, plus an interactive **Simulate OOM-Killer** button.

### 3. Docker CLI Console

* Built-in floating terminal supporting native Docker commands:
* `docker ps` / `docker ps -a`
* `docker stats`
* `docker run -d --name <name> -m <limit> <image> <cmd>`
* `docker stop <id>` / `docker kill <id>`
* `docker inspect <id>`
* `docker images`
* `docker version`


* Command history navigation using ↑ and ↓ arrow keys.

### 4. Under The Hood: Kernel Primitives

An educational interactive guide exploring how Docker translates high-level commands into Linux system calls:

* **`CLONE_NEWPID`**: Process ID isolation, giving containers their own isolated PID 1.
* **`CLONE_NEWUTS`**: Hostname and domain name virtualization.
* **`CLONE_NEWNET`**: Network namespace with independent loopback and `veth` virtual ethernet pairs.
* **`CLONE_NEWNS` & `pivot_root**`: Mount namespace isolation swapping the host rootfs with an OverlayFS root.
* **`cgroups v2` (`memory.max` & `cpu.max`)**: Hierarchical resource controls and CFS bandwidth throttling.

### 5. Images & Storage Volumes

* **Base Images Catalog**: Inspect layer sizes, hashes, and run containers directly from `alpine:3.19`, `busybox:1.36`, `ubuntu:22.04`, and `nginx:alpine`.
* **Volume & Mounts Explorer**: Inspect OverlayFS directories (`LowerDir`, `UpperDir`, `MergedDir`, `WorkDir`).

---

## Tech Stack & Architecture

| Layer | Technologies & Components |
| --- | --- |
| **Frontend** | • [Vue 3](Composition API, `<script setup>`)<br>• [TypeScript]<br>• [Tailwind CSS] (Docker Desktop authentic dark styling)<br>• [Lucide Icons] (`lucide-vue-next`)<br>• [Vite] |
| **Backend Runtime Engine** | • [Golang] (`backend/main.go` & `backend/runtime/*`)<br>• RESTful API with automated cgroup metrics, simulated process trees, and CLI engine parser.<br>• Proxying configured seamlessly through Vite (`/api` → `http://127.0.0.1:9090`). |

---

## Available Scripts

| Command | Description |
| --- | --- |
| `npm run dev` | Starts Vite dev server and automatically spawns the Go engine. |
| `npm run build` | Compiles the Go engine binary and builds the production Vue 3 static assets into `dist/`. |
| `npm run lint` | Runs TypeScript type checking (`tsc --noEmit`). |
| `npm run preview` | Previews the production build locally. |
| `npm run clean` | Cleans up built artifacts in `dist/` and `bin/`. |

---

## How It Works Under The Hood

1. **Creating a Container**: The spawner allocates a container structure with virtual PID 1, generates unique namespace descriptors (`CLONE_NEWPID`, `CLONE_NEWUTS`, `CLONE_NEWNET`, `CLONE_NEWNS`), sets up OverlayFS mount paths, and binds cgroup limits.
2. **Resource Constraints**: When memory allocation exceeds `memory.max`, the engine triggers an authentic Linux kernel OOM event (`SIGKILL`), updating the container status to `oom_killed` and appending diagnostic logs.
3. **Interactive Exec**: Executes simulated Unix utilities inside the container's isolated context with environment variable injection and virtual process hierarchies.

---

## Getting Started

### Prerequisites

* **Node.js**: v18.0.0 or higher
* **Go**: v1.20 or higher (for building the native container engine)
* **npm** or **yarn**

### Installation

1. Clone or navigate to the repository:
```bash
git clone https://github.com/mhdhamka/Minijail-Go.git
cd Minijail-Go

```


2. Install frontend dependencies:
```bash
npm install

```


3. Build the Go engine binary:
```bash
mkdir -p bin
go build -o bin/minijail-engine ./backend/main.go

```


4. Start the development server:
```bash
npm run dev

```

The Vite dev server will launch on port `3000` and automatically spawn the Go runtime backend on port `9090`.
5. Open your browser and navigate to:
```text
http://localhost:3000

```

---

## Contributing

Contributions, feedback, and regional sensor integrations are warmly welcomed!

1. Fork the Project (`https://github.com/mhdhamka/Minijail-Go/fork`)
2. Create your Feature Branch (`git checkout -b feature/NewKernelPrimitive`)
3. Commit your Changes (`git commit -m 'Add support for additional minijail simulation'`)
4. Push to the Branch (`git push origin feature/NewKernelPrimitive`)
5. Open a Pull Request

---

## License

Distributed under the **MIT License**. See `LICENSE` for more information.

<div align="center">
  <br>
  <sub>Engineered with precision for low-level systems programming and container runtime exploration.</sub>
  <br>
  <b>Developed & Maintained by <a href="https://github.com/mhdhamka">mdhamka</a></b>
</div>
