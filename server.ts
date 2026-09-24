import express from 'express';
import type { Request, Response } from 'express';
import { createServer as createViteServer } from 'vite';
import path from 'path';
import { fileURLToPath } from 'url';
import crypto from 'crypto';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// --- Types ---
export type ContainerStatus = 'created' | 'running' | 'paused' | 'stopped' | 'oom_killed';

export interface LogEntry {
  timestamp: string;
  stream: 'stdout' | 'stderr' | 'system';
  message: string;
}

export interface ProcessInfo {
  hostPid: number;
  containerPid: number;
  user: string;
  cpu: number;
  memoryMb: number;
  command: string;
}

export interface NamespaceInfo {
  pidInode: string;
  utsInode: string;
  mntInode: string;
  netInode: string;
  ipcInode: string;
  userInode: string;
  cgroupInode: string;
  hostname: string;
  virtualIp: string;
  vethPair: string;
}

export interface CgroupLimits {
  cgroupPath: string;
  memoryMaxBytes: number;
  memoryUsageBytes: number;
  memoryMaxUsage: number;
  cpuQuotaUs: number;
  cpuPeriodUs: number;
  cpuShares: number;
  cpuPercent: number;
  throttlePeriods: number;
  throttledTimeUs: number;
  oomKillEvents: number;
  oomScoreAdj: number;
}

export interface OverlayFileDiff {
  path: string;
  action: string;
  sizeBytes: number;
  timestamp: string;
}

export interface RootfsOverlay {
  baseImage: string;
  lowerDir: string;
  upperDir: string;
  workDir: string;
  mergedDir: string;
  modifiedFiles: OverlayFileDiff[];
}

export interface MetricPoint {
  timestamp: string;
  cpuPercent: number;
  cpuQuotaPercent: number;
  memoryUsageBytes: number;
  memoryMaxBytes: number;
  isThrottled: boolean;
  throttlePeriods: number;
  throttledTimeUs: number;
  oomKillEvents: number;
}

export interface Container {
  id: string;
  name: string;
  image: string;
  command: string;
  status: ContainerStatus;
  createdAt: string;
  startedAt?: string;
  hostPid: number;
  containerPid: number;
  portBindings: string[];
  env: string[];
  namespaces: NamespaceInfo;
  cgroups: CgroupLimits;
  rootfs: RootfsOverlay;
  processes: ProcessInfo[];
  logs: LogEntry[];
  metricsHistory: MetricPoint[];
}

export interface CreateContainerRequest {
  name?: string;
  image?: string;
  command?: string;
  memoryLimitMb?: number;
  cpuQuotaCores?: number;
  enablePidNs?: boolean;
  enableUtsNs?: boolean;
  enableNetNs?: boolean;
  enableMntNs?: boolean;
  hostname?: string;
  portBindings?: string[];
  env?: string[];
}

// --- Helper Functions ---
function getTimestamp(offsetSec = 0): string {
  const d = new Date(Date.now() + offsetSec * 1000);
  return d.toTimeString().split(' ')[0];
}

function generateID(): string {
  return crypto.randomBytes(32).toString('hex');
}

function randomHostPID(): number {
  return Math.floor(Math.random() * 50000) + 12000;
}

function generateNamespaceInodes(containerId: string): NamespaceInfo {
  const baseInode = 4026532000 + Math.floor(Math.random() * 10000);
  const shortID = containerId.substring(0, 6);
  const ipOctet = Math.floor(Math.random() * 250) + 2;

  return {
    pidInode: `pid:[${baseInode + 1}]`,
    utsInode: `uts:[${baseInode + 2}]`,
    mntInode: `mnt:[${baseInode + 3}]`,
    netInode: `net:[${baseInode + 4}]`,
    ipcInode: `ipc:[${baseInode + 5}]`,
    userInode: `user:[${baseInode + 6}]`,
    cgroupInode: `cgroup:[${baseInode + 7}]`,
    hostname: `container-${shortID}`,
    virtualIp: `172.18.0.${ipOctet}/16`,
    vethPair: `veth_${shortID} <-> eth0`,
  };
}

function initCgroupLimits(id: string, memoryMB: number, cpuCores: number): CgroupLimits {
  const memMb = memoryMB > 0 ? memoryMB : 128;
  const cores = cpuCores > 0 ? cpuCores : 1.0;
  const memBytes = memMb * 1024 * 1024;
  const periodUs = 100000;
  const quotaUs = Math.floor(cores * periodUs);
  const baselineMem = (8 + Math.floor(Math.random() * 10)) * 1024 * 1024;

  return {
    cgroupPath: `/sys/fs/cgroup/minijail/${id}`,
    memoryMaxBytes: memBytes,
    memoryUsageBytes: baselineMem,
    memoryMaxUsage: baselineMem,
    cpuQuotaUs: quotaUs,
    cpuPeriodUs: periodUs,
    cpuShares: 1024,
    cpuPercent: +(Math.random() * 10 + 3).toFixed(1),
    throttlePeriods: 0,
    throttledTimeUs: 0,
    oomKillEvents: 0,
    oomScoreAdj: 0,
  };
}

function initRootfs(id: string, image: string): RootfsOverlay {
  const shortID = id.substring(0, 8);
  const sanitizedImage = image || 'alpine:3.19';

  return {
    baseImage: sanitizedImage,
    lowerDir: `/var/lib/minijail/images/${sanitizedImage}/rootfs`,
    upperDir: `/var/lib/minijail/containers/${shortID}/diff`,
    workDir: `/var/lib/minijail/containers/${shortID}/work`,
    mergedDir: `/var/lib/minijail/containers/${shortID}/merged`,
    modifiedFiles: [
      { path: '/etc/hostname', action: 'CREATED', sizeBytes: 14, timestamp: getTimestamp() },
      { path: '/etc/hosts', action: 'CREATED', sizeBytes: 184, timestamp: getTimestamp() },
      { path: '/etc/resolv.conf', action: 'CREATED', sizeBytes: 62, timestamp: getTimestamp() },
    ],
  };
}

function generateSeedHistory(c: Container, count = 25): MetricPoint[] {
  const points: MetricPoint[] = [];
  const maxQuota = (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs) * 100.0;
  const now = Date.now();
  for (let i = count; i >= 0; i--) {
    const d = new Date(now - i * 2000);
    const t = d.toTimeString().split(' ')[0];
    const isRunning = c.status === 'running';
    let cpu = 0;
    let isThrottled = false;
    if (isRunning) {
      const base = Math.min(c.cgroups.cpuPercent, maxQuota);
      cpu = Math.max(1.0, +(base + (Math.random() * 4 - 2)).toFixed(1));
      if (cpu >= maxQuota) {
        cpu = maxQuota;
        isThrottled = true;
      }
    }
    const mem = isRunning ? Math.round(c.cgroups.memoryUsageBytes * (0.96 + Math.random() * 0.08)) : 0;
    points.push({
      timestamp: t,
      cpuPercent: cpu,
      cpuQuotaPercent: maxQuota,
      memoryUsageBytes: mem,
      memoryMaxBytes: c.cgroups.memoryMaxBytes,
      isThrottled,
      throttlePeriods: c.cgroups.throttlePeriods,
      throttledTimeUs: c.cgroups.throttledTimeUs,
      oomKillEvents: c.cgroups.oomKillEvents,
    });
  }
  return points;
}

// --- Container Engine Simulation ---
class ContainerEngine {
  containers: Map<string, Container> = new Map();

  constructor() {
    this.seedInitialContainers();
    this.startBackgroundWorker();
  }

  private seedInitialContainers() {
    // 1. alpine-web
    const c1Id = 'c7a8f912e34b1509d845e67041ab34fe20194851239857ab90e812d46e10f391';
    const c1HostPid = randomHostPID();
    const c1CreatedAt = new Date(Date.now() - 27 * 60000).toISOString();
    const c1StartedAt = new Date(Date.now() - 25 * 60000).toISOString();
    const c1Ns = generateNamespaceInodes(c1Id);
    c1Ns.hostname = 'alpine-web';

    const c1: Container = {
      id: c1Id,
      name: 'alpine-web',
      image: 'alpine:3.19',
      command: "sh -c 'while true; do echo \"[HTTP 200] GET /api/v1/health - 12ms\"; sleep 4; done'",
      status: 'running',
      createdAt: c1CreatedAt,
      startedAt: c1StartedAt,
      hostPid: c1HostPid,
      containerPid: 1,
      portBindings: ['8080:80/tcp'],
      env: ['PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin', 'ENV=production'],
      namespaces: c1Ns,
      cgroups: initCgroupLimits(c1Id, 64, 0.5),
      rootfs: initRootfs(c1Id, 'alpine:3.19'),
      processes: [
        { hostPid: c1HostPid, containerPid: 1, user: 'root', cpu: 2.1, memoryMb: 18.4, command: 'sh -c while true; do echo [HTTP 200]...' },
        { hostPid: c1HostPid + 1, containerPid: 8, user: 'root', cpu: 0.1, memoryMb: 4.2, command: 'sleep 4' },
      ],
      logs: [
        { timestamp: getTimestamp(-1500), stream: 'system', message: 'Mounted OverlayFS lowerdir=/var/lib/minijail/images/alpine:3.19/rootfs,upperdir=/var/lib/minijail/containers/c7a8f912/diff' },
        { timestamp: getTimestamp(-1499), stream: 'system', message: 'pivot_root into /var/lib/minijail/containers/c7a8f912/merged; old root unmounted with MNT_DETACH' },
        { timestamp: getTimestamp(-1498), stream: 'system', message: 'Linux cgroups v2 applied: memory.max=64MB, cpu.max=50000/100000 (0.50 cores)' },
        { timestamp: getTimestamp(-1497), stream: 'stdout', message: 'Server listening on 0.0.0.0:80 (veth_c7a8 -> eth0, IP: 172.18.0.2)' },
        { timestamp: getTimestamp(-10), stream: 'stdout', message: '[HTTP 200] GET /api/v1/health - 12ms' },
      ],
      metricsHistory: [],
    };
    c1.cgroups.memoryUsageBytes = 22 * 1024 * 1024;
    c1.cgroups.cpuPercent = 2.1;
    c1.metricsHistory = generateSeedHistory(c1);

    // 2. worker-daemon
    const c2Id = '9f14b2d8e0917234aa6b71f021c38941da782910e54361ab8920194bcde5912a';
    const c2HostPid = randomHostPID();
    const c2CreatedAt = new Date(Date.now() - 13 * 60000).toISOString();
    const c2StartedAt = new Date(Date.now() - 12 * 60000).toISOString();
    const c2Ns = generateNamespaceInodes(c2Id);
    c2Ns.hostname = 'worker-daemon';

    const c2: Container = {
      id: c2Id,
      name: 'worker-daemon',
      image: 'busybox:1.36',
      command: "sh -c 'while true; do echo \"[Job-Queue] Processed payload chunk\"; sleep 8; done'",
      status: 'running',
      createdAt: c2CreatedAt,
      startedAt: c2StartedAt,
      hostPid: c2HostPid,
      containerPid: 1,
      portBindings: [],
      env: ['PATH=/bin:/sbin', 'WORKER_THREADS=4'],
      namespaces: c2Ns,
      cgroups: initCgroupLimits(c2Id, 128, 1.0),
      rootfs: initRootfs(c2Id, 'busybox:1.36'),
      processes: [
        { hostPid: c2HostPid, containerPid: 1, user: 'root', cpu: 1.4, memoryMb: 12.1, command: 'sh -c while true...' },
      ],
      logs: [
        { timestamp: getTimestamp(-720), stream: 'system', message: 'Namespaces cloned: CLONE_NEWPID | CLONE_NEWUTS | CLONE_NEWNS | CLONE_NEWNET' },
        { timestamp: getTimestamp(-719), stream: 'stdout', message: 'Worker initialized. Subscribed to internal event bus.' },
        { timestamp: getTimestamp(-20), stream: 'stdout', message: '[Job-Queue] Processed payload chunk batch #148' },
      ],
      metricsHistory: [],
    };
    c2.cgroups.memoryUsageBytes = 16 * 1024 * 1024;
    c2.cgroups.cpuPercent = 1.4;
    c2.metricsHistory = generateSeedHistory(c2);

    // 3. redis-cache (stopped)
    const c3Id = 'e34b1509d845e67041ab34fe20194851239857ab90e812d46e10f391c7a8f912';
    const c3CreatedAt = new Date(Date.now() - 45 * 60000).toISOString();
    const c3Ns = generateNamespaceInodes(c3Id);
    c3Ns.hostname = 'redis-cache';

    const c3: Container = {
      id: c3Id,
      name: 'redis-cache',
      image: 'alpine:3.19',
      command: 'redis-server --protected-mode no',
      status: 'stopped',
      createdAt: c3CreatedAt,
      hostPid: 0,
      containerPid: 0,
      portBindings: ['6379:6379/tcp'],
      env: ['ALLOW_EMPTY_PASSWORD=yes'],
      namespaces: c3Ns,
      cgroups: initCgroupLimits(c3Id, 256, 1.5),
      rootfs: initRootfs(c3Id, 'alpine:3.19'),
      processes: [],
      logs: [
        { timestamp: getTimestamp(-2700), stream: 'system', message: 'Container created with PID/UTS/MNT namespaces' },
        { timestamp: getTimestamp(-2580), stream: 'system', message: 'Received SIGTERM (stop requested). Graceful shutdown complete.' },
      ],
      metricsHistory: [],
    };
    c3.cgroups.memoryUsageBytes = 0;
    c3.cgroups.cpuPercent = 0;
    c3.metricsHistory = generateSeedHistory(c3);

    this.containers.set(c1.id, c1);
    this.containers.set(c2.id, c2);
    this.containers.set(c3.id, c3);
  }

  list(): Container[] {
    return Array.from(this.containers.values());
  }

  get(idOrName: string): Container | undefined {
    for (const [id, c] of this.containers.entries()) {
      if (id === idOrName || id.startsWith(idOrName) || c.name === idOrName) {
        return c;
      }
    }
    return undefined;
  }

  create(req: CreateContainerRequest): Container {
    const id = generateID();
    const name = req.name || `container-${id.substring(0, 8)}`;
    const image = req.image || 'alpine:3.19';
    const command = req.command || 'sh';
    const memLimitMb = req.memoryLimitMb && req.memoryLimitMb > 0 ? req.memoryLimitMb : 128;
    const cpuCores = req.cpuQuotaCores && req.cpuQuotaCores > 0 ? req.cpuQuotaCores : 1.0;

    for (const c of this.containers.values()) {
      if (c.name === name) {
        throw new Error(`conflict: container name '${name}' is already in use`);
      }
    }

    const ns = generateNamespaceInodes(id);
    ns.hostname = req.hostname || name;

    const now = new Date().toISOString();
    const cont: Container = {
      id,
      name,
      image,
      command,
      status: 'created',
      createdAt: now,
      portBindings: req.portBindings || [],
      env: req.env || [],
      namespaces: ns,
      cgroups: initCgroupLimits(id, memLimitMb, cpuCores),
      rootfs: initRootfs(id, image),
      processes: [],
      logs: [
        { timestamp: getTimestamp(), stream: 'system', message: `Created container ${id.substring(0, 12)} with image ${image}` },
        { timestamp: getTimestamp(), stream: 'system', message: `Cgroups v2 assigned: memory.max=${memLimitMb}MB, cpu.max=${cpuCores.toFixed(2)} cores` },
      ],
      hostPid: 0,
      containerPid: 0,
      metricsHistory: [],
    };
    cont.metricsHistory = generateSeedHistory(cont, 15);

    this.containers.set(id, cont);
    return cont;
  }

  start(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status === 'running') throw new Error('container is already running');

    const hostPid = randomHostPID();
    c.hostPid = hostPid;
    c.containerPid = 1;
    c.status = 'running';
    c.startedAt = new Date().toISOString();

    c.cgroups.memoryUsageBytes = 14 * 1024 * 1024;
    c.cgroups.cpuPercent = 3.5;
    c.processes = [
      { hostPid, containerPid: 1, user: 'root', cpu: 3.5, memoryMb: 14.0, command: c.command },
    ];

    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: `Kernel clone invoked: host PID=${hostPid} mapped to container PID=1 in ${c.namespaces.pidInode}`,
    });
    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: `Veth pair established: ${c.namespaces.vethPair} (IP: ${c.namespaces.virtualIp})`,
    });
    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'stdout',
      message: `Process started: ${c.command}`,
    });

    return c;
  }

  stop(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'running' && c.status !== 'paused') {
      throw new Error('container is not running');
    }

    c.status = 'stopped';
    c.hostPid = 0;
    c.containerPid = 0;
    c.processes = [];
    c.cgroups.cpuPercent = 0;
    c.cgroups.memoryUsageBytes = 0;

    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: 'Sent SIGTERM to container PID 1. Gracefully stopped.',
    });

    return c;
  }

  pause(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'running') throw new Error('can only pause running containers');

    c.status = 'paused';
    c.cgroups.cpuPercent = 0;
    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: 'cgroup.freeze = 1 applied: all threads in container suspended',
    });
    return c;
  }

  unpause(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'paused') throw new Error('container is not paused');

    c.status = 'running';
    c.cgroups.cpuPercent = 2.5;
    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: 'cgroup.freeze = 0 applied: threads unfrozen',
    });
    return c;
  }

  kill(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);

    c.status = 'stopped';
    c.hostPid = 0;
    c.containerPid = 0;
    c.processes = [];
    c.cgroups.cpuPercent = 0;
    c.cgroups.memoryUsageBytes = 0;

    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: 'Sent SIGKILL to container process tree. Process terminated immediately (exit 137).',
    });
    return c;
  }

  remove(idOrName: string): void {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status === 'running') {
      throw new Error('cannot remove a running container. Stop or kill it first');
    }
    this.containers.delete(c.id);
  }

  stressMemory(idOrName: string, deltaMB = 24): { container: Container; oomKilled: boolean } {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'running') throw new Error('container must be running to stress memory');

    const deltaBytes = deltaMB * 1024 * 1024;
    c.cgroups.memoryUsageBytes += deltaBytes;
    if (c.cgroups.memoryUsageBytes > c.cgroups.memoryMaxUsage) {
      c.cgroups.memoryMaxUsage = c.cgroups.memoryUsageBytes;
    }

    const oomTriggered = c.cgroups.memoryUsageBytes >= c.cgroups.memoryMaxBytes;

    if (oomTriggered) {
      c.cgroups.oomKillEvents++;
      c.cgroups.memoryUsageBytes = c.cgroups.memoryMaxBytes;
      c.status = 'oom_killed';
      c.hostPid = 0;
      c.containerPid = 0;
      c.processes = [];
      c.cgroups.cpuPercent = 0;

      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'stderr',
        message: `[cgroups v2] memory.max (${Math.round(c.cgroups.memoryMaxBytes / (1024 * 1024))} MB) exceeded!`,
      });
      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'system',
        message: '[kernel OOM-Killer] Memory cgroup out of memory: Killed process 1 (sh) total-vm:131072kB, anon-rss:65536kB, file-rss:0kB',
      });
      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'system',
        message: 'Container exited with code 137 (Fatal error signal 9 - SIGKILL due to OOM)',
      });
    } else {
      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'stdout',
        message: `Allocated +${deltaMB}MB in anonymous RSS. Total usage: ${Math.round(c.cgroups.memoryUsageBytes / (1024 * 1024))}MB / ${Math.round(c.cgroups.memoryMaxBytes / (1024 * 1024))}MB`,
      });
    }

    return { container: c, oomKilled: oomTriggered };
  }

  stressCpu(idOrName: string): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'running') throw new Error('container must be running to stress CPU');

    const maxAllowed = (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs) * 100.0;
    c.cgroups.cpuPercent = maxAllowed;
    c.cgroups.throttlePeriods += Math.floor(Math.random() * 5) + 3;
    c.cgroups.throttledTimeUs += Math.floor(Math.random() * 12000) + 5000;

    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'stdout',
      message: `Spawned CPU stress worker threads. Pegged at CFS quota (${maxAllowed.toFixed(2)}%)`,
    });
    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'system',
      message: `[cgroups cpu.stat] nr_throttled=${c.cgroups.throttlePeriods} throttled_usec=${c.cgroups.throttledTimeUs} (CFS scheduler actively enforced quota)`,
    });

    return c;
  }

  exec(idOrName: string, command: string): { exitCode: number; stdout: string; stderr: string } {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);
    if (c.status !== 'running') throw new Error('container is not running');

    const trimmed = command.trim();
    const parts = trimmed.split(/\s+/);
    if (!trimmed || parts.length === 0) {
      return { exitCode: 0, stdout: '', stderr: '' };
    }

    const baseCmd = parts[0];
    let stdout = '';
    const stderr = '';
    let exitCode = 0;

    switch (baseCmd) {
      case 'uname':
        stdout = `Linux ${c.namespaces.hostname} 4.19.0-minijail #1 SMP x86_64 GNU/Linux\n`;
        break;
      case 'hostname':
        stdout = `${c.namespaces.hostname}\n`;
        break;
      case 'id':
      case 'whoami':
        stdout = 'uid=0(root) gid=0(root) groups=0(root),1(bin),2(daemon),10(wheel)\n';
        break;
      case 'uptime':
        stdout = ' 15:42:10 up 25 min,  0 users,  load average: 0.12, 0.08, 0.04\n';
        break;
      case 'ip':
      case 'ifconfig':
        stdout = `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN\n    inet 127.0.0.1/8 scope host lo\n2: eth0@if14: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 state UP\n    inet ${c.namespaces.virtualIp} scope global eth0\n    link/ether 02:42:ac:12:00:02 brd ff:ff:ff:ff:ff:ff\n`;
        break;
      case 'cat':
        if (parts.length > 1) {
          const target = parts[1];
          if (target === '/etc/os-release' || target === '/etc/alpine-release') {
            if (c.image.includes('alpine')) {
              stdout = 'NAME="Alpine Linux"\nID=alpine\nVERSION_ID=3.19.1\nPRETTY_NAME="Alpine Linux v3.19"\nHOME_URL="https://alpinelinux.org/"\n';
            } else if (c.image.includes('ubuntu')) {
              stdout = 'NAME="Ubuntu"\nVERSION="22.04.4 LTS (Jammy Jellyfish)"\nID=ubuntu\nPRETTY_NAME="Ubuntu 22.04.4 LTS"\n';
            } else {
              stdout = 'BusyBox v1.36.1 multi-call binary.\n';
            }
          } else if (target === '/etc/hostname') {
            stdout = `${c.namespaces.hostname}\n`;
          } else if (target === '/etc/hosts') {
            stdout = `127.0.0.1\tlocalhost\n::1\tlocalhost ip6-localhost ip6-loopback\n${c.namespaces.virtualIp.split('/')[0]}\t${c.namespaces.hostname}\n`;
          } else if (target === '/etc/resolv.conf') {
            stdout = 'nameserver 127.0.0.11\noptions ndots:0\n';
          } else if (target === '/proc/1/cmdline') {
            stdout = `${c.command}\n`;
          } else if (target === '/sys/fs/cgroup/memory.max') {
            stdout = `${c.cgroups.memoryMaxBytes}\n`;
          } else {
            stdout = `cat: ${target}: No such file or directory\n`;
            exitCode = 1;
          }
        }
        break;
      case 'ps':
        stdout = 'PID   USER     TIME  COMMAND\n';
        for (const p of c.processes) {
          stdout += `${p.containerPid.toString().padEnd(5)} ${p.user.padEnd(8)} 0:00 ${p.command}\n`;
        }
        break;
      case 'env':
        stdout = c.env.join('\n') + '\n';
        break;
      case 'touch':
        if (parts.length > 1) {
          const filePath = parts[1];
          c.rootfs.modifiedFiles.push({
            path: filePath,
            action: 'CREATED',
            sizeBytes: filePath.length * 8,
            timestamp: getTimestamp(),
          });
          stdout = '';
        }
        break;
      case 'ls':
        stdout = 'bin   dev   etc   home  lib   media mnt   opt   proc  root  run   sbin  srv   sys   tmp   usr   var\n';
        break;
      case 'echo':
        stdout = parts.slice(1).join(' ') + '\n';
        break;
      case 'df':
        stdout = 'Filesystem           1K-blocks      Used Available Use% Mounted on\noverlay               61255492   4194304  53937188   7% /\ntmpfs                    65536         0     65536   0% /dev\nshm                      65536         0     65536   0% /dev/shm\n';
        break;
      default:
        stdout = `${command}: command executed inside container PID 1 namespace\n`;
    }

    c.logs.push({
      timestamp: getTimestamp(),
      stream: 'stdout',
      message: `$ ${command} -> ${stdout.trim()}`,
    });

    return { exitCode, stdout, stderr };
  }

  updateCgroupLimits(idOrName: string, updates: { cpuQuotaCores?: number; memoryLimitMb?: number }): Container {
    const c = this.get(idOrName);
    if (!c) throw new Error(`no such container: ${idOrName}`);

    if (typeof updates.cpuQuotaCores === 'number' && updates.cpuQuotaCores > 0) {
      c.cgroups.cpuQuotaUs = Math.floor(updates.cpuQuotaCores * c.cgroups.cpuPeriodUs);
      const newQuotaPercent = (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs) * 100.0;
      if (c.cgroups.cpuPercent > newQuotaPercent) {
        c.cgroups.cpuPercent = newQuotaPercent;
        c.cgroups.throttlePeriods += 5;
        c.cgroups.throttledTimeUs += 15000;
      }
      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'system',
        message: `[cgroups v2 update] cpu.max adjusted to ${c.cgroups.cpuQuotaUs} ${c.cgroups.cpuPeriodUs} (${updates.cpuQuotaCores.toFixed(2)} cores / ${newQuotaPercent.toFixed(0)}% quota)`,
      });
    }

    if (typeof updates.memoryLimitMb === 'number' && updates.memoryLimitMb > 0) {
      const newLimitBytes = updates.memoryLimitMb * 1024 * 1024;
      c.cgroups.memoryMaxBytes = newLimitBytes;
      c.logs.push({
        timestamp: getTimestamp(),
        stream: 'system',
        message: `[cgroups v2 update] memory.max set to ${newLimitBytes} bytes (${updates.memoryLimitMb}MB)`,
      });

      // If new limit is lower than current usage, trigger OOM killer immediately!
      if (c.status === 'running' && c.cgroups.memoryUsageBytes > newLimitBytes) {
        c.status = 'oom_killed';
        c.cgroups.oomKillEvents += 1;
        c.hostPid = 0;
        c.processes = [];
        c.logs.push({
          timestamp: getTimestamp(),
          stream: 'system',
          message: `[kernel OOM-Killer] memory.max reduced below active RSS usage (${Math.round(c.cgroups.memoryUsageBytes / (1024 * 1024))}MB > ${updates.memoryLimitMb}MB). Killed PID 1 with SIGKILL.`,
        });
      }
    }

    return c;
  }

  private startBackgroundWorker() {
    setInterval(() => {
      const nowTime = getTimestamp();
      for (const c of this.containers.values()) {
        const maxAllowed = (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs) * 100.0;
        let isThrottled = false;

        if (c.status === 'running') {
          const fluctuation = (Math.random() * 5 - 2.4);
          let newPercent = c.cgroups.cpuPercent + fluctuation;
          if (newPercent < 1.0) newPercent = 1.4;
          if (newPercent >= maxAllowed) {
            newPercent = maxAllowed;
            c.cgroups.throttlePeriods += Math.floor(Math.random() * 3) + 1;
            c.cgroups.throttledTimeUs += Math.floor(Math.random() * 8000) + 2000;
            isThrottled = true;
          }
          c.cgroups.cpuPercent = +newPercent.toFixed(1);

          // Micro variations in memory
          const memNoise = Math.floor((Math.random() * 1024 - 450) * 1024);
          c.cgroups.memoryUsageBytes = Math.max(8 * 1024 * 1024, c.cgroups.memoryUsageBytes + memNoise);
          if (c.cgroups.memoryUsageBytes > c.cgroups.memoryMaxUsage) {
            c.cgroups.memoryMaxUsage = c.cgroups.memoryUsageBytes;
          }

          if (c.logs.length < 100 && Math.random() < 0.1) {
            c.logs.push({
              timestamp: nowTime,
              stream: 'stdout',
              message: `Heartbeat tick: ${c.name} alive [CPU: ${c.cgroups.cpuPercent}%, RAM: ${(c.cgroups.memoryUsageBytes / (1024 * 1024)).toFixed(1)}MiB]`,
            });
          }
        } else {
          c.cgroups.cpuPercent = 0;
        }

        if (!c.metricsHistory) c.metricsHistory = [];
        c.metricsHistory.push({
          timestamp: nowTime,
          cpuPercent: c.status === 'running' ? c.cgroups.cpuPercent : 0,
          cpuQuotaPercent: maxAllowed,
          memoryUsageBytes: c.status === 'running' ? c.cgroups.memoryUsageBytes : 0,
          memoryMaxBytes: c.cgroups.memoryMaxBytes,
          isThrottled,
          throttlePeriods: c.cgroups.throttlePeriods,
          throttledTimeUs: c.cgroups.throttledTimeUs,
          oomKillEvents: c.cgroups.oomKillEvents,
        });

        if (c.metricsHistory.length > 50) {
          c.metricsHistory.shift();
        }
      }
    }, 1800);
  }
}

// --- Docker CLI Interpreter ---
class DockerCliInterpreter {
  engine: ContainerEngine;

  constructor(engine: ContainerEngine) {
    this.engine = engine;
  }

  execute(cmdLine: string): { output: string; error?: string; exitCode: number } {
    const trimmed = cmdLine.trim();
    if (!trimmed) return { output: '', exitCode: 0 };

    const parts = trimmed.split(/\s+/);
    const root = parts[0];
    if (root !== 'docker' && root !== 'minijail') {
      return {
        output: '',
        error: `command not found: ${root}. Supported root commands: 'docker', 'minijail'`,
        exitCode: 127,
      };
    }

    if (parts.length === 1) {
      return { output: this.helpText(), exitCode: 0 };
    }

    const subCmd = parts[1];
    const args = parts.slice(2);

    switch (subCmd) {
      case 'ps': {
        const all = args.includes('-a') || args.includes('--all');
        const containers = this.engine.list();
        let out = 'CONTAINER ID   IMAGE            COMMAND              STATUS         PORTS            NAMES\n';
        let count = 0;
        for (const c of containers) {
          if (!all && c.status !== 'running') continue;
          let shortCmd = c.command;
          if (shortCmd.length > 18) shortCmd = shortCmd.substring(0, 15) + '...';
          const ports = c.portBindings.join(', ') || '-';
          let statusStr = c.status as string;
          if (c.status === 'running') statusStr = 'Up (Active)';
          else if (c.status === 'oom_killed') statusStr = 'Exited (137) OOM';
          else if (c.status === 'stopped') statusStr = 'Exited (0)';

          out += `${c.id.substring(0, 12).padEnd(14)} ${c.image.padEnd(16)} ${(`"${shortCmd}"`).padEnd(20)} ${statusStr.padEnd(14)} ${ports.padEnd(16)} ${c.name}\n`;
          count++;
        }
        if (count === 0) out += '(no active containers. Use \'docker ps -a\' to see all)\n';
        return { output: out, exitCode: 0 };
      }

      case 'run': {
        const req: CreateContainerRequest = {
          image: 'alpine:3.19',
          command: 'sh',
          memoryLimitMb: 128,
          cpuQuotaCores: 1.0,
          portBindings: [],
          env: [],
        };
        for (let i = 0; i < args.length; i++) {
          const arg = args[i];
          if (arg === '-d' || arg === '--detach') {
            // detach
          } else if ((arg === '--name' || arg === '-n') && i + 1 < args.length) {
            req.name = args[++i];
          } else if ((arg === '-m' || arg === '--memory') && i + 1 < args.length) {
            const mStr = args[++i].toLowerCase();
            if (mStr.endsWith('m')) req.memoryLimitMb = parseInt(mStr.replace('m', ''), 10);
            else if (mStr.endsWith('g')) req.memoryLimitMb = parseInt(mStr.replace('g', ''), 10) * 1024;
            else req.memoryLimitMb = parseInt(mStr, 10);
          } else if ((arg === '--cpus' || arg === '-c') && i + 1 < args.length) {
            req.cpuQuotaCores = parseFloat(args[++i]);
          } else if ((arg === '-p' || arg === '--publish') && i + 1 < args.length) {
            req.portBindings?.push(args[++i]);
          } else if ((arg === '-h' || arg === '--hostname') && i + 1 < args.length) {
            req.hostname = args[++i];
          } else if ((arg === '-e' || arg === '--env') && i + 1 < args.length) {
            req.env?.push(args[++i]);
          } else if (!arg.startsWith('-')) {
            if (req.image === 'alpine:3.19' && arg !== 'alpine:3.19') {
              req.image = arg;
            } else {
              req.command = args.slice(i).join(' ');
              break;
            }
          }
        }
        try {
          const created = this.engine.create(req);
          this.engine.start(created.id);
          return { output: created.id, exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'stop': {
        if (!args[0]) return { output: '', error: 'docker stop requires at least 1 container ID/name', exitCode: 1 };
        try {
          this.engine.stop(args[0]);
          return { output: args[0], exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'start': {
        if (!args[0]) return { output: '', error: 'docker start requires at least 1 container ID/name', exitCode: 1 };
        try {
          this.engine.start(args[0]);
          return { output: args[0], exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'restart': {
        if (!args[0]) return { output: '', error: 'docker restart requires at least 1 container ID/name', exitCode: 1 };
        try {
          try { this.engine.stop(args[0]); } catch {}
          this.engine.start(args[0]);
          return { output: args[0], exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'kill': {
        if (!args[0]) return { output: '', error: 'docker kill requires at least 1 container ID/name', exitCode: 1 };
        try {
          this.engine.kill(args[0]);
          return { output: args[0], exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'rm': {
        if (!args[0]) return { output: '', error: 'docker rm requires at least 1 container ID/name', exitCode: 1 };
        try {
          this.engine.remove(args[0]);
          return { output: args[0], exitCode: 0 };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'logs': {
        if (!args[0]) return { output: '', error: 'docker logs requires at least 1 container ID/name', exitCode: 1 };
        const c = this.engine.get(args[0]);
        if (!c) return { output: '', error: `no such container: ${args[0]}`, exitCode: 1 };
        const lines = c.logs.map(l => `[${l.timestamp}] (${l.stream}) ${l.message}`);
        return { output: lines.join('\n'), exitCode: 0 };
      }

      case 'stats': {
        const containers = this.engine.list().filter(c => c.status === 'running');
        let out = 'CONTAINER ID   NAME             CPU %      MEM USAGE / LIMIT        MEM %      THROTTLED      PIDS\n';
        for (const c of containers) {
          const memUsedMB = c.cgroups.memoryUsageBytes / (1024 * 1024);
          const memLimitMB = c.cgroups.memoryMaxBytes / (1024 * 1024);
          const memPercent = (memUsedMB / memLimitMB) * 100;
          const memStr = `${memUsedMB.toFixed(1)}MiB / ${memLimitMB.toFixed(1)}MiB`;
          out += `${c.id.substring(0, 12).padEnd(14)} ${c.name.padEnd(16)} ${c.cgroups.cpuPercent.toFixed(2).padEnd(10)} ${memStr.padEnd(24)} ${memPercent.toFixed(1).padEnd(10)} ${c.cgroups.throttlePeriods.toString().padEnd(14)} ${c.processes.length}\n`;
        }
        return { output: out, exitCode: 0 };
      }

      case 'top': {
        if (!args[0]) return { output: '', error: 'docker top requires at least 1 container ID/name', exitCode: 1 };
        const c = this.engine.get(args[0]);
        if (!c) return { output: '', error: `no such container: ${args[0]}`, exitCode: 1 };
        let out = 'UID   PID   PPID  C  STIME  TTY   TIME      CMD\n';
        for (const p of c.processes) {
          out += `${p.user.padEnd(5)} ${p.hostPid.toString().padEnd(5)} 1     ${Math.round(p.cpu).toString().padEnd(2)} 00:00  ?     00:00:01  ${p.command}\n`;
        }
        return { output: out, exitCode: 0 };
      }

      case 'inspect': {
        if (!args[0]) return { output: '', error: 'docker inspect requires at least 1 container ID/name', exitCode: 1 };
        const c = this.engine.get(args[0]);
        if (!c) return { output: '', error: `no such container: ${args[0]}`, exitCode: 1 };
        return { output: JSON.stringify(c, null, 2), exitCode: 0 };
      }

      case 'exec': {
        if (args.length < 2) return { output: '', error: 'usage: docker exec <container> <command>', exitCode: 1 };
        const target = args[0];
        let cmdIdx = 1;
        while (cmdIdx < args.length && args[cmdIdx].startsWith('-')) {
          cmdIdx++;
        }
        if (cmdIdx >= args.length) return { output: '', error: 'no command specified for exec', exitCode: 1 };
        const cmdStr = args.slice(cmdIdx).join(' ');
        try {
          const res = this.engine.exec(target, cmdStr);
          let out = res.stdout;
          if (res.stderr) out += '\n' + res.stderr;
          return { output: out, exitCode: res.exitCode };
        } catch (e: any) {
          return { output: '', error: e.message, exitCode: 1 };
        }
      }

      case 'images': {
        const out = 'REPOSITORY   TAG       IMAGE ID       SIZE\n' +
          'alpine       3.19      9a807d4b9261   7.38MB\n' +
          'busybox      1.36      e97db6746ef7   4.26MB\n' +
          'ubuntu       22.04     5921820980f7   77.8MB\n' +
          'nginx        alpine    f801646274b5   23.4MB\n';
        return { output: out, exitCode: 0 };
      }

      case 'version': {
        const out = 'Client: Minijail Engine / Docker CLI\n' +
          ' Version:           26.0.0-sim\n' +
          ' API version:       1.45 (TypeScript + Node runtime)\n' +
          ' Node version:      ' + process.version + '\n' +
          ' OS/Arch:           linux/amd64\n' +
          ' Experimental:      true\n' +
          ' Kernel Primitives: Linux Namespaces (PID, UTS, MNT, NET, IPC, USER), cgroups v2, OverlayFS\n';
        return { output: out, exitCode: 0 };
      }

      case 'help':
      case '--help':
      case '-h':
        return { output: this.helpText(), exitCode: 0 };

      default:
        return {
          output: '',
          error: `docker: '${subCmd}' is not a docker command. See 'docker --help'`,
          exitCode: 1,
        };
    }
  }

  helpText(): string {
    return `Minijail / Docker CLI Simulator
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
  version  Show the Docker version information`;
  }
}

// --- Educational Primitives Guide ---
function getKernelPrimitivesData() {
  return [
    {
      name: 'PID Namespace (Process Isolation)',
      linuxFlag: 'syscall.CLONE_NEWPID',
      syscallFile: '/proc/[pid]/ns/pid',
      goCode: `cmd := exec.Command("/proc/self/exe", "child", command)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
}
// The first spawned process inside becomes PID 1
// Ps inside only sees this child tree!`,
      explanation: 'Isolates the process ID space. The container\'s primary process becomes PID 1 (init) inside its own virtual tree, preventing it from seeing or signaling any processes running on the host OS or in other containers.',
    },
    {
      name: 'UTS Namespace (Hostname & Domain)',
      linuxFlag: 'syscall.CLONE_NEWUTS',
      syscallFile: '/proc/[pid]/ns/uts',
      goCode: `cmd.SysProcAttr.Cloneflags |= syscall.CLONE_NEWUTS
// Inside child handler:
if err := syscall.Sethostname([]byte(containerHostname)); err != nil {
    log.Fatalf("failed to set container hostname: %v", err)
}`,
      explanation: 'Isolates system identifiers including hostname and NIS domain. Setting the hostname inside the container modifies only its local UTS namespace without changing the host system\'s identity.',
    },
    {
      name: 'Mount Namespace & Pivot Root (Filesystem)',
      linuxFlag: 'syscall.CLONE_NEWNS',
      syscallFile: '/proc/[pid]/ns/mnt',
      goCode: `// 1. Remount root as private to avoid host propagation
syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
// 2. Bind mount target rootfs
syscall.Mount(newRoot, newRoot, "bind", syscall.MS_BIND|syscall.MS_REC, "")
// 3. Pivot root swap
os.MkdirAll(filepath.Join(newRoot, ".oldroot"), 0700)
syscall.PivotRoot(newRoot, filepath.Join(newRoot, ".oldroot"))
os.Chdir("/")
syscall.Unmount("/.oldroot", syscall.MNT_DETACH)
os.Remove("/.oldroot")`,
      explanation: 'Isolates the filesystem mount table. Together with pivot_root, it detaches the host filesystem and pivots the container root to an isolated OverlayFS directory, making escape to parent directories impossible.',
    },
    {
      name: 'Network Namespace (Veth & IP)',
      linuxFlag: 'syscall.CLONE_NEWNET',
      syscallFile: '/proc/[pid]/ns/net',
      goCode: `// Virtual ethernet pairs
// Host side: vethX plugged into bridge docker0
// Container side: peer moved into container netns and renamed to eth0
netlink.LinkAdd(&netlink.Veth{
    LinkAttrs: netlink.LinkAttrs{Name: "veth_c1"},
    PeerName:  "eth0",
})
netlink.LinkSetNsPid(peerLink, containerHostPid)`,
      explanation: 'Provides an isolated network stack including network device interfaces, IPv4/IPv6 addresses, IP routing tables, firewall rules (/proc/net and iptables), and sockets.',
    },
    {
      name: 'Control Groups (cgroups v2 Resource Throttling)',
      linuxFlag: '/sys/fs/cgroup/',
      syscallFile: '/sys/fs/cgroup/container/<id>/memory.max',
      goCode: `cgroupPath := "/sys/fs/cgroup/minijail/" + containerID
os.MkdirAll(cgroupPath, 0755)
// Memory limit: 64MB
os.WriteFile(cgroupPath+"/memory.max", []byte("67108864"), 0644)
// CPU quota: 50ms every 100ms (0.5 core)
os.WriteFile(cgroupPath+"/cpu.max", []byte("50000 100000"), 0644)
// Add container PID to cgroup tasks
os.WriteFile(cgroupPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)`,
      explanation: 'Linux cgroups v2 restricts the hardware consumption (CPU time quotas via CFS scheduler, Memory limits, and PID forks) of a process hierarchy and triggers the kernel OOM Killer if thresholds are exceeded.',
    },
  ];
}

// --- App and Routes Initialization ---
async function startServer() {
  const app = express();
  app.use(express.json());

  const engine = new ContainerEngine();
  const cli = new DockerCliInterpreter(engine);

  // Health
  app.get('/api/health', (_req: Request, res: Response) => {
    res.json({
      status: 'ok',
      runtime: 'Minijail Engine v1.0',
      kernel: 'Linux Namespaces + Cgroups v2 Simulation',
    });
  });

  // Primitives
  app.get('/api/primitives', (_req: Request, res: Response) => {
    res.json(getKernelPrimitivesData());
  });

  // CLI
  app.post('/api/cli', (req: Request, res: Response) => {
    const { command } = req.body || {};
    if (typeof command !== 'string') {
      res.status(400).json({ error: 'Invalid CLI command payload' });
      return;
    }
    const result = cli.execute(command);
    res.json(result);
  });

  // Containers List & Create
  app.get('/api/containers', (_req: Request, res: Response) => {
    res.json(engine.list());
  });

  app.post('/api/containers', (req: Request, res: Response) => {
    try {
      const created = engine.create(req.body);
      engine.start(created.id);
      res.status(201).json(created);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  // Container Detail & Actions
  app.get('/api/containers/:id', (req: Request, res: Response) => {
    const c = engine.get(req.params.id);
    if (!c) {
      res.status(404).json({ error: `no such container: ${req.params.id}` });
      return;
    }
    res.json(c);
  });

  app.delete('/api/containers/:id', (req: Request, res: Response) => {
    try {
      engine.remove(req.params.id);
      res.json({ status: 'deleted', id: req.params.id });
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/start', (req: Request, res: Response) => {
    try {
      const c = engine.start(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/stop', (req: Request, res: Response) => {
    try {
      const c = engine.stop(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/pause', (req: Request, res: Response) => {
    try {
      const c = engine.pause(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/unpause', (req: Request, res: Response) => {
    try {
      const c = engine.unpause(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/kill', (req: Request, res: Response) => {
    try {
      const c = engine.kill(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/exec', (req: Request, res: Response) => {
    try {
      const { command } = req.body || {};
      const result = engine.exec(req.params.id, command || '');
      res.json(result);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/stress-mem', (req: Request, res: Response) => {
    try {
      const deltaMb = req.body?.deltaMb || 24;
      const result = engine.stressMemory(req.params.id, deltaMb);
      res.json(result);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.post('/api/containers/:id/stress-cpu', (req: Request, res: Response) => {
    try {
      const c = engine.stressCpu(req.params.id);
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  app.patch('/api/containers/:id/cgroup', (req: Request, res: Response) => {
    try {
      const c = engine.updateCgroupLimits(req.params.id, req.body || {});
      res.json(c);
    } catch (e: any) {
      res.status(400).json({ error: e.message });
    }
  });

  // Catch-all for unmatched /api routes to prevent falling through to Vite SPA index.html
  app.all('/api/*', (req: Request, res: Response) => {
    res.status(404).json({ error: `API route not found: ${req.method} ${req.path}` });
  });

  // Dev vs Prod Vite Integration
  const isProd = process.env.NODE_ENV === 'production';
  if (!isProd) {
    const vite = await createViteServer({
      server: { middlewareMode: true },
      appType: 'spa',
    });
    app.use(vite.middlewares);
  } else {
    const distPath = path.resolve(__dirname, 'dist');
    if (fs.existsSync(distPath)) {
      app.use(express.static(distPath));
      app.get('*', (_req: Request, res: Response) => {
        res.sendFile(path.join(distPath, 'index.html'));
      });
    }
  }

  // Must run on port 3000 in AI Studio environment
  const PORT = 3000;
  app.listen(PORT, '0.0.0.0', () => {
    console.log(`[Minijail] Container Simulator running at http://localhost:${PORT}`);
  });
}

startServer().catch((err) => {
  console.error('[Minijail] Server startup error:', err);
  process.exit(1);
});
