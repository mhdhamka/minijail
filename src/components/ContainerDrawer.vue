<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue';
import type { Container } from '../types/container';
import { 
  execCommand, 
  stressMemory, 
  stressCpu, 
  startContainer, 
  stopContainer, 
  pauseContainer, 
  unpauseContainer, 
  killContainer 
} from '../api';
import { 
  X, 
  Play, 
  Square, 
  Pause, 
  Flame, 
  Terminal as TerminalIcon, 
  FileText, 
  Code, 
  FolderTree, 
  Activity, 
  HardDrive, 
  Cpu, 
  Skull, 
  Send, 
  Plus, 
  Sparkles, 
  ExternalLink,
  ShieldCheck,
  Search,
  Copy,
  Check,
  RefreshCw,
  AlertTriangle
} from 'lucide-vue-next';

interface Props {
  container: Container;
  initialTab?: 'logs' | 'exec' | 'inspect' | 'files' | 'stats';
}

const props = withDefaults(defineProps<Props>(), {
  initialTab: 'logs'
});

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
}>();

const activeTab = ref<'logs' | 'exec' | 'inspect' | 'files' | 'stats'>(props.initialTab);

// Logs State
const logFilter = ref('');
const logsContainerRef = ref<HTMLElement | null>(null);

// Exec / Terminal State
const inputCommand = ref('');
const execLoading = ref(false);
const terminalLines = ref<{ stream: 'stdout' | 'stderr' | 'stdin' | 'system'; text: string; time: string }[]>([]);
const terminalBody = ref<HTMLElement | null>(null);

// Files State
const newFileName = ref('');
const isCreatingFile = ref(false);

// Stats & Cgroups State
const statsLoading = ref(false);
const statsAlert = ref<string | null>(null);

// Inspect State
const copiedInspect = ref(false);

onMounted(() => {
  // Load initial logs
  props.container.logs.forEach(log => {
    terminalLines.value.push({
      stream: log.stream,
      text: log.message,
      time: log.timestamp,
    });
  });
  scrollLogsToBottom();
});

function scrollLogsToBottom() {
  nextTick(() => {
    if (logsContainerRef.value) {
      logsContainerRef.value.scrollTop = logsContainerRef.value.scrollHeight;
    }
  });
}

function scrollTerminalToBottom() {
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight;
    }
  });
}

// Filtered logs
const filteredLogs = computed(() => {
  if (!logFilter.value) return props.container.logs;
  const q = logFilter.value.toLowerCase();
  return props.container.logs.filter(l => l.message.toLowerCase().includes(q));
});

// Memory & CPU calculations
const memUsedMB = computed(() => {
  return (props.container.cgroups.memoryUsageBytes / (1024 * 1024)).toFixed(1);
});

const memLimitMB = computed(() => {
  return (props.container.cgroups.memoryMaxBytes / (1024 * 1024)).toFixed(0);
});

const memPercent = computed(() => {
  if (props.container.cgroups.memoryMaxBytes === 0) return 0;
  const pct = (props.container.cgroups.memoryUsageBytes / props.container.cgroups.memoryMaxBytes) * 100;
  return Math.min(Math.round(pct), 100);
});

const cpuQuotaCores = computed(() => {
  if (props.container.cgroups.cpuPeriodUs === 0) return '1.00';
  return (props.container.cgroups.cpuQuotaUs / props.container.cgroups.cpuPeriodUs).toFixed(2);
});

// Exec command
async function runCommand(cmd?: string) {
  const target = (cmd || inputCommand.value).trim();
  if (!target || execLoading.value) return;

  const now = new Date().toTimeString().slice(0, 8);
  terminalLines.value.push({
    stream: 'stdin',
    text: `$ ${target}`,
    time: now,
  });

  inputCommand.value = '';
  execLoading.value = true;
  scrollTerminalToBottom();

  try {
    const res = await execCommand(props.container.id, target);
    if (res.stdout) {
      terminalLines.value.push({
        stream: 'stdout',
        text: res.stdout.trimEnd(),
        time: now,
      });
    }
    if (res.stderr) {
      terminalLines.value.push({
        stream: 'stderr',
        text: res.stderr.trimEnd(),
        time: now,
      });
    }
    emit('refresh');
  } catch (err: any) {
    terminalLines.value.push({
      stream: 'stderr',
      text: `exec error: ${err.message}`,
      time: now,
    });
  } finally {
    execLoading.value = false;
    scrollTerminalToBottom();
  }
}

// File creation (COW)
async function handleCreateFile() {
  if (!newFileName.value.trim() || isCreatingFile.value) return;
  const path = newFileName.value.startsWith('/') ? newFileName.value : `/${newFileName.value}`;
  isCreatingFile.value = true;
  try {
    await execCommand(props.container.id, `touch ${path}`);
    newFileName.value = '';
    emit('refresh');
  } catch (err) {
    console.error(err);
  } finally {
    isCreatingFile.value = false;
  }
}

// Cgroups stress actions
async function handleStressMem(deltaMb: number) {
  statsLoading.value = true;
  statsAlert.value = null;
  try {
    const res = await stressMemory(props.container.id, deltaMb);
    emit('refresh');
    if (res.oomKilled) {
      statsAlert.value = `[OOM-KILLER TRIGGERED] Container breached ${memLimitMB.value}MB memory.max! Kernel sent SIGKILL (Exit code 137).`;
    } else {
      statsAlert.value = `Allocated +${deltaMb}MB. Memory: ${(res.container.cgroups.memoryUsageBytes / (1024*1024)).toFixed(1)}MB / ${memLimitMB.value}MB.`;
    }
  } catch (err: any) {
    statsAlert.value = `Error: ${err.message}`;
  } finally {
    statsLoading.value = false;
  }
}

async function handleForceOom() {
  const remainingBytes = props.container.cgroups.memoryMaxBytes - props.container.cgroups.memoryUsageBytes;
  const deltaMb = Math.max(Math.ceil(remainingBytes / (1024 * 1024)) + 12, 32);
  await handleStressMem(deltaMb);
}

async function handleStressCpu() {
  statsLoading.value = true;
  statsAlert.value = null;
  try {
    await stressCpu(props.container.id);
    emit('refresh');
    statsAlert.value = `CFS scheduler throttled process! Quota exceeded.`;
  } catch (err: any) {
    statsAlert.value = `Error: ${err.message}`;
  } finally {
    statsLoading.value = false;
  }
}

// Copy JSON Inspect
function copyInspect() {
  const inspectData = {
    Id: props.container.id,
    Created: props.container.createdAt,
    Path: props.container.command,
    State: {
      Status: props.container.status,
      Running: props.container.status === 'running',
      Paused: props.container.status === 'paused',
      OOMKilled: props.container.status === 'oom_killed',
      Pid: props.container.hostPid,
      ExitCode: props.container.status === 'oom_killed' ? 137 : 0,
      StartedAt: props.container.startedAt,
    },
    Image: props.container.image,
    HostConfig: {
      Memory: props.container.cgroups.memoryMaxBytes,
      CpuQuota: props.container.cgroups.cpuQuotaUs,
      CpuPeriod: props.container.cgroups.cpuPeriodUs,
      CgroupPath: props.container.cgroups.cgroupPath,
    },
    NetworkSettings: {
      IPAddress: props.container.namespaces.virtualIp,
      Ports: props.container.portBindings,
      VethPair: props.container.namespaces.vethPair,
    },
    Mounts: [
      {
        Type: "overlay",
        Source: props.container.rootfs.lowerDir,
        Destination: "/",
        Mode: "rw",
        RW: true,
        Propagation: "rprivate"
      }
    ]
  };
  navigator.clipboard.writeText(JSON.stringify(inspectData, null, 2));
  copiedInspect.value = true;
  setTimeout(() => { copiedInspect.value = false; }, 2000);
}

// Lifecycle buttons
async function handleAction(action: 'start' | 'stop' | 'pause' | 'unpause' | 'kill') {
  try {
    if (action === 'start') await startContainer(props.container.id);
    if (action === 'stop') await stopContainer(props.container.id);
    if (action === 'pause') await pauseContainer(props.container.id);
    if (action === 'unpause') await unpauseContainer(props.container.id);
    if (action === 'kill') await killContainer(props.container.id);
    emit('refresh');
  } catch (e) {
    console.error(e);
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/75 backdrop-blur-sm select-none">
    <div 
      id="docker-container-drawer"
      class="bg-[#161B22] border border-[#2D3848] rounded-xl w-full max-w-5xl h-[88vh] flex flex-col shadow-2xl overflow-hidden text-slate-200"
    >
      <!-- Top Title Bar (Authentic Docker Desktop Container Header) -->
      <div class="px-5 py-3.5 bg-[#11161D] border-b border-[#232A35] flex items-center justify-between">
        <div class="flex items-center gap-3">
          <!-- Status icon dot -->
          <div class="relative flex items-center justify-center">
            <span 
              class="w-3 h-3 rounded-full" 
              :class="[
                container.status === 'running' ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)] animate-pulse' :
                container.status === 'paused' ? 'bg-amber-500' :
                container.status === 'oom_killed' ? 'bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.8)] animate-pulse' :
                'bg-slate-500'
              ]"
            ></span>
          </div>

          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-base font-bold text-white font-mono">{{ container.name }}</h2>
              <span 
                class="text-[11px] font-mono px-2 py-0.5 rounded font-bold uppercase"
                :class="[
                  container.status === 'running' ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800' :
                  container.status === 'paused' ? 'bg-amber-950/80 text-amber-400 border border-amber-800' :
                  container.status === 'oom_killed' ? 'bg-rose-950 text-rose-300 border border-rose-800' :
                  'bg-slate-800 text-slate-400 border border-slate-700'
                ]"
              >
                {{ container.status === 'oom_killed' ? 'OOM-Killed (137)' : container.status }}
              </span>
            </div>
            <div class="text-xs text-slate-400 font-mono flex items-center gap-2 mt-0.5">
              <span>Image: <strong class="text-[#0db7ed]">{{ container.image }}</strong></span>
              <span>•</span>
              <span>ID: {{ container.id.slice(0, 12) }}</span>
              <span>•</span>
              <span>IP: {{ container.namespaces.virtualIp }}</span>
              <span v-if="container.portBindings.length > 0">•</span>
              <span v-if="container.portBindings.length > 0" class="text-cyan-300">{{ container.portBindings.join(', ') }}</span>
            </div>
          </div>
        </div>

        <!-- Quick Lifecycle Controls -->
        <div class="flex items-center gap-2">
          <template v-if="container.status === 'running'">
            <button 
              @click="handleAction('pause')"
              class="px-2.5 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3446] text-amber-400 border border-[#2E3B4E] text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
              title="Pause (cgroup.freeze)"
            >
              <Pause class="w-3.5 h-3.5" />
              <span>Pause</span>
            </button>
            <button 
              @click="handleAction('stop')"
              class="px-2.5 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3446] text-slate-200 border border-[#2E3B4E] text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
              title="Stop (SIGTERM)"
            >
              <Square class="w-3.5 h-3.5" />
              <span>Stop</span>
            </button>
            <button 
              @click="handleAction('kill')"
              class="px-2.5 py-1.5 rounded-lg bg-rose-950/70 hover:bg-rose-900 text-rose-300 border border-rose-800/80 text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
              title="Kill (SIGKILL 137)"
            >
              <Flame class="w-3.5 h-3.5" />
              <span>Kill</span>
            </button>
          </template>
          <template v-else-if="container.status === 'paused'">
            <button 
              @click="handleAction('unpause')"
              class="px-2.5 py-1.5 rounded-lg bg-emerald-950 hover:bg-emerald-900 text-emerald-300 border border-emerald-800 text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Play class="w-3.5 h-3.5" />
              <span>Resume</span>
            </button>
            <button 
              @click="handleAction('stop')"
              class="px-2.5 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3446] text-slate-200 border border-[#2E3B4E] text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Square class="w-3.5 h-3.5" />
              <span>Stop</span>
            </button>
          </template>
          <template v-else>
            <button 
              @click="handleAction('start')"
              class="px-3 py-1.5 rounded-lg bg-emerald-700 hover:bg-emerald-600 text-white text-xs font-medium flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Play class="w-3.5 h-3.5" />
              <span>Start</span>
            </button>
          </template>

          <button 
            @click="emit('close')"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-[#222A38] transition-colors cursor-pointer ml-2"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Docker Desktop 5 Signature Tabs: Logs, Inspect, Exec, Files, Stats -->
      <div class="px-5 bg-[#11161D] border-b border-[#232A35] flex items-center gap-2 overflow-x-auto text-xs font-medium">
        <!-- 1. Logs Tab -->
        <button 
          @click="activeTab = 'logs'"
          class="px-4 py-2.5 border-b-2 transition-all flex items-center gap-2 cursor-pointer"
          :class="activeTab === 'logs' ? 'border-[#1D63ED] text-white font-semibold' : 'border-transparent text-slate-400 hover:text-slate-200'"
        >
          <FileText class="w-4 h-4 text-cyan-400" />
          <span>Logs</span>
        </button>

        <!-- 2. Exec (Terminal) Tab -->
        <button 
          @click="activeTab = 'exec'"
          class="px-4 py-2.5 border-b-2 transition-all flex items-center gap-2 cursor-pointer font-mono"
          :class="activeTab === 'exec' ? 'border-[#1D63ED] text-white font-semibold' : 'border-transparent text-slate-400 hover:text-slate-200'"
        >
          <TerminalIcon class="w-4 h-4 text-emerald-400" />
          <span>Exec</span>
        </button>

        <!-- 3. Inspect (JSON) Tab -->
        <button 
          @click="activeTab = 'inspect'"
          class="px-4 py-2.5 border-b-2 transition-all flex items-center gap-2 cursor-pointer font-mono"
          :class="activeTab === 'inspect' ? 'border-[#1D63ED] text-white font-semibold' : 'border-transparent text-slate-400 hover:text-slate-200'"
        >
          <Code class="w-4 h-4 text-amber-400" />
          <span>Inspect</span>
        </button>

        <!-- 4. Files (OverlayFS) Tab -->
        <button 
          @click="activeTab = 'files'"
          class="px-4 py-2.5 border-b-2 transition-all flex items-center gap-2 cursor-pointer"
          :class="activeTab === 'files' ? 'border-[#1D63ED] text-white font-semibold' : 'border-transparent text-slate-400 hover:text-slate-200'"
        >
          <FolderTree class="w-4 h-4 text-indigo-400" />
          <span>Files</span>
        </button>

        <!-- 5. Stats & Cgroups Tab -->
        <button 
          @click="activeTab = 'stats'"
          class="px-4 py-2.5 border-b-2 transition-all flex items-center gap-2 cursor-pointer"
          :class="activeTab === 'stats' ? 'border-[#1D63ED] text-white font-semibold' : 'border-transparent text-slate-400 hover:text-slate-200'"
        >
          <Activity class="w-4 h-4 text-rose-400" />
          <span>Stats & Cgroups</span>
        </button>
      </div>

      <!-- Tab Content Area -->
      <div class="flex-1 overflow-hidden flex flex-col bg-[#0F1318]">
        <!-- TAB 1: LOGS -->
        <div v-if="activeTab === 'logs'" class="flex-1 flex flex-col overflow-hidden">
          <div class="p-2.5 px-4 bg-[#161B22] border-b border-[#232A35] flex items-center justify-between gap-3 text-xs">
            <div class="relative flex-1 max-w-sm">
              <Search class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input 
                v-model="logFilter"
                type="text" 
                placeholder="Filter logs..."
                class="w-full pl-8 pr-3 py-1 rounded bg-[#0E1217] border border-[#2B3545] text-xs font-mono text-slate-200 focus:outline-none focus:border-[#1D63ED]"
              />
            </div>
            <div class="flex items-center gap-3 text-slate-400 font-mono text-[11px]">
              <span>{{ filteredLogs.length }} lines</span>
              <button 
                @click="emit('refresh')"
                class="p-1 rounded hover:bg-slate-800 text-slate-300 hover:text-white transition-colors cursor-pointer"
                title="Refresh Logs"
              >
                <RefreshCw class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>

          <div 
            ref="logsContainerRef"
            class="flex-1 p-4 overflow-y-auto font-mono text-xs space-y-1 text-slate-300 selection:bg-cyan-500/30 leading-relaxed"
          >
            <div 
              v-for="(log, idx) in filteredLogs" 
              :key="idx" 
              class="flex items-start gap-3 hover:bg-[#1A222D]/60 px-1 py-0.5 rounded"
            >
              <span class="text-slate-500 text-[11px] select-none shrink-0 w-20">{{ log.timestamp }}</span>
              <span 
                class="text-[10px] uppercase font-bold px-1 rounded select-none shrink-0"
                :class="log.stream === 'stderr' ? 'bg-rose-950 text-rose-400' : log.stream === 'system' ? 'bg-indigo-950 text-indigo-300' : 'bg-slate-800 text-cyan-300'"
              >
                {{ log.stream }}
              </span>
              <span 
                class="whitespace-pre-wrap break-all"
                :class="log.stream === 'stderr' ? 'text-rose-300' : log.stream === 'system' ? 'text-indigo-200' : 'text-slate-200'"
              >
                {{ log.message }}
              </span>
            </div>
            <div v-if="filteredLogs.length === 0" class="text-slate-500 py-8 text-center">
              No logs recorded yet.
            </div>
          </div>
        </div>

        <!-- TAB 2: EXEC (TERMINAL) -->
        <div v-else-if="activeTab === 'exec'" class="flex-1 flex flex-col overflow-hidden">
          <!-- Terminal Quick commands -->
          <div class="px-4 py-2 bg-[#161B22] border-b border-[#232A35] flex items-center gap-2 overflow-x-auto text-[11px] font-mono">
            <span class="text-slate-400 shrink-0 flex items-center gap-1 text-[10px]">
              <Sparkles class="w-3 h-3 text-cyan-400" /> Quick Exec:
            </span>
            <button @click="runCommand('uname -a')" class="px-2 py-0.5 rounded bg-[#202836] hover:bg-[#2A3649] text-cyan-300 border border-[#2D394C] transition-colors shrink-0 cursor-pointer">
              uname -a
            </button>
            <button @click="runCommand('hostname')" class="px-2 py-0.5 rounded bg-[#202836] hover:bg-[#2A3649] text-cyan-300 border border-[#2D394C] transition-colors shrink-0 cursor-pointer">
              hostname
            </button>
            <button @click="runCommand('ps aux')" class="px-2 py-0.5 rounded bg-[#202836] hover:bg-[#2A3649] text-emerald-300 border border-[#2D394C] transition-colors shrink-0 cursor-pointer">
              ps aux
            </button>
            <button @click="runCommand('cat /etc/os-release')" class="px-2 py-0.5 rounded bg-[#202836] hover:bg-[#2A3649] text-slate-300 border border-[#2D394C] transition-colors shrink-0 cursor-pointer">
              cat /etc/os-release
            </button>
            <button @click="runCommand('ip a')" class="px-2 py-0.5 rounded bg-[#202836] hover:bg-[#2A3649] text-blue-300 border border-[#2D394C] transition-colors shrink-0 cursor-pointer">
              ip a
            </button>
          </div>

          <!-- Terminal viewport -->
          <div 
            ref="terminalBody"
            class="flex-1 p-4 overflow-y-auto bg-[#0A0D12] font-mono text-xs space-y-2 selection:bg-[#1D63ED]/40"
          >
            <div 
              v-for="(line, idx) in terminalLines" 
              :key="idx" 
              class="flex items-start gap-2 leading-relaxed"
            >
              <span class="text-slate-600 text-[10px] select-none shrink-0 w-14">{{ line.time }}</span>
              <template v-if="line.stream === 'stdin'">
                <span class="text-[#0db7ed] font-semibold select-none">root@{{ container.namespaces.hostname }}:/#</span>
                <span class="text-white font-medium whitespace-pre-wrap break-all">{{ line.text.replace('$ ', '') }}</span>
              </template>
              <template v-else-if="line.stream === 'stderr'">
                <span class="text-rose-400 select-none">[stderr]</span>
                <span class="text-rose-300 whitespace-pre-wrap break-all">{{ line.text }}</span>
              </template>
              <template v-else>
                <span class="text-slate-300 whitespace-pre-wrap break-all">{{ line.text }}</span>
              </template>
            </div>
            <div v-if="execLoading" class="flex items-center gap-2 text-cyan-400 pt-2 text-[11px]">
              <span class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-ping"></span>
              <span>Executing in container PID namespace...</span>
            </div>
          </div>

          <!-- Interactive CLI prompt -->
          <div class="p-3 border-t border-[#232A35] bg-[#161B22]">
            <form @submit.prevent="runCommand()" class="flex items-center gap-2 font-mono text-xs">
              <span class="text-[#0db7ed] font-semibold shrink-0 select-none">
                root@{{ container.namespaces.hostname }}:/#
              </span>
              <input 
                v-model="inputCommand"
                type="text" 
                placeholder="Type command (e.g. uname -a, ps aux, touch /tmp/demo.txt)..."
                class="flex-1 bg-transparent text-slate-100 placeholder:text-slate-600 focus:outline-none"
                :disabled="execLoading || container.status !== 'running'"
              />
              <button 
                type="submit" 
                :disabled="execLoading || !inputCommand.trim() || container.status !== 'running'"
                class="px-3 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-30 text-white font-medium flex items-center gap-1 transition-colors cursor-pointer"
              >
                <Send class="w-3.5 h-3.5" />
                <span>Send</span>
              </button>
            </form>
          </div>
        </div>

        <!-- TAB 3: INSPECT (JSON) -->
        <div v-else-if="activeTab === 'inspect'" class="flex-1 flex flex-col overflow-hidden p-4">
          <div class="flex items-center justify-between mb-3">
            <span class="text-xs font-mono text-slate-400">docker inspect {{ container.id.slice(0, 12) }}</span>
            <button 
              @click="copyInspect"
              class="px-3 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3649] text-xs font-medium text-slate-200 border border-[#2D394C] flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Check v-if="copiedInspect" class="w-3.5 h-3.5 text-emerald-400" />
              <Copy v-else class="w-3.5 h-3.5 text-slate-400" />
              <span>{{ copiedInspect ? 'Copied JSON' : 'Copy JSON' }}</span>
            </button>
          </div>
          <pre class="flex-1 p-4 rounded-xl bg-[#0B0E14] border border-[#232A35] text-cyan-200 text-xs font-mono overflow-auto leading-relaxed selection:bg-[#1D63ED]/30"><code>{{ JSON.stringify({
  Id: container.id,
  Created: container.createdAt,
  Path: container.command,
  State: {
    Status: container.status,
    Running: container.status === 'running',
    Paused: container.status === 'paused',
    OOMKilled: container.status === 'oom_killed',
    Pid: container.hostPid,
    ExitCode: container.status === 'oom_killed' ? 137 : 0,
    StartedAt: container.startedAt
  },
  Image: container.image,
  HostConfig: {
    Memory: container.cgroups.memoryMaxBytes,
    CpuQuota: container.cgroups.cpuQuotaUs,
    CpuPeriod: container.cgroups.cpuPeriodUs,
    CgroupPath: container.cgroups.cgroupPath,
    Namespaces: {
      CLONE_NEWPID: container.namespaces.pidInode,
      CLONE_NEWUTS: container.namespaces.utsInode,
      CLONE_NEWNET: container.namespaces.netInode,
      CLONE_NEWNS: container.namespaces.mntInode
    }
  },
  NetworkSettings: {
    IPAddress: container.namespaces.virtualIp,
    Ports: container.portBindings,
    VethPair: container.namespaces.vethPair
  },
  Mounts: [
    {
      Type: "overlay",
      LowerDir: container.rootfs.lowerDir,
      UpperDir: container.rootfs.upperDir,
      MergedDir: container.rootfs.mergedDir
    }
  ]
}, null, 2) }}</code></pre>
        </div>

        <!-- TAB 4: FILES (OVERLAYFS) -->
        <div v-else-if="activeTab === 'files'" class="flex-1 overflow-y-auto p-5 space-y-5 font-mono text-xs">
          <!-- Architecture Banner -->
          <div class="p-4 rounded-xl bg-[#161B22] border border-[#232A35] space-y-3">
            <div class="text-xs font-bold text-slate-200 flex items-center justify-between">
              <span class="flex items-center gap-2">
                <FolderTree class="w-4 h-4 text-cyan-400" />
                <span>OverlayFS Layering & Copy-on-Write (COW)</span>
              </span>
              <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-950 text-emerald-300 border border-emerald-800">
                pivot_root active
              </span>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-[11px]">
              <div class="p-3 rounded-lg bg-[#0F1318] border border-[#232A35]">
                <div class="text-slate-400 font-bold">MergedDir (Container /)</div>
                <div class="text-emerald-400 font-semibold mt-1 truncate">{{ container.rootfs.mergedDir }}</div>
                <div class="text-[10px] text-slate-500 mt-0.5">Isolated mount namespace root</div>
              </div>
              <div class="p-3 rounded-lg bg-[#0F1318] border border-[#232A35]">
                <div class="text-slate-400 font-bold">UpperDir (Diff Layer)</div>
                <div class="text-cyan-400 font-semibold mt-1 truncate">{{ container.rootfs.upperDir }}</div>
                <div class="text-[10px] text-slate-500 mt-0.5">{{ container.rootfs.modifiedFiles.length }} modified files</div>
              </div>
              <div class="p-3 rounded-lg bg-[#0F1318] border border-[#232A35]">
                <div class="text-slate-400 font-bold">LowerDir (Base Image)</div>
                <div class="text-slate-300 font-semibold mt-1 truncate">{{ container.rootfs.lowerDir }}</div>
                <div class="text-[10px] text-slate-500 mt-0.5">Read-only base tarball</div>
              </div>
            </div>
          </div>

          <!-- Touch file demo -->
          <div class="p-4 rounded-xl bg-[#161B22] border border-[#232A35] space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-200">Copy-on-Write (UpperDir) File Mutations</span>
              <span class="text-[10px] text-slate-500">Tracked in real-time</span>
            </div>

            <form @submit.prevent="handleCreateFile" class="flex items-center gap-2">
              <input 
                v-model="newFileName"
                type="text" 
                placeholder="/app/config.json or /tmp/test.log"
                class="flex-1 px-3 py-1.5 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
                :disabled="container.status !== 'running'"
              />
              <button 
                type="submit"
                :disabled="!newFileName.trim() || isCreatingFile || container.status !== 'running'"
                class="px-3 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-30 text-white font-medium text-xs flex items-center gap-1 transition-colors cursor-pointer"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>Touch File</span>
              </button>
            </form>

            <div class="rounded-lg border border-[#232A35] overflow-hidden">
              <table class="w-full text-left text-xs">
                <thead class="bg-[#11161D] text-slate-400 border-b border-[#232A35]">
                  <tr>
                    <th class="px-3 py-2">File Path</th>
                    <th class="px-3 py-2">Action</th>
                    <th class="px-3 py-2">Size</th>
                    <th class="px-3 py-2">Timestamp</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-[#232A35]">
                  <tr v-for="(file, idx) in container.rootfs.modifiedFiles" :key="idx" class="hover:bg-[#1A222D]">
                    <td class="px-3 py-2 text-cyan-300 font-semibold">{{ file.path }}</td>
                    <td class="px-3 py-2">
                      <span class="px-1.5 py-0.5 rounded text-[10px] bg-cyan-950 text-cyan-400 border border-cyan-800">
                        {{ file.action }}
                      </span>
                    </td>
                    <td class="px-3 py-2 text-slate-400">{{ file.sizeBytes }} B</td>
                    <td class="px-3 py-2 text-slate-500">{{ file.timestamp }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- TAB 5: STATS & CGROUPS -->
        <div v-else-if="activeTab === 'stats'" class="flex-1 overflow-y-auto p-5 space-y-5 font-mono text-xs">
          <!-- Alert notification if OOM-killed -->
          <div v-if="statsAlert" class="p-3 bg-rose-950/80 border border-rose-800 text-rose-200 text-xs flex items-start gap-2 rounded-lg">
            <AlertTriangle class="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
            <div>{{ statsAlert }}</div>
          </div>

          <!-- Memory cgroup -->
          <div class="p-4 rounded-xl bg-[#161B22] border border-[#232A35] space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-bold text-slate-200 flex items-center gap-2">
                <HardDrive class="w-4 h-4 text-cyan-400" />
                <span>Memory Controller (memory.max)</span>
              </span>
              <span class="text-xs px-2 py-0.5 rounded bg-[#202836] text-slate-300 border border-[#2B3545]">
                OOM Kill Events: <strong class="text-rose-400">{{ container.cgroups.oomKillEvents }}</strong>
              </span>
            </div>

            <div>
              <div class="flex justify-between text-xs mb-1.5">
                <span class="text-slate-400">Memory Usage vs Hard Limit:</span>
                <span :class="memPercent > 80 ? 'text-rose-400 font-bold' : 'text-slate-200'">
                  {{ memUsedMB }} MB / {{ memLimitMB }} MB ({{ memPercent }}%)
                </span>
              </div>
              <div class="w-full h-3 rounded-full bg-[#0F1318] border border-[#232A35] overflow-hidden">
                <div 
                  class="h-full rounded-full transition-all duration-300"
                  :class="memPercent > 85 ? 'bg-rose-500' : memPercent > 60 ? 'bg-amber-500' : 'bg-[#1D63ED]'"
                  :style="{ width: `${memPercent}%` }"
                ></div>
              </div>
            </div>

            <div class="pt-2 border-t border-[#232A35] flex items-center gap-2 flex-wrap">
              <span class="text-[11px] text-slate-400">Stress Allocator:</span>
              <button
                @click="handleStressMem(16)"
                :disabled="statsLoading || container.status !== 'running'"
                class="px-2.5 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3446] disabled:opacity-30 text-cyan-300 border border-[#2D394C] transition-colors cursor-pointer text-xs"
              >
                +16 MB
              </button>
              <button
                @click="handleStressMem(32)"
                :disabled="statsLoading || container.status !== 'running'"
                class="px-2.5 py-1.5 rounded-lg bg-[#202836] hover:bg-[#2A3446] disabled:opacity-30 text-cyan-300 border border-[#2D394C] transition-colors cursor-pointer text-xs"
              >
                +32 MB
              </button>
              <button
                @click="handleForceOom"
                :disabled="statsLoading || container.status !== 'running'"
                class="px-3 py-1.5 rounded-lg bg-rose-950 hover:bg-rose-900 disabled:opacity-30 text-rose-300 border border-rose-800 transition-colors cursor-pointer text-xs font-bold flex items-center gap-1.5 ml-auto"
                title="Exceed memory.max to trigger kernel OOM killer"
              >
                <Skull class="w-4 h-4 text-rose-400" />
                <span>Trigger Kernel OOM Killer</span>
              </button>
            </div>
          </div>

          <!-- CPU cgroup -->
          <div class="p-4 rounded-xl bg-[#161B22] border border-[#232A35] space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-bold text-slate-200 flex items-center gap-2">
                <Cpu class="w-4 h-4 text-indigo-400" />
                <span>CFS Bandwidth Control (cpu.max)</span>
              </span>
              <span class="text-xs px-2 py-0.5 rounded bg-[#202836] text-indigo-300 border border-[#2B3545]">
                Quota: {{ container.cgroups.cpuQuotaUs }}us / {{ container.cgroups.cpuPeriodUs }}us
              </span>
            </div>

            <div class="grid grid-cols-2 gap-3 text-xs">
              <div class="p-3 rounded-lg bg-[#0F1318] border border-[#232A35]">
                <div class="text-slate-500 text-[10px]">CFS Quota Allocation</div>
                <div class="text-base font-bold text-slate-100 mt-0.5">{{ cpuQuotaCores }} Cores</div>
              </div>
              <div class="p-3 rounded-lg bg-[#0F1318] border border-[#232A35]">
                <div class="text-slate-500 text-[10px]">CFS Throttled Periods</div>
                <div class="text-base font-bold text-amber-400 mt-0.5">{{ container.cgroups.throttlePeriods }} events</div>
              </div>
            </div>

            <div class="pt-2 border-t border-[#232A35] flex items-center justify-between">
              <span class="text-[11px] text-slate-400">CFS Scheduler Stress:</span>
              <button
                @click="handleStressCpu"
                :disabled="statsLoading || container.status !== 'running'"
                class="px-3 py-1.5 rounded-lg bg-indigo-950 hover:bg-indigo-900 disabled:opacity-30 text-indigo-200 border border-indigo-800 transition-colors cursor-pointer text-xs flex items-center gap-1.5"
              >
                <Flame class="w-3.5 h-3.5 text-indigo-400" />
                <span>Stress CPU (Trigger Throttling)</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
