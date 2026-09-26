<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue';
import type { Container, CreateContainerPayload } from '../../types/container.ts';
import { 
  fetchContainers, 
  executeCli, 
  createContainer, 
  startContainer, 
  stopContainer, 
  pauseContainer, 
  unpauseContainer, 
  killContainer, 
  deleteContainer,
  stressMemory 
} from '../../api.ts';
import { 
  Terminal, 
  Play, 
  Square, 
  Pause, 
  RotateCw, 
  Trash2, 
  Box, 
  Sliders, 
  Cpu, 
  HardDrive, 
  Send, 
  RefreshCw, 
  Search, 
  FileText, 
  Info, 
  Check, 
  Copy, 
  Shield,
  Layers,
  Zap,
} from 'lucide-vue-next';
import DockerLogo from '../common/DockerLogo.vue';

interface Props {
  initialContainers?: Container[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'refresh'): void;
  (e: 'inspect-container', container: Container): void;
  (e: 'open-metrics'): void;
  (e: 'open-kernel'): void;
}>();

// Containers state
const containers = ref<Container[]>([]);
const loadingList = ref(false);
const listFilter = ref<'all' | 'running' | 'paused' | 'stopped' | 'oom_killed'>('all');
const searchQuery = ref('');

// Run Form state
const runName = ref('');
const runImage = ref('alpine:3.19');
const runCommand = ref('sh -c "while true; do echo \\"[HTTP 200] GET /health - 2ms\\"; sleep 4; done"');
const runMemoryMb = ref(64);
const runCpuCores = ref(0.5);
const runPort = ref('8080:80/tcp');
const runningCreate = ref(false);

// Terminal state
interface TerminalEntry {
  id: string;
  command: string;
  output: string;
  error?: string;
  time: string;
  source: 'user' | 'panel' | 'system';
}

const terminalEntries = ref<TerminalEntry[]>([]);
const commandInput = ref('docker ps');
const isExecutingCli = ref(false);
const commandHistory = ref<string[]>([]);
const historyIndex = ref(-1);
const autoScroll = ref(true);
const terminalBody = ref<HTMLElement | null>(null);
const copiedOutput = ref(false);
const activeTabRight = ref<'list' | 'run' | 'bulk'>('list');

// Active selected container for quick focus
const focusedContainerId = ref<string | null>(null);

onMounted(async () => {
  await loadContainerList();
  
  // Print initial banner in terminal
  appendTerminalEntry({
    command: 'docker version',
    output: `Client: Docker Engine - Community
 Version:           26.1.0
 API version:       1.45
 Go version:        go1.22.4
 Git commit:        9714da6
 Built:             Wed Sep 17 07:11:00 2026
 OS/Arch:           linux/amd64
 Context:           default

Server: Docker Engine - Minijail Simulation
 Engine:
  Version:          1.0.0-minijail
  API version:      1.45 (minimum version 1.24)
  Kernel:           Linux 6.6.0-minijail (CLONE_NEWPID, CLONE_NEWNET, cgroups v2)
  Storage Driver:   overlayfs (UpperDir / LowerDir COW)`,
    source: 'system'
  });

  // Run initial docker ps in terminal
  await executeTerminalCommand('docker ps -a');
});

watch(() => props.initialContainers, (newVal) => {
  if (newVal && newVal.length > 0) {
    containers.value = newVal;
  }
}, { immediate: true });

// Load containers from API
async function loadContainerList() {
  loadingList.value = true;
  try {
    const data = await fetchContainers();
    containers.value = data;
    emit('refresh');
  } catch (err: any) {
    console.warn('Waiting for container engine...', err);
  } finally {
    loadingList.value = false;
  }
}

// Terminal scrolling
function scrollToBottom() {
  if (!autoScroll.value) return;
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight;
    }
  });
}

function appendTerminalEntry(entry: { command: string; output: string; error?: string; source?: 'user' | 'panel' | 'system' }) {
  const now = new Date().toTimeString().slice(0, 8);
  terminalEntries.value.push({
    id: `term-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    command: entry.command,
    output: entry.output,
    error: entry.error,
    time: now,
    source: entry.source || 'user'
  });
  scrollToBottom();
}

// Execute command in simulated terminal
async function executeTerminalCommand(cmdString?: string) {
  const cmd = (cmdString || commandInput.value).trim();
  if (!cmd || isExecutingCli.value) return;

  if (cmd === 'clear') {
    terminalEntries.value = [];
    commandInput.value = '';
    return;
  }

  commandHistory.value.push(cmd);
  historyIndex.value = commandHistory.value.length;
  commandInput.value = '';
  isExecutingCli.value = true;

  try {
    const res = await executeCli(cmd);
    appendTerminalEntry({
      command: cmd,
      output: res.output,
      error: res.error,
      source: 'user'
    });
    // Refresh container list after CLI execution
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({
      command: cmd,
      output: '',
      error: err.message || 'Execution failure',
      source: 'user'
    });
  } finally {
    isExecutingCli.value = false;
    scrollToBottom();
  }
}

// History navigation
function handleTerminalKeyDown(e: KeyboardEvent) {
  if (e.key === 'ArrowUp') {
    if (historyIndex.value > 0) {
      historyIndex.value--;
      commandInput.value = commandHistory.value[historyIndex.value];
    }
  } else if (e.key === 'ArrowDown') {
    if (historyIndex.value < commandHistory.value.length - 1) {
      historyIndex.value++;
      commandInput.value = commandHistory.value[historyIndex.value];
    } else {
      historyIndex.value = commandHistory.value.length;
      commandInput.value = '';
    }
  }
}

// Copy full terminal text
function copyTerminalLogs() {
  const text = terminalEntries.value
    .map(e => `[${e.time}] $ ${e.command}\n${e.output || ''}${e.error ? '\nError: ' + e.error : ''}`)
    .join('\n\n');
  navigator.clipboard.writeText(text);
  copiedOutput.value = true;
  setTimeout(() => { copiedOutput.value = false; }, 2000);
}

// Random container name helper
function generateRandomName() {
  const prefixes = ['gateway', 'worker', 'canary', 'redis-cache', 'api-service', 'web-proxy', 'queue-consumer'];
  const randPrefix = prefixes[Math.floor(Math.random() * prefixes.length)];
  const randHash = Math.random().toString(36).substring(2, 6);
  runName.value = `${randPrefix}-${randHash}`;
}

// Presets
function applyPreset(presetType: string) {
  if (presetType === 'web') {
    runName.value = 'micro-web';
    runImage.value = 'alpine:3.19';
    runCommand.value = 'sh -c "while true; do echo \\"[HTTP 200] GET /health - 2ms\\"; sleep 3; done"';
    runMemoryMb.value = 64;
    runCpuCores.value = 0.5;
    runPort.value = '8080:80/tcp';
  } else if (presetType === 'worker') {
    runName.value = 'queue-worker';
    runImage.value = 'busybox:1.36';
    runCommand.value = 'sh -c "while true; do echo \\"[Queue] Processed task batch\\"; sleep 5; done"';
    runMemoryMb.value = 128;
    runCpuCores.value = 1.0;
    runPort.value = '';
  } else if (presetType === 'canary') {
    runName.value = 'oom-canary';
    runImage.value = 'alpine:3.19';
    runCommand.value = 'sh -c "while true; do echo \\"[Canary] Polling telemetry\\"; sleep 3; done"';
    runMemoryMb.value = 32; // Low limit to demo OOM
    runCpuCores.value = 0.25;
    runPort.value = '';
  } else if (presetType === 'ubuntu') {
    runName.value = 'ubuntu-core';
    runImage.value = 'ubuntu:22.04';
    runCommand.value = 'sh -c "while true; do echo \\"[Systemd] Multi-user target active\\"; sleep 6; done"';
    runMemoryMb.value = 256;
    runCpuCores.value = 1.5;
    runPort.value = '3000:3000/tcp';
  }
}

// -------------------------------------------------------------
// CONTROL PANEL CONTAINER OPERATIONS (WITH TERMINAL REFLECTION)
// -------------------------------------------------------------

// Run Container via Control Panel
async function handleControlPanelRun() {
  const cName = runName.value.trim() || `docker-${Date.now().toString(36).slice(-4)}`;
  const img = runImage.value;
  const cmd = runCommand.value;
  const mem = runMemoryMb.value;
  const cpu = runCpuCores.value;
  const ports = runPort.value ? runPort.value.split(',').map(s => s.trim()).filter(Boolean) : [];

  const cliRepresentation = `docker run -d --name ${cName} -m ${mem}m --cpus ${cpu} ${ports.length ? '-p ' + ports[0] : ''} ${img} ${cmd}`;
  
  runningCreate.value = true;
  try {
    const payload: CreateContainerPayload = {
      name: cName,
      image: img,
      command: cmd,
      memoryLimitMb: mem,
      cpuQuotaCores: cpu,
      portBindings: ports,
      enablePidNs: true,
      enableUtsNs: true,
      enableNetNs: true,
      enableMntNs: true,
      hostname: cName,
      env: ['ENV=production', 'CONTAINER_RUNTIME=minijail']
    };

    const newContainer = await createContainer(payload);
    
    appendTerminalEntry({
      command: cliRepresentation,
      output: `[Minijail Engine] Unsharing Linux namespaces (CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNET, CLONE_NEWNS)
[Minijail Engine] Bound cgroups v2: /sys/fs/cgroup/minijail/${cName}/memory.max = ${mem * 1024 * 1024} B
[Minijail Engine] Mounting OverlayFS upper layer and pivot_root to /
${newContainer.id.slice(0, 12)}`,
      source: 'panel'
    });

    await loadContainerList();
    activeTabRight.value = 'list';
    focusedContainerId.value = newContainer.id;
  } catch (err: any) {
    appendTerminalEntry({
      command: cliRepresentation,
      output: '',
      error: err.message || 'Failed to run container',
      source: 'panel'
    });
  } finally {
    runningCreate.value = false;
  }
}

// Start Container
async function handleControlPanelStart(c: Container) {
  const shortId = c.id.slice(0, 12);
  const cmd = `docker start ${shortId}`;
  try {
    await startContainer(c.id);
    appendTerminalEntry({
      command: cmd,
      output: `${shortId}\nContainer ${c.name} transitioned to status: RUNNING`,
      source: 'panel'
    });
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({
      command: cmd,
      output: '',
      error: err.message,
      source: 'panel'
    });
  }
}

// Stop Container
async function handleControlPanelStop(c: Container) {
  const shortId = c.id.slice(0, 12);
  const cmd = `docker stop ${shortId}`;
  try {
    await stopContainer(c.id);
    appendTerminalEntry({
      command: cmd,
      output: `Sending SIGTERM to process PID 1 (${c.pid})...\n${shortId}\nContainer ${c.name} gracefully stopped.`,
      source: 'panel'
    });
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({
      command: cmd,
      output: '',
      error: err.message,
      source: 'panel'
    });
  }
}

// Pause / Unpause Container
async function handleControlPanelTogglePause(c: Container) {
  const shortId = c.id.slice(0, 12);
  if (c.status === 'paused') {
    const cmd = `docker unpause ${shortId}`;
    try {
      await unpauseContainer(c.id);
      appendTerminalEntry({
        command: cmd,
        output: `${shortId}\nContainer ${c.name} unpaused via SIGCONT (cgroups freezer resumed).`,
        source: 'panel'
      });
      await loadContainerList();
    } catch (err: any) {
      appendTerminalEntry({ command: cmd, output: '', error: err.message, source: 'panel' });
    }
  } else {
    const cmd = `docker pause ${shortId}`;
    try {
      await pauseContainer(c.id);
      appendTerminalEntry({
        command: cmd,
        output: `${shortId}\nContainer ${c.name} paused via SIGSTOP (cgroups freezer frozen).`,
        source: 'panel'
      });
      await loadContainerList();
    } catch (err: any) {
      appendTerminalEntry({ command: cmd, output: '', error: err.message, source: 'panel' });
    }
  }
}

// Kill Container
async function handleControlPanelKill(c: Container) {
  const shortId = c.id.slice(0, 12);
  const cmd = `docker kill ${shortId}`;
  try {
    await killContainer(c.id);
    appendTerminalEntry({
      command: cmd,
      output: `${shortId}\nSent SIGKILL to PID 1 (${c.pid}) in namespace ${c.namespaces.pid}.`,
      source: 'panel'
    });
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({ command: cmd, output: '', error: err.message, source: 'panel' });
  }
}

// Delete Container
async function handleControlPanelDelete(c: Container) {
  const shortId = c.id.slice(0, 12);
  const cmd = `docker rm ${shortId}`;
  try {
    await deleteContainer(c.id);
    appendTerminalEntry({
      command: cmd,
      output: `${shortId}\nUnmounted OverlayFS directory, cleaned up cgroups v2 path.`,
      source: 'panel'
    });
    if (focusedContainerId.value === c.id) {
      focusedContainerId.value = null;
    }
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({ command: cmd, output: '', error: err.message, source: 'panel' });
  }
}

// Inspect Container into Terminal
function handleControlPanelInspect(c: Container) {
  const shortId = c.id.slice(0, 12);
  const jsonStr = JSON.stringify({
    Id: c.id,
    Created: c.createdAt,
    State: {
      Status: c.status,
      Running: c.status === 'running',
      Paused: c.status === 'paused',
      OOMKilled: c.status === 'oom_killed',
      Pid: c.pid,
      ExitCode: c.exitCode
    },
    Image: c.image,
    Command: c.command,
    Namespaces: {
      PID: c.namespaces.pid,
      UTS: c.namespaces.uts,
      NET: c.namespaces.net,
      MNT: c.namespaces.mnt
    },
    Cgroups: {
      MemoryMaxBytes: c.cgroups.memoryMaxBytes,
      CpuQuotaUs: c.cgroups.cpuQuotaUs,
      CpuPeriodUs: c.cgroups.cpuPeriodUs,
      MemoryCurrentBytes: c.cgroups.memoryCurrentBytes
    },
    NetworkSettings: {
      IPAddress: c.ipAddress,
      PortBindings: c.portBindings
    },
    Mounts: {
      OverlayFS: c.overlayFs
    }
  }, null, 2);

  appendTerminalEntry({
    command: `docker inspect ${shortId}`,
    output: jsonStr,
    source: 'panel'
  });
}

// Output Container Logs into Terminal
function handleControlPanelLogs(c: Container) {
  const shortId = c.id.slice(0, 12);
  const logLines = c.logs.length > 0 
    ? c.logs.map(l => `[${new Date(l.timestamp).toLocaleTimeString()}] [${l.stream.toUpperCase()}] ${l.message}`).join('\n')
    : `(no standard output recorded for container ${shortId})`;

  appendTerminalEntry({
    command: `docker logs --tail 20 ${shortId}`,
    output: logLines,
    source: 'panel'
  });
}

// Stress Memory & Trigger Linux OOM
async function handleControlPanelStressOOM(c: Container) {
  const shortId = c.id.slice(0, 12);
  const cmd = `docker exec ${shortId} stress-mem --bytes +24MB`;
  try {
    const res = await stressMemory(c.id, 24);
    if (res.oomKilled) {
      appendTerminalEntry({
        command: cmd,
        output: `Allocating +24MB memory in process heap...
[KERNEL WARNING] Out of memory: Kill process ${c.pid} (${c.name}) score 982 or sacrifice child
[KERNEL EVENT] Killed process ${c.pid} (${c.name}) total-vm: 72000kB, anon-rss: 68100kB
Container ${shortId} has been OOM-Killed (status: oom_killed, exit code 137).`,
        error: `OOM Killer invoked: cgroups memory.max (${Math.round(c.cgroups.memoryMaxBytes / (1024*1024))}MB) exceeded!`,
        source: 'panel'
      });
    } else {
      appendTerminalEntry({
        command: cmd,
        output: `Allocated +24MB. Current memory: ${Math.round(res.container.cgroups.memoryCurrentBytes / (1024*1024))}MB / ${Math.round(res.container.cgroups.memoryMaxBytes / (1024*1024))}MB.`,
        source: 'panel'
      });
    }
    await loadContainerList();
  } catch (err: any) {
    appendTerminalEntry({ command: cmd, output: '', error: err.message, source: 'panel' });
  }
}

// Bulk Actions
async function handleStopAll() {
  const running = containers.value.filter(c => c.status === 'running');
  if (running.length === 0) {
    appendTerminalEntry({
      command: 'docker stop $(docker ps -q)',
      output: 'No active running containers to stop.',
      source: 'panel'
    });
    return;
  }
  appendTerminalEntry({
    command: 'docker stop $(docker ps -q)',
    output: `Stopping ${running.length} containers...`,
    source: 'panel'
  });
  for (const c of running) {
    await stopContainer(c.id).catch(() => {});
  }
  await loadContainerList();
  appendTerminalEntry({
    command: '# docker stop complete',
    output: `All ${running.length} containers stopped cleanly.`,
    source: 'panel'
  });
}

async function handleStartAll() {
  const stopped = containers.value.filter(c => c.status === 'stopped' || c.status === 'oom_killed');
  if (stopped.length === 0) {
    appendTerminalEntry({
      command: 'docker start $(docker ps -a -q -f status=exited)',
      output: 'No stopped containers found to start.',
      source: 'panel'
    });
    return;
  }
  appendTerminalEntry({
    command: 'docker start $(docker ps -a -q -f status=exited)',
    output: `Starting ${stopped.length} containers...`,
    source: 'panel'
  });
  for (const c of stopped) {
    await startContainer(c.id).catch(() => {});
  }
  await loadContainerList();
  appendTerminalEntry({
    command: '# docker start complete',
    output: `All stopped containers restarted.`,
    source: 'panel'
  });
}

async function handlePruneContainers() {
  const stopped = containers.value.filter(c => c.status === 'stopped' || c.status === 'oom_killed');
  appendTerminalEntry({
    command: 'docker container prune -f',
    output: `Deleted Containers:\n` + stopped.map(s => s.id.slice(0, 12)).join('\n') + `\nTotal reclaimed space: ${stopped.length * 14.2}MB`,
    source: 'panel'
  });
  for (const c of stopped) {
    await deleteContainer(c.id).catch(() => {});
  }
  await loadContainerList();
}

// Computed Filtered Containers
const filteredContainers = computed(() => {
  return containers.value.filter(c => {
    if (listFilter.value !== 'all' && c.status !== listFilter.value) {
      return false;
    }
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase();
      return (
        c.name.toLowerCase().includes(q) ||
        c.id.toLowerCase().includes(q) ||
        c.image.toLowerCase().includes(q)
      );
    }
    return true;
  });
});

const runningCount = computed(() => containers.value.filter(c => c.status === 'running').length);
const totalMemoryMB = computed(() => {
  const bytes = containers.value
    .filter(c => c.status === 'running')
    .reduce((sum, c) => sum + c.cgroups.memoryMaxBytes, 0);
  return Math.round(bytes / (1024 * 1024));
});
</script>

<template>
  <div id="container-workbench" class="space-y-4 select-none">
    <!-- Top Summary Banner -->
    <div class="bg-[#161B22] border border-[#232A35] rounded-xl p-4 flex flex-col md:flex-row md:items-center md:justify-between gap-3 shadow-md">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-[#1D63ED]/20 text-[#0db7ed] flex items-center justify-center border border-[#1D63ED]/30 shrink-0">
          <DockerLogo :size="28" />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h1 class="text-base font-bold text-white font-sans">Container Operations Workbench</h1>
            <span class="text-[10px] px-2 py-0.5 rounded-full bg-[#1D63ED]/20 text-[#0db7ed] border border-[#1D63ED]/30 font-mono font-semibold">
              Split Terminal & Control Panel
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-0.5">
            Simulated interactive terminal on the left coupled with real-time container lifecycle operations on the right.
          </p>
        </div>
      </div>

      <!-- Quick Metrics Header -->
      <div class="flex items-center gap-3 text-xs font-mono">
        <div class="px-3 py-1.5 rounded-lg bg-[#0E1217] border border-[#232A35] flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span class="text-slate-400">Running:</span>
          <span class="text-emerald-400 font-bold">{{ runningCount }}/{{ containers.length }}</span>
        </div>
        <div class="px-3 py-1.5 rounded-lg bg-[#0E1217] border border-[#232A35] flex items-center gap-2">
          <HardDrive class="w-3.5 h-3.5 text-cyan-400" />
          <span class="text-slate-400">Mem Limit:</span>
          <span class="text-cyan-300 font-bold">{{ totalMemoryMB }} MB</span>
        </div>
        <button 
          @click="emit('open-metrics')"
          class="px-2.5 py-1.5 rounded-lg bg-[#142339] hover:bg-[#1B3050] text-cyan-300 border border-cyan-800/60 flex items-center gap-1.5 transition-colors cursor-pointer text-xs font-mono font-medium"
          title="Open Real-Time Cgroups & Resource Monitor"
        >
          <Activity class="w-3.5 h-3.5 text-cyan-400" />
          <span>Monitor</span>
        </button>
        <button 
          @click="emit('open-kernel')"
          class="px-2.5 py-1.5 rounded-lg bg-[#1D1B28] hover:bg-[#2A263D] text-indigo-300 border border-indigo-800/60 flex items-center gap-1.5 transition-colors cursor-pointer text-xs font-mono font-medium"
          title="Open Interactive Linux Namespaces Isolation Diagram"
        >
          <Layers class="w-3.5 h-3.5 text-indigo-400" />
          <span>Namespaces</span>
        </button>
        <button 
          @click="loadContainerList"
          :disabled="loadingList"
          class="p-2 rounded-lg bg-[#1A222D] hover:bg-[#243040] text-slate-300 hover:text-white border border-[#2D3848] transition-colors cursor-pointer"
          title="Refresh Container States"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loadingList }" />
        </button>
      </div>
    </div>

    <!-- MAIN DUAL-PANE WORKSPACE: LEFT TERMINAL + RIGHT CONTROL PANEL -->
    <div class="grid grid-cols-1 xl:grid-cols-12 gap-5 items-start">
      
      <!-- ========================================================= -->
      <!-- LEFT PANE: SIMULATED TERMINAL AREA (xl:col-span-6) -->
      <!-- ========================================================= -->
      <div class="xl:col-span-6 flex flex-col h-[780px] bg-[#0A0D12] border border-[#232A35] rounded-xl overflow-hidden shadow-2xl">
        <!-- Terminal Window Titlebar -->
        <div class="px-4 py-3 bg-[#11161D] border-b border-[#232A35] flex items-center justify-between shrink-0">
          <div class="flex items-center gap-2.5">
            <!-- macOS / Linux dots -->
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded-full bg-[#E05252] border border-black/30"></div>
              <div class="w-3 h-3 rounded-full bg-[#E6A23C] border border-black/30"></div>
              <div class="w-3 h-3 rounded-full bg-[#3FB950] border border-black/30"></div>
            </div>
            <div class="text-xs font-mono text-slate-300 flex items-center gap-2 font-semibold ml-1">
              <Terminal class="w-3.5 h-3.5 text-[#0db7ed]" />
              <span>root@minijail:~#</span>
              <span class="text-[10px] text-slate-500 font-normal">/var/run/docker.sock</span>
            </div>
          </div>

          <!-- Terminal Action Buttons -->
          <div class="flex items-center gap-1.5 text-xs font-mono">
            <button 
              @click="copyTerminalLogs"
              class="px-2 py-1 rounded bg-[#161B22] hover:bg-[#202836] text-slate-400 hover:text-slate-200 border border-[#2B3545] transition-colors flex items-center gap-1 text-[11px] cursor-pointer"
              title="Copy all terminal output"
            >
              <Check v-if="copiedOutput" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3" />
              <span>{{ copiedOutput ? 'Copied' : 'Copy' }}</span>
            </button>
            <button 
              @click="terminalEntries = []"
              class="px-2 py-1 rounded bg-[#161B22] hover:bg-[#202836] text-slate-400 hover:text-rose-400 border border-[#2B3545] transition-colors text-[11px] cursor-pointer"
              title="Clear terminal buffer"
            >
              Clear
            </button>
            <button 
              @click="autoScroll = !autoScroll"
              class="px-2 py-1 rounded border transition-colors text-[11px] cursor-pointer"
              :class="autoScroll 
                ? 'bg-[#1D63ED]/20 text-[#0db7ed] border-[#1D63ED]/40' 
                : 'bg-[#161B22] text-slate-500 border-[#2B3545]'"
              title="Toggle auto-scroll on new output"
            >
              Auto-scroll
            </button>
          </div>
        </div>

        <!-- Terminal Quick Action Bar -->
        <div class="px-3 py-1.5 bg-[#0E1217] border-b border-[#232A35] flex items-center gap-1.5 overflow-x-auto text-[11px] font-mono shrink-0">
          <span class="text-slate-500 text-[10px] font-sans flex items-center gap-1 shrink-0">
            Quick CLI:
          </span>
          <button 
            @click="executeTerminalCommand('docker ps -a')"
            class="px-2 py-0.5 rounded bg-[#161B22] hover:bg-[#202836] text-slate-300 border border-[#2B3545] transition-colors shrink-0 cursor-pointer"
          >
            docker ps -a
          </button>
          <button 
            @click="executeTerminalCommand('docker stats')"
            class="px-2 py-0.5 rounded bg-[#161B22] hover:bg-[#202836] text-cyan-300 border border-[#2B3545] transition-colors shrink-0 cursor-pointer"
          >
            docker stats
          </button>
          <button 
            @click="executeTerminalCommand('docker images')"
            class="px-2 py-0.5 rounded bg-[#161B22] hover:bg-[#202836] text-slate-300 border border-[#2B3545] transition-colors shrink-0 cursor-pointer"
          >
            docker images
          </button>
          <button 
            @click="executeTerminalCommand('docker version')"
            class="px-2 py-0.5 rounded bg-[#161B22] hover:bg-[#202836] text-indigo-300 border border-[#2B3545] transition-colors shrink-0 cursor-pointer"
          >
            docker version
          </button>
          <button 
            @click="executeTerminalCommand('docker --help')"
            class="px-2 py-0.5 rounded bg-[#161B22] hover:bg-[#202836] text-slate-400 border border-[#2B3545] transition-colors shrink-0 cursor-pointer"
          >
            --help
          </button>
        </div>

        <!-- Terminal Output Stream Body -->
        <div 
          ref="terminalBody"
          class="flex-1 p-4 overflow-y-auto font-mono text-xs space-y-3.5 selection:bg-[#1D63ED]/30 text-slate-200 leading-relaxed"
        >
          <div v-for="item in terminalEntries" :key="item.id" class="space-y-1">
            <!-- Command line line -->
            <div class="flex items-center gap-2 text-slate-400">
              <span class="text-slate-600 text-[10px] select-none">[{{ item.time }}]</span>
              <span class="text-[#0db7ed] font-bold select-none">$</span>
              <span class="text-white font-semibold">{{ item.command }}</span>
              <span 
                v-if="item.source === 'panel'"
                class="text-[9px] px-1 py-0.2 rounded bg-indigo-950 text-indigo-300 border border-indigo-800/60 ml-auto select-none"
              >
                Panel Action
              </span>
            </div>

            <!-- Output Block -->
            <div 
              v-if="item.output" 
              class="bg-[#11161D] p-3 rounded-lg border border-[#232A35] text-slate-200 whitespace-pre font-mono text-[11px] overflow-x-auto"
            >{{ item.output }}</div>

            <!-- Error Block -->
            <div 
              v-if="item.error" 
              class="bg-rose-950/40 p-2.5 rounded-lg border border-rose-800/60 text-rose-300 whitespace-pre-wrap font-mono text-[11px]"
            >Error: {{ item.error }}</div>
          </div>

          <!-- Loading state -->
          <div v-if="isExecutingCli" class="flex items-center gap-2 text-[#0db7ed] pt-2 text-[11px]">
            <span class="w-1.5 h-1.5 rounded-full bg-[#0db7ed] animate-ping"></span>
            <span>Minijail engine evaluating system call...</span>
          </div>
        </div>

        <!-- Terminal Interactive Input -->
        <div class="p-3 bg-[#11161D] border-t border-[#232A35] shrink-0">
          <form @submit.prevent="executeTerminalCommand()" class="flex items-center gap-2 font-mono text-xs">
            <span class="text-[#0db7ed] font-bold text-sm shrink-0 select-none">$</span>
            <input 
              v-model="commandInput"
              @keydown="handleTerminalKeyDown"
              type="text" 
              placeholder="Type docker command (e.g. docker ps, docker run -d --name app alpine:3.19 sh)..."
              class="flex-1 bg-transparent text-slate-100 placeholder:text-slate-600 focus:outline-none text-xs"
              :disabled="isExecutingCli"
            />
            <button 
              type="submit" 
              :disabled="isExecutingCli || !commandInput.trim()"
              class="px-3.5 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-30 text-white font-medium flex items-center gap-1 transition-colors cursor-pointer text-xs"
            >
              <Send class="w-3.5 h-3.5" />
              <span>Execute</span>
            </button>
          </form>
        </div>
      </div>

      <!-- ========================================================= -->
      <!-- RIGHT PANE: CONTAINER OPERATIONS CONTROL PANEL (xl:col-span-6) -->
      <!-- ========================================================= -->
      <div class="xl:col-span-6 flex flex-col h-[780px] bg-[#161B22] border border-[#232A35] rounded-xl overflow-hidden shadow-2xl">
        <!-- Control Panel Header & Tabs -->
        <div class="px-5 py-3.5 bg-[#11161D] border-b border-[#232A35] flex items-center justify-between shrink-0">
          <div class="flex items-center gap-2">
            <Sliders class="w-4 h-4 text-[#0db7ed]" />
            <h2 class="text-sm font-bold text-white font-sans">Container Operations</h2>
          </div>

          <!-- Section Switcher Tabs -->
          <div class="flex items-center bg-[#0E1217] p-1 rounded-lg border border-[#232A35] text-xs font-mono">
            <button
              @click="activeTabRight = 'list'"
              class="px-3 py-1 rounded-md transition-colors cursor-pointer"
              :class="activeTabRight === 'list' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-slate-200'"
            >
              List & Manage ({{ containers.length }})
            </button>
            <button
              @click="activeTabRight = 'run'"
              class="px-3 py-1 rounded-md transition-colors cursor-pointer flex items-center gap-1"
              :class="activeTabRight === 'run' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-slate-200'"
            >
              <Play class="w-3 h-3" />
              <span>Run New</span>
            </button>
            <button
              @click="activeTabRight = 'bulk'"
              class="px-3 py-1 rounded-md transition-colors cursor-pointer"
              :class="activeTabRight === 'bulk' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-slate-200'"
            >
              Bulk & Ops
            </button>
          </div>
        </div>

        <!-- -------------------------------------------------------- -->
        <!-- TAB 1: LIST & MANAGE CONTAINERS (LIVE docker ps) -->
        <!-- -------------------------------------------------------- -->
        <div v-if="activeTabRight === 'list'" class="flex-1 flex flex-col overflow-hidden">
          <!-- Search & Filter Ribbon -->
          <div class="p-3 bg-[#0E1217] border-b border-[#232A35] flex items-center gap-2 shrink-0">
            <!-- Search -->
            <div class="relative flex-1">
              <Search class="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input 
                v-model="searchQuery"
                type="text" 
                placeholder="Filter by name, ID, or image..."
                class="w-full pl-8 pr-3 py-1.5 rounded-lg bg-[#161B22] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
              />
            </div>

            <!-- Status Filter Dropdown -->
            <select 
              v-model="listFilter"
              class="px-2.5 py-1.5 rounded-lg bg-[#161B22] border border-[#2B3545] text-slate-200 text-xs focus:outline-none focus:border-[#1D63ED] font-mono cursor-pointer"
            >
              <option value="all">All States ({{ containers.length }})</option>
              <option value="running">Running ({{ containers.filter(c => c.status === 'running').length }})</option>
              <option value="paused">Paused ({{ containers.filter(c => c.status === 'paused').length }})</option>
              <option value="stopped">Stopped ({{ containers.filter(c => c.status === 'stopped').length }})</option>
              <option value="oom_killed">OOM Killed ({{ containers.filter(c => c.status === 'oom_killed').length }})</option>
            </select>
          </div>

          <!-- Containers List Scrollable -->
          <div class="flex-1 overflow-y-auto p-4 space-y-3">
            <div 
              v-for="c in filteredContainers" 
              :key="c.id"
              class="p-3.5 rounded-xl border transition-all font-mono text-xs"
              :class="focusedContainerId === c.id 
                ? 'bg-[#1D63ED]/10 border-[#1D63ED] shadow-md' 
                : 'bg-[#11161D] border-[#232A35] hover:border-[#2D3848]'"
            >
              <!-- Container Card Header -->
              <div class="flex items-start justify-between gap-2">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-white text-sm font-sans">{{ c.name }}</span>
                    <!-- Status Badge -->
                    <span 
                      class="text-[10px] px-2 py-0.5 rounded-full font-bold flex items-center gap-1.5"
                      :class="{
                        'bg-emerald-950/80 text-emerald-400 border border-emerald-800/80': c.status === 'running',
                        'bg-amber-950/80 text-amber-400 border border-amber-800/80': c.status === 'paused',
                        'bg-rose-950/80 text-rose-300 border border-rose-800/80': c.status === 'oom_killed',
                        'bg-[#1E2633] text-slate-400 border border-[#2B3545]': c.status === 'stopped'
                      }"
                    >
                      <span 
                        class="w-1.5 h-1.5 rounded-full"
                        :class="{
                          'bg-emerald-400 animate-pulse': c.status === 'running',
                          'bg-amber-400': c.status === 'paused',
                          'bg-rose-400': c.status === 'oom_killed',
                          'bg-slate-500': c.status === 'stopped'
                        }"
                      ></span>
                      <span>{{ c.status.toUpperCase() }}</span>
                    </span>
                  </div>

                  <div class="text-[11px] text-slate-400 flex items-center gap-2">
                    <span class="text-cyan-400 font-semibold">{{ c.image }}</span>
                    <span class="text-slate-600">•</span>
                    <span class="text-slate-500">{{ c.id.slice(0, 12) }}</span>
                    <span class="text-slate-600">•</span>
                    <span class="text-slate-400">PID {{ c.pid }}</span>
                  </div>
                </div>

                <!-- Resource Metrics Gauge -->
                <div class="text-right text-[11px] space-y-0.5 shrink-0">
                  <div class="text-slate-300">
                    <span class="text-slate-500">Mem:</span> 
                    <span class="text-cyan-300 font-bold ml-1">
                      {{ Math.round(c.cgroups.memoryCurrentBytes / (1024*1024)) }}MB / {{ Math.round(c.cgroups.memoryMaxBytes / (1024*1024)) }}MB
                    </span>
                  </div>
                  <div class="text-slate-400">
                    <span class="text-slate-500">CPU:</span> 
                    <span class="text-indigo-300 font-bold ml-1">
                      {{ (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs).toFixed(2) }} cores
                    </span>
                  </div>
                </div>
              </div>

              <!-- Action Bar for this container -->
              <div class="mt-3 pt-2.5 border-t border-[#232A35] flex items-center justify-between flex-wrap gap-1.5">
                <!-- Left: State Transitions -->
                <div class="flex items-center gap-1.5">
                  <!-- Start -->
                  <button 
                    v-if="c.status === 'stopped' || c.status === 'oom_killed'"
                    @click="handleControlPanelStart(c)"
                    class="px-2.5 py-1 rounded bg-emerald-950/70 hover:bg-emerald-900 text-emerald-300 border border-emerald-800 transition-colors flex items-center gap-1 text-[11px] font-semibold cursor-pointer"
                    title="docker start"
                  >
                    <Play class="w-3 h-3 fill-current" />
                    <span>Start</span>
                  </button>

                  <!-- Stop -->
                  <button 
                    v-if="c.status === 'running' || c.status === 'paused'"
                    @click="handleControlPanelStop(c)"
                    class="px-2.5 py-1 rounded bg-[#1E2633] hover:bg-[#2B3545] text-slate-200 border border-[#3B4758] transition-colors flex items-center gap-1 text-[11px] font-semibold cursor-pointer"
                    title="docker stop"
                  >
                    <Square class="w-3 h-3 fill-current" />
                    <span>Stop</span>
                  </button>

                  <!-- Pause / Resume -->
                  <button 
                    v-if="c.status === 'running' || c.status === 'paused'"
                    @click="handleControlPanelTogglePause(c)"
                    class="px-2.5 py-1 rounded bg-amber-950/70 hover:bg-amber-900 text-amber-300 border border-amber-800 transition-colors flex items-center gap-1 text-[11px] cursor-pointer"
                    :title="c.status === 'paused' ? 'docker unpause' : 'docker pause'"
                  >
                    <Pause class="w-3 h-3" />
                    <span>{{ c.status === 'paused' ? 'Unpause' : 'Pause' }}</span>
                  </button>

                  <!-- Kill (SIGKILL) -->
                  <button 
                    v-if="c.status === 'running' || c.status === 'paused'"
                    @click="handleControlPanelKill(c)"
                    class="px-2 py-1 rounded bg-rose-950/60 hover:bg-rose-900 text-rose-300 border border-rose-800/80 transition-colors flex items-center gap-1 text-[11px] cursor-pointer"
                    title="docker kill (SIGKILL)"
                  >
                    <Zap class="w-3 h-3" />
                    <span>Kill</span>
                  </button>
                </div>

                <!-- Right: Diagnostics & Inspect Output -->
                <div class="flex items-center gap-1.5 ml-auto">
                  <!-- Stress Memory (OOM Trigger) -->
                  <button 
                    v-if="c.status === 'running'"
                    @click="handleControlPanelStressOOM(c)"
                    class="px-2 py-1 rounded bg-purple-950/60 hover:bg-purple-900 text-purple-300 border border-purple-800/70 transition-colors text-[10px] cursor-pointer"
                    title="Simulate memory pressure to trigger Linux OOM-killer"
                  >
                    +24M (OOM Stress)
                  </button>

                  <!-- Logs -->
                  <button 
                    @click="handleControlPanelLogs(c)"
                    class="px-2 py-1 rounded bg-[#161B22] hover:bg-[#202836] text-slate-300 hover:text-white border border-[#2B3545] transition-colors flex items-center gap-1 text-[10px] cursor-pointer"
                    title="Output logs into left terminal"
                  >
                    <FileText class="w-3 h-3" />
                    <span>Logs</span>
                  </button>

                  <!-- Inspect -->
                  <button 
                    @click="handleControlPanelInspect(c)"
                    class="px-2 py-1 rounded bg-[#161B22] hover:bg-[#202836] text-slate-300 hover:text-white border border-[#2B3545] transition-colors flex items-center gap-1 text-[10px] cursor-pointer"
                    title="Output inspect JSON into left terminal"
                  >
                    <Info class="w-3 h-3" />
                    <span>Inspect</span>
                  </button>

                  <!-- Delete -->
                  <button 
                    @click="handleControlPanelDelete(c)"
                    class="p-1 rounded bg-[#161B22] hover:bg-rose-950 text-slate-400 hover:text-rose-400 border border-[#2B3545] hover:border-rose-800 transition-colors cursor-pointer"
                    title="docker rm"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Empty List State -->
            <div v-if="filteredContainers.length === 0" class="text-center py-12 space-y-3">
              <Box class="w-10 h-10 mx-auto text-slate-600" />
              <p class="text-xs text-slate-400 font-sans">No containers match the current filter.</p>
              <button 
                @click="activeTabRight = 'run'"
                class="px-3.5 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] text-white text-xs font-semibold cursor-pointer"
              >
                + Run a Container
              </button>
            </div>
          </div>
        </div>

        <!-- -------------------------------------------------------- -->
        <!-- TAB 2: RUN NEW CONTAINER (docker run form) -->
        <!-- -------------------------------------------------------- -->
        <div v-else-if="activeTabRight === 'run'" class="flex-1 overflow-y-auto p-5 space-y-4 font-mono text-xs">
          <!-- Quick Presets -->
          <div class="p-3 rounded-xl bg-[#0E1217] border border-[#232A35] space-y-2">
            <span class="text-[11px] text-slate-400 flex items-center gap-1 font-semibold font-sans">
              One-Click Presets:
            </span>
            <div class="grid grid-cols-2 gap-2">
              <button 
                type="button" 
                @click="applyPreset('web')"
                class="px-2.5 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202836] text-cyan-300 border border-[#2B3545] text-left transition-colors cursor-pointer"
              >
                <div class="font-bold">Alpine Web</div>
                <div class="text-[10px] text-slate-500">64MB • 0.5 CPU</div>
              </button>
              <button 
                type="button" 
                @click="applyPreset('canary')"
                class="px-2.5 py-1.5 rounded-lg bg-rose-950/40 hover:bg-rose-900/60 text-rose-300 border border-rose-800/60 text-left transition-colors cursor-pointer"
              >
                <div class="font-bold">OOM Canary</div>
                <div class="text-[10px] text-rose-400/80">32MB (Breach Demo)</div>
              </button>
              <button 
                type="button" 
                @click="applyPreset('worker')"
                class="px-2.5 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202836] text-slate-300 border border-[#2B3545] text-left transition-colors cursor-pointer"
              >
                <div class="font-bold">Busybox Worker</div>
                <div class="text-[10px] text-slate-500">128MB • 1.0 CPU</div>
              </button>
              <button 
                type="button" 
                @click="applyPreset('ubuntu')"
                class="px-2.5 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202836] text-slate-300 border border-[#2B3545] text-left transition-colors cursor-pointer"
              >
                <div class="font-bold">Ubuntu Service</div>
                <div class="text-[10px] text-slate-500">256MB • 1.5 CPU</div>
              </button>
            </div>
          </div>

          <!-- Name & Image -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="text-slate-300 font-medium">Container Name</label>
                <button 
                  type="button" 
                  @click="generateRandomName"
                  class="text-[10px] text-[#0db7ed] hover:underline cursor-pointer"
                >
                  Generate
                </button>
              </div>
              <input 
                v-model="runName"
                type="text" 
                placeholder="e.g. web-gateway-01"
                class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
              />
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Image Tag</label>
              <select 
                v-model="runImage"
                class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
              >
                <option value="alpine:3.19">alpine:3.19 (7.38 MB)</option>
                <option value="busybox:1.36">busybox:1.36 (4.26 MB)</option>
                <option value="ubuntu:22.04">ubuntu:22.04 (77.8 MB)</option>
                <option value="nginx:alpine">nginx:alpine (23.4 MB)</option>
              </select>
            </div>
          </div>

          <!-- Command -->
          <div>
            <label class="block text-slate-300 font-medium mb-1">Entrypoint / Command</label>
            <input 
              v-model="runCommand"
              type="text" 
              placeholder="sh -c 'echo Started; sleep 3600'"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            />
          </div>

          <!-- Cgroups v2 Constraints -->
          <div class="p-3.5 rounded-xl bg-[#0E1217] border border-[#232A35] space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-cyan-400 font-bold flex items-center gap-1.5">
                <HardDrive class="w-3.5 h-3.5" />
                <span>Memory Limit (memory.max)</span>
              </span>
              <span class="text-white font-bold">{{ runMemoryMb }} MB</span>
            </div>
            <input 
              v-model.number="runMemoryMb"
              type="range" 
              min="16" 
              max="512" 
              step="16"
              class="w-full accent-[#1D63ED] cursor-pointer"
            />
            <div class="flex justify-between text-[10px] text-slate-500">
              <span>16 MB</span>
              <span>128 MB</span>
              <span>512 MB</span>
            </div>

            <div class="flex items-center justify-between pt-2 border-t border-[#232A35]">
              <span class="text-indigo-400 font-bold flex items-center gap-1.5">
                <Cpu class="w-3.5 h-3.5" />
                <span>CPU CFS Quota (cpu.max)</span>
              </span>
              <span class="text-white font-bold">{{ runCpuCores.toFixed(2) }} Cores</span>
            </div>
            <input 
              v-model.number="runCpuCores"
              type="range" 
              min="0.1" 
              max="2.0" 
              step="0.1"
              class="w-full accent-indigo-500 cursor-pointer"
            />
          </div>

          <!-- Port Mapping -->
          <div>
            <label class="block text-slate-300 font-medium mb-1">Port Mapping (-p)</label>
            <input 
              v-model="runPort"
              type="text" 
              placeholder="8080:80/tcp"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            />
          </div>

          <!-- Submit Run Button -->
          <button 
            type="button" 
            @click="handleControlPanelRun"
            :disabled="runningCreate"
            class="w-full py-2.5 rounded-xl bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-50 text-white font-bold flex items-center justify-center gap-2 shadow-lg transition-all cursor-pointer text-xs"
          >
            <Play class="w-4 h-4 fill-current" />
            <span>{{ runningCreate ? 'Deploying Container...' : 'Execute: docker run' }}</span>
          </button>
        </div>

        <!-- -------------------------------------------------------- -->
        <!-- TAB 3: BULK & GLOBAL OPERATIONS -->
        <!-- -------------------------------------------------------- -->
        <div v-else-if="activeTabRight === 'bulk'" class="flex-1 overflow-y-auto p-5 space-y-4 font-mono text-xs">
          <div class="p-4 rounded-xl bg-[#0E1217] border border-[#232A35] space-y-3">
            <h3 class="text-white font-bold font-sans flex items-center gap-2">
              <Zap class="w-4 h-4 text-[#0db7ed]" />
              <span>Bulk Engine Orchestration</span>
            </h3>
            <p class="text-slate-400 text-xs font-sans">
              Perform fleet-wide operations across all simulated Linux processes in Minijail.
            </p>

            <div class="space-y-2 pt-2">
              <button 
                type="button" 
                @click="handleStopAll"
                class="w-full p-2.5 rounded-lg bg-[#161B22] hover:bg-[#202836] border border-[#2B3545] text-slate-200 hover:text-white flex items-center justify-between transition-colors cursor-pointer"
              >
                <div class="flex items-center gap-2">
                  <Square class="w-3.5 h-3.5 text-amber-400" />
                  <span>Stop All Running Containers</span>
                </div>
                <span class="text-[10px] text-slate-500 font-mono">docker stop $(docker ps -q)</span>
              </button>

              <button 
                type="button" 
                @click="handleStartAll"
                class="w-full p-2.5 rounded-lg bg-[#161B22] hover:bg-[#202836] border border-[#2B3545] text-slate-200 hover:text-white flex items-center justify-between transition-colors cursor-pointer"
              >
                <div class="flex items-center gap-2">
                  <Play class="w-3.5 h-3.5 text-emerald-400" />
                  <span>Start All Stopped Containers</span>
                </div>
                <span class="text-[10px] text-slate-500 font-mono">docker start ...</span>
              </button>

              <button 
                type="button" 
                @click="handlePruneContainers"
                class="w-full p-2.5 rounded-lg bg-rose-950/30 hover:bg-rose-950/60 border border-rose-800/50 text-rose-200 flex items-center justify-between transition-colors cursor-pointer"
              >
                <div class="flex items-center gap-2">
                  <Trash2 class="w-3.5 h-3.5 text-rose-400" />
                  <span>Prune Exited Containers</span>
                </div>
                <span class="text-[10px] text-rose-400/70 font-mono">docker container prune</span>
              </button>
            </div>
          </div>

          <!-- Kernel Primitives Quick Summary -->
          <div class="p-4 rounded-xl bg-[#0E1217] border border-[#232A35] space-y-2">
            <h4 class="text-white font-bold font-sans flex items-center gap-2">
              <Shield class="w-4 h-4 text-emerald-400" />
              <span>Active Kernel Primitives</span>
            </h4>
            <div class="grid grid-cols-2 gap-2 text-[11px] text-slate-400">
              <div class="p-2 rounded bg-[#161B22] border border-[#2B3545]">
                <div class="text-cyan-400 font-bold">CLONE_NEWPID</div>
                <div class="text-[10px] text-slate-500">Isolated PID 1 tree</div>
              </div>
              <div class="p-2 rounded bg-[#161B22] border border-[#2B3545]">
                <div class="text-cyan-400 font-bold">CLONE_NEWNET</div>
                <div class="text-[10px] text-slate-500">Virtual veth interface</div>
              </div>
              <div class="p-2 rounded bg-[#161B22] border border-[#2B3545]">
                <div class="text-emerald-400 font-bold">memory.max</div>
                <div class="text-[10px] text-slate-500">OOM Killer threshold</div>
              </div>
              <div class="p-2 rounded bg-[#161B22] border border-[#2B3545]">
                <div class="text-indigo-400 font-bold">cpu.max</div>
                <div class="text-[10px] text-slate-500">CFS bandwidth quota</div>
              </div>
            </div>
          </div>
        </div>

      </div>

    </div>
  </div>
</template>
