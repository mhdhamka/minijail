<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted } from 'vue';
import type { Container, MetricPoint } from '../../types/container';
import { stressCpu, stressMemory, startContainer, updateCgroupLimits } from '../../api';
import { 
  Cpu, 
  HardDrive, 
  AlertTriangle, 
  ShieldAlert, 
  Sliders, 
  Play, 
  Pause,
  CheckCircle2, 
  FileCode2, 
  Info,
  Download,
} from 'lucide-vue-next';

interface Props {
  containers: Container[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'refresh'): void;
}>();

const selectedContainerId = ref<string>('');
const isStressingCpu = ref(false);
const isAllocatingMem = ref(false);
const isUpdatingCgroups = ref(false);
const actionMessage = ref<string | null>(null);

// Advanced features: Stream freeze & Chart hover inspection
const isStreamPaused = ref(false);
const hoveredCpuPoint = ref<MetricPoint | null>(null);
const hoveredMemPoint = ref<MetricPoint | null>(null);

// Editable cgroups values for fine-tuning
const targetCpuCores = ref<number>(0.5);
const targetMemMb = ref<number>(64);

// Initialize selected container
onMounted(() => {
  if (props.containers.length > 0) {
    const running = props.containers.find(c => c.status === 'running');
    selectedContainerId.value = running ? running.id : props.containers[0].id;
  }
});

watch(
  () => props.containers,
  (newVal) => {
    if (isStreamPaused.value) return; // Respect pause state
    if (!selectedContainerId.value && newVal.length > 0) {
      const running = newVal.find(c => c.status === 'running');
      selectedContainerId.value = running ? running.id : newVal[0].id;
    }
    const current = currentContainer.value;
    if (current) {
      const quotaCores = +(current.cgroups.cpuQuotaUs / current.cgroups.cpuPeriodUs).toFixed(2);
      const memMb = Math.round(current.cgroups.memoryMaxBytes / (1024 * 1024));
      targetCpuCores.value = quotaCores;
      targetMemMb.value = memMb;
    }
  },
  { deep: true }
);

const currentContainer = computed<Container | undefined>(() => {
  return props.containers.find(c => c.id === selectedContainerId.value);
});

const history = computed<MetricPoint[]>(() => {
  if (!currentContainer.value?.metricsHistory) return [];
  return currentContainer.value.metricsHistory;
});

const currentCpuPercent = computed(() => {
  if (!currentContainer.value) return 0;
  return currentContainer.value.status === 'running' ? currentContainer.value.cgroups.cpuPercent : 0;
});

const cpuQuotaPercent = computed(() => {
  if (!currentContainer.value) return 100;
  return (currentContainer.value.cgroups.cpuQuotaUs / currentContainer.value.cgroups.cpuPeriodUs) * 100;
});

const isCurrentlyThrottled = computed(() => {
  if (!currentContainer.value || currentContainer.value.status !== 'running') return false;
  return currentCpuPercent.value >= (cpuQuotaPercent.value - 0.5);
});

const currentMemMb = computed(() => {
  if (!currentContainer.value) return 0;
  return +(currentContainer.value.cgroups.memoryUsageBytes / (1024 * 1024)).toFixed(1);
});

const maxMemMb = computed(() => {
  if (!currentContainer.value) return 128;
  return Math.round(currentContainer.value.cgroups.memoryMaxBytes / (1024 * 1024));
});

const memUsagePercent = computed(() => {
  if (maxMemMb.value <= 0) return 0;
  return Math.min(100, +((currentMemMb.value / maxMemMb.value) * 100).toFixed(1));
});

// SVG Chart Metrics & Scaling
const chartWidth = 560;
const chartHeight = 160;
const padding = { top: 15, right: 15, bottom: 25, left: 35 };

const cpuChartPaths = computed(() => {
  const pts = history.value;
  if (!pts || pts.length < 2) return { line: '', area: '', quotaY: 50, coords: [] };

  const innerW = chartWidth - padding.left - padding.right;
  const innerH = chartHeight - padding.top - padding.bottom;

  const maxQuota = cpuQuotaPercent.value;
  const maxVal = Math.max(100, maxQuota * 1.25);

  const coords = pts.map((p, idx) => {
    const x = padding.left + (idx / (pts.length - 1)) * innerW;
    const yVal = Math.min(maxVal, Math.max(0, p.cpuPercent));
    const y = padding.top + innerH - (yVal / maxVal) * innerH;
    return { x, y, isThrottled: p.isThrottled, point: p };
  });

  const line = coords.map((c, i) => `${i === 0 ? 'M' : 'L'} ${c.x.toFixed(1)} ${c.y.toFixed(1)}`).join(' ');
  const area = `${line} L ${coords[coords.length - 1].x.toFixed(1)} ${(padding.top + innerH).toFixed(1)} L ${coords[0].x.toFixed(1)} ${(padding.top + innerH).toFixed(1)} Z`;
  const quotaY = padding.top + innerH - (maxQuota / maxVal) * innerH;

  return { line, area, coords, quotaY, maxVal };
});

const memoryChartPaths = computed(() => {
  const pts = history.value;
  if (!pts || pts.length < 2) return { line: '', area: '', limitY: 40, coords: [] };

  const innerW = chartWidth - padding.left - padding.right;
  const innerH = chartHeight - padding.top - padding.bottom;

  const limitBytes = currentContainer.value?.cgroups.memoryMaxBytes || 64 * 1024 * 1024;
  const maxBytes = limitBytes * 1.15;

  const coords = pts.map((p, idx) => {
    const x = padding.left + (idx / (pts.length - 1)) * innerW;
    const yVal = Math.min(maxBytes, Math.max(0, p.memoryUsageBytes));
    const y = padding.top + innerH - (yVal / maxBytes) * innerH;
    return { x, y, point: p };
  });

  const line = coords.map((c, i) => `${i === 0 ? 'M' : 'L'} ${c.x.toFixed(1)} ${c.y.toFixed(1)}`).join(' ');
  const area = `${line} L ${coords[coords.length - 1].x.toFixed(1)} ${(padding.top + innerH).toFixed(1)} L ${coords[0].x.toFixed(1)} ${(padding.top + innerH).toFixed(1)} Z`;
  const limitY = padding.top + innerH - (limitBytes / maxBytes) * innerH;
  const warningY = padding.top + innerH - ((limitBytes * 0.85) / maxBytes) * innerH;

  return { line, area, coords, limitY, warningY, maxBytes };
});

// Interactive SVG Mouse Move handlers for tooltips
function handleCpuMouseMove(event: MouseEvent) {
  const rect = (event.currentTarget as SVGElement).getBoundingClientRect();
  const mouseX = event.clientX - rect.left;
  const svgX = (mouseX / rect.width) * chartWidth;
  
  const coords = cpuChartPaths.value.coords;
  if (!coords || coords.length === 0) return;

  let closest = coords[0];
  let minDst = Math.abs(coords[0].x - svgX);
  for (const c of coords) {
    const dst = Math.abs(c.x - svgX);
    if (dst < minDst) {
      minDst = dst;
      closest = c;
    }
  }
  hoveredCpuPoint.value = closest.point;
}

function handleMemMouseMove(event: MouseEvent) {
  const rect = (event.currentTarget as SVGElement).getBoundingClientRect();
  const mouseX = event.clientX - rect.left;
  const svgX = (mouseX / rect.width) * chartWidth;
  
  const coords = memoryChartPaths.value.coords;
  if (!coords || coords.length === 0) return;

  let closest = coords[0];
  let minDst = Math.abs(coords[0].x - svgX);
  for (const c of coords) {
    const dst = Math.abs(c.x - svgX);
    if (dst < minDst) {
      minDst = dst;
      closest = c;
    }
  }
  hoveredMemPoint.value = closest.point;
}

async function handleStressCpu() {
  if (!currentContainer.value || isStressingCpu.value) return;
  isStressingCpu.value = true;
  actionMessage.value = `[CFS Load Generator] Spawning busy CPU threads to saturate quota...`;
  try {
    await stressCpu(currentContainer.value.id);
    emit('refresh');
    actionMessage.value = `CFS scheduler actively throttling container threads at ${cpuQuotaPercent.value.toFixed(1)}% quota.`;
  } catch (err: any) {
    actionMessage.value = `Error: ${err.message}`;
  } finally {
    setTimeout(() => { isStressingCpu.value = false; }, 800);
  }
}

async function handleStressMemory(deltaMb = 24) {
  if (!currentContainer.value || isAllocatingMem.value) return;
  isAllocatingMem.value = true;
  actionMessage.value = `[Kernel Allocator] Allocating +${deltaMb}MB in anonymous RSS memory...`;
  try {
    const res = await stressMemory(currentContainer.value.id, deltaMb);
    emit('refresh');
    if (res.oomKilled) {
      actionMessage.value = `OOM-Killer triggered! Exceeded memory.max (${maxMemMb.value}MB). Container killed with SIGKILL (137).`;
    } else {
      actionMessage.value = `Allocated +${deltaMb}MB. Usage is now ${+(res.container.cgroups.memoryUsageBytes / (1024 * 1024)).toFixed(1)}MB / ${maxMemMb.value}MB.`;
    }
  } catch (err: any) {
    actionMessage.value = `Error: ${err.message}`;
  } finally {
    setTimeout(() => { isAllocatingMem.value = false; }, 800);
  }
}

async function handleApplyCgroups() {
  if (!currentContainer.value || isUpdatingCgroups.value) return;
  isUpdatingCgroups.value = true;
  actionMessage.value = `Writing limits to /sys/fs/cgroup/minijail/${currentContainer.value.id.slice(0, 8)}...`;
  try {
    await updateCgroupLimits(currentContainer.value.id, {
      cpuQuotaCores: targetCpuCores.value,
      memoryLimitMb: targetMemMb.value,
    });
    emit('refresh');
    actionMessage.value = `cgroups v2 limits updated successfully.`;
  } catch (err: any) {
    actionMessage.value = `Failed to update cgroups: ${err.message}`;
  } finally {
    isUpdatingCgroups.value = false;
  }
}

async function handleRestartContainer() {
  if (!currentContainer.value) return;
  try {
    await startContainer(currentContainer.value.id);
    emit('refresh');
    actionMessage.value = `Container ${currentContainer.value.name} restarted in isolated PID & cgroup hierarchy.`;
  } catch (err: any) {
    actionMessage.value = `Restart error: ${err.message}`;
  }
}

function exportDiagnostics() {
  if (!currentContainer.value) return;
  const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(currentContainer.value, null, 2));
  const downloadAnchor = document.createElement('a');
  downloadAnchor.setAttribute("href", dataStr);
  downloadAnchor.setAttribute("download", `cgroups-diag-${currentContainer.value.name}-${Date.now()}.json`);
  document.body.appendChild(downloadAnchor);
  downloadAnchor.click();
  downloadAnchor.remove();
  actionMessage.value = `Exported diagnostic package for ${currentContainer.value.name}.`;
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-slate-800 pb-4">
      <div>
        <div class="flex items-center gap-2">
          <h2 class="text-xl font-bold text-white tracking-tight">Real-Time Resource & Cgroups Throttling</h2>
          <span 
            class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-mono font-medium border transition-colors"
            :class="isStreamPaused ? 'bg-amber-500/10 text-amber-400 border-amber-500/30' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'"
          >
            {{ isStreamPaused ? 'Stream Paused' : 'Live (1.8s kernel tick)' }}
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Observing how the Linux Completely Fair Scheduler (CFS) and cgroups v2 enforce CPU period quotas and memory OOM limits.
        </p>
      </div>

      <!-- Controls & Container Selector (Forced to Single Line) -->
      <div class="flex items-center gap-2.5 shrink-0">
        <button 
          @click="isStreamPaused = !isStreamPaused"
          class="px-3 py-2 bg-slate-900 hover:bg-slate-800 border border-slate-700 text-slate-200 rounded-lg text-xs font-mono flex items-center gap-1.5 transition-all cursor-pointer shrink-0"
          :title="isStreamPaused ? 'Resume live updates' : 'Pause live telemetry stream'"
        >
          <component :is="isStreamPaused ? Play : Pause" class="w-3.5 h-3.5 text-cyan-400" />
          {{ isStreamPaused ? 'Resume' : 'Pause' }}
        </button>

        <button 
          @click="exportDiagnostics"
          class="px-3 py-2 bg-slate-900 hover:bg-slate-800 border border-slate-700 text-slate-200 rounded-lg text-xs font-mono flex items-center gap-1.5 transition-all cursor-pointer shrink-0"
          title="Export JSON Diagnostic Bundle"
        >
          <Download class="w-3.5 h-3.5 text-blue-400" />
          Export JSON
        </button>

        <select 
          v-model="selectedContainerId"
          class="bg-slate-900 border border-slate-700 text-xs font-mono text-slate-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition-all truncate shrink-0"
        >
          <option v-for="c in containers" :key="c.id" :value="c.id">
            {{ c.name }} ({{ c.status }}) · {{ c.image }}
          </option>
        </select>
      </div>
    </div>

    <!-- Active Container Status & Cgroups Overview Banner -->
    <div v-if="currentContainer" class="bg-slate-900/80 backdrop-blur border border-slate-800 rounded-xl p-5 shadow-lg">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <!-- Container Identity -->
        <div class="flex items-center gap-3.5">
          <div 
            class="w-11 h-11 rounded-xl flex items-center justify-center font-mono font-bold text-sm shadow-inner"
            :class="currentContainer.status === 'running' 
              ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-700/50' 
              : currentContainer.status === 'oom_killed'
              ? 'bg-rose-950/80 text-rose-400 border border-rose-700/50'
              : 'bg-slate-800 text-slate-400 border border-slate-700'"
          >
            {{ currentContainer.name.slice(0, 2).toUpperCase() }}
          </div>
          <div>
            <div class="flex items-center gap-2.5 flex-wrap">
              <span class="text-sm font-bold text-white font-mono">{{ currentContainer.name }}</span>
              <span class="text-xs text-slate-500 font-mono">ID: {{ currentContainer.id.slice(0, 12) }}</span>
              <span 
                class="text-[11px] font-mono font-medium px-2.5 py-0.5 rounded-md"
                :class="currentContainer.status === 'running'
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                  : currentContainer.status === 'oom_killed'
                  ? 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
                  : 'bg-slate-800 text-slate-400 border border-slate-700'"
              >
                {{ currentContainer.status === 'running' ? 'Running' : currentContainer.status === 'oom_killed' ? 'OOM-Killed (137)' : currentContainer.status }}
              </span>
            </div>
            <div class="text-xs text-slate-400 font-mono mt-1 flex items-center gap-3 flex-wrap">
              <span>Image: <strong class="text-slate-300">{{ currentContainer.image }}</strong></span>
              <span>•</span>
              <span>Host PID: <strong class="text-slate-300">{{ currentContainer.hostPid || '-' }}</strong></span>
              <span>•</span>
              <span class="text-slate-500">/sys/fs/cgroup/minijail/{{ currentContainer.id.slice(0, 8) }}</span>
            </div>
          </div>
        </div>

        <!-- Real-Time Metrics Badges -->
        <div class="flex items-center gap-3">
          <!-- CPU Badge -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl px-3.5 py-2.5 min-w-[140px]">
            <div class="text-[10px] text-slate-400 flex items-center justify-between font-mono">
              <span class="flex items-center gap-1.5"><Cpu class="w-3.5 h-3.5 text-cyan-400" /> CPU USAGE</span>
              <span :class="isCurrentlyThrottled ? 'text-amber-400 font-bold animate-pulse' : 'text-slate-500'">
                {{ isCurrentlyThrottled ? 'THROTTLED' : 'NORMAL' }}
              </span>
            </div>
            <div class="text-base font-bold font-mono text-white mt-0.5">
              {{ currentCpuPercent.toFixed(1) }}%
              <span class="text-xs font-normal text-slate-400">/ {{ cpuQuotaPercent.toFixed(0) }}% cap</span>
            </div>
          </div>

          <!-- Memory Badge -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl px-3.5 py-2.5 min-w-[140px]">
            <div class="text-[10px] text-slate-400 flex items-center justify-between font-mono">
              <span class="flex items-center gap-1.5"><HardDrive class="w-3.5 h-3.5 text-indigo-400" /> RSS MEMORY</span>
              <span :class="memUsagePercent > 85 ? 'text-rose-400 font-bold' : 'text-slate-500'">
                {{ memUsagePercent }}%
              </span>
            </div>
            <div class="text-base font-bold font-mono text-white mt-0.5">
              {{ currentMemMb }}MB
              <span class="text-xs font-normal text-slate-400">/ {{ maxMemMb }}MB</span>
            </div>
          </div>

          <!-- Restart button if killed -->
          <button 
            v-if="currentContainer.status !== 'running'"
            @click="handleRestartContainer"
            class="px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-xs font-semibold flex items-center gap-2 transition-all shadow-lg shadow-emerald-900/20 cursor-pointer"
          >
            <Play class="w-3.5 h-3.5 fill-current" />
            Restart Container
          </button>
        </div>
      </div>

      <!-- Action status notification bar -->
      <div v-if="actionMessage" class="mt-4 px-3.5 py-2 bg-slate-950/80 border border-slate-800 rounded-xl text-xs font-mono text-cyan-300 flex items-center gap-2.5 shadow-inner">
        <Info class="w-4 h-4 text-cyan-400 shrink-0" />
        <span class="truncate">{{ actionMessage }}</span>
      </div>
    </div>

    <!-- MAIN TWO INTERACTIVE CHARTS: CPU & MEMORY -->
    <div v-if="currentContainer" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      
      <!-- CHART 1: CPU USAGE & CFS QUOTA THROTTLING -->
      <div class="bg-slate-900/80 border border-slate-800 rounded-xl p-5 flex flex-col justify-between shadow-lg">
        <div>
          <!-- Chart Header -->
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2.5">
              <div class="w-3 h-3 rounded-full bg-cyan-400 shadow-sm shadow-cyan-400/50"></div>
              <h3 class="text-sm font-bold text-white tracking-tight">CPU Usage vs. CFS Scheduler Quota</h3>
            </div>
            <span class="text-xs font-mono text-slate-400">
              Limit: {{ (cpuQuotaPercent / 100).toFixed(2) }} cores ({{ cpuQuotaPercent.toFixed(0) }}%)
            </span>
          </div>

          <div class="flex items-center justify-between text-xs text-slate-400 mb-4">
            <span>Linux CFS restricts runtime per period. Hover chart to inspect history.</span>
            <span v-if="hoveredCpuPoint" class="font-mono text-cyan-300 bg-cyan-950/50 px-2 py-0.5 rounded border border-cyan-800/50">
              {{ hoveredCpuPoint.cpuPercent.toFixed(1) }}% {{ hoveredCpuPoint.isThrottled ? '[Throttled]' : '' }}
            </span>
          </div>

          <!-- SVG Chart Area with Mouse Move Tooltip -->
          <div 
            class="relative bg-slate-950 rounded-xl p-3 border border-slate-800/80 overflow-hidden shadow-inner cursor-crosshair"
            @mousemove="handleCpuMouseMove"
            @mouseleave="hoveredCpuPoint = null"
          >
            <div 
              v-if="isCurrentlyThrottled"
              class="absolute top-3 right-3 z-10 px-2.5 py-1 rounded-md bg-amber-500/20 border border-amber-500/40 text-amber-300 text-[10px] font-mono flex items-center gap-1.5 backdrop-blur animate-pulse pointer-events-none"
            >
              <AlertTriangle class="w-3 h-3 text-amber-400" />
              <span>CFS Quota Throttling Active</span>
            </div>

            <svg :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="w-full h-44 overflow-visible">
              <defs>
                <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#06b6d4" stop-opacity="0.35" />
                  <stop offset="100%" stop-color="#06b6d4" stop-opacity="0.0" />
                </linearGradient>
              </defs>

              <line :x1="padding.left" :y1="padding.top" :x2="chartWidth - padding.right" :y2="padding.top" stroke="#1e293b" stroke-dasharray="3 3" />
              <line :x1="padding.left" :y1="padding.top + (chartHeight - padding.top - padding.bottom) / 2" :x2="chartWidth - padding.right" :y2="padding.top + (chartHeight - padding.top - padding.bottom) / 2" stroke="#1e293b" stroke-dasharray="3 3" />
              <line :x1="padding.left" :y1="chartHeight - padding.bottom" :x2="chartWidth - padding.right" :y2="chartHeight - padding.bottom" stroke="#334155" />

              <line :x1="padding.left" :y1="cpuChartPaths.quotaY" :x2="chartWidth - padding.right" :y2="cpuChartPaths.quotaY" stroke="#f59e0b" stroke-width="1.8" stroke-dasharray="5 3" />
              <text :x="chartWidth - padding.right - 4" :y="cpuChartPaths.quotaY - 4" fill="#f59e0b" font-size="9" text-anchor="end" font-family="monospace">
                cpu.max ceiling: {{ cpuQuotaPercent.toFixed(0) }}%
              </text>

              <text :x="padding.left - 6" :y="padding.top + 4" fill="#64748b" font-size="9" text-anchor="end" font-family="monospace">
                {{ cpuChartPaths.maxVal?.toFixed(0) }}%
              </text>
              <text :x="padding.left - 6" :y="chartHeight - padding.bottom" fill="#64748b" font-size="9" text-anchor="end" font-family="monospace">
                0%
              </text>

              <path v-if="cpuChartPaths.area" :d="cpuChartPaths.area" fill="url(#cpuGradient)" />
              <path v-if="cpuChartPaths.line" :d="cpuChartPaths.line" fill="none" stroke="#06b6d4" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />

              <g v-if="cpuChartPaths.coords">
                <circle 
                  v-for="(pt, idx) in cpuChartPaths.coords" 
                  :key="idx" 
                  :cx="pt.x" 
                  :cy="pt.y" 
                  :r="pt.isThrottled ? 3.5 : 2" 
                  :fill="pt.isThrottled ? '#f59e0b' : '#06b6d4'" 
                  :stroke="pt.isThrottled ? '#b45309' : '#0891b2'"
                  stroke-width="1"
                />
              </g>
            </svg>
          </div>

          <!-- Bottom Telemetry Stats -->
          <div class="grid grid-cols-3 gap-3 mt-4 pt-4 border-t border-slate-800 text-xs font-mono">
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">Current CPU</div>
              <div class="text-white font-bold text-sm mt-0.5">{{ currentCpuPercent.toFixed(1) }}%</div>
            </div>
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">nr_throttled</div>
              <div class="text-amber-400 font-bold text-sm mt-0.5">{{ currentContainer.cgroups.throttlePeriods }} periods</div>
            </div>
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">throttled_usec</div>
              <div class="text-slate-300 font-bold text-sm mt-0.5">{{ (currentContainer.cgroups.throttledTimeUs / 1000).toFixed(1) }} ms</div>
            </div>
          </div>
        </div>

        <!-- Quick Stress Action for CPU -->
        <div class="mt-5 pt-4 border-t border-slate-800 flex items-center justify-between">
          <span class="text-xs text-slate-400">Trigger CPU work to observe CFS throttling:</span>
          <button 
            @click="handleStressCpu"
            :disabled="currentContainer.status !== 'running' || isStressingCpu"
            class="px-4 py-2 rounded-xl text-xs font-semibold bg-cyan-600 hover:bg-cyan-500 text-white flex items-center gap-1.5 transition-all shadow-md shadow-cyan-950/40 cursor-pointer disabled:opacity-50"
          >
            {{ isStressingCpu ? 'Generating Load...' : 'Stress CPU Worker' }}
          </button>
        </div>
      </div>

      <!-- CHART 2: MEMORY ANONYMOUS RSS & OOM KILLER -->
      <div class="bg-slate-900/80 border border-slate-800 rounded-xl p-5 flex flex-col justify-between shadow-lg">
        <div>
          <!-- Chart Header -->
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2.5">
              <div class="w-3 h-3 rounded-full bg-indigo-400 shadow-sm shadow-indigo-400/50"></div>
              <h3 class="text-sm font-bold text-white tracking-tight">Memory RSS vs. cgroups memory.max Limit</h3>
            </div>
            <span class="text-xs font-mono text-slate-400">
              Hard Ceiling: {{ maxMemMb }}MB
            </span>
          </div>

          <div class="flex items-center justify-between text-xs text-slate-400 mb-4">
            <span>cgroups v2 enforces strict memory.max ceilings.</span>
            <span v-if="hoveredMemPoint" class="font-mono text-indigo-300 bg-indigo-950/50 px-2 py-0.5 rounded border border-indigo-800/50">
              {{ +(hoveredMemPoint.memoryUsageBytes / (1024 * 1024)).toFixed(1) }} MB
            </span>
          </div>

          <!-- SVG Chart Area with Mouse Move Tooltip -->
          <div 
            class="relative bg-slate-950 rounded-xl p-3 border border-slate-800/80 overflow-hidden shadow-inner cursor-crosshair"
            @mousemove="handleMemMouseMove"
            @mouseleave="hoveredMemPoint = null"
          >
            <div 
              v-if="currentContainer.status === 'oom_killed'"
              class="absolute top-3 right-3 z-10 px-2.5 py-1 rounded-md bg-rose-500/20 border border-rose-500/40 text-rose-300 text-[10px] font-mono flex items-center gap-1.5 backdrop-blur pointer-events-none"
            >
              <ShieldAlert class="w-3 h-3 text-rose-400" />
              <span>OOM-Killed by Kernel</span>
            </div>

            <svg :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="w-full h-44 overflow-visible">
              <defs>
                <linearGradient id="memGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#818cf8" stop-opacity="0.35" />
                  <stop offset="100%" stop-color="#818cf8" stop-opacity="0.0" />
                </linearGradient>
              </defs>

              <line :x1="padding.left" :y1="padding.top" :x2="chartWidth - padding.right" :y2="padding.top" stroke="#1e293b" stroke-dasharray="3 3" />
              <line :x1="padding.left" :y1="chartHeight - padding.bottom" :x2="chartWidth - padding.right" :y2="chartHeight - padding.bottom" stroke="#334155" />

              <line :x1="padding.left" :y1="memoryChartPaths.warningY" :x2="chartWidth - padding.right" :y2="memoryChartPaths.warningY" stroke="#f97316" stroke-width="1.2" stroke-dasharray="4 3" stroke-opacity="0.6" />
              <line :x1="padding.left" :y1="memoryChartPaths.limitY" :x2="chartWidth - padding.right" :y2="memoryChartPaths.limitY" stroke="#ef4444" stroke-width="1.8" stroke-dasharray="5 3" />
              
              <text :x="chartWidth - padding.right - 4" :y="memoryChartPaths.limitY - 4" fill="#ef4444" font-size="9" text-anchor="end" font-family="monospace">
                memory.max: {{ maxMemMb }}MB (OOM Trigger)
              </text>

              <text :x="padding.left - 6" :y="padding.top + 4" fill="#64748b" font-size="9" text-anchor="end" font-family="monospace">
                {{ Math.round((memoryChartPaths.maxBytes || 0) / (1024 * 1024)) }}MB
              </text>
              <text :x="padding.left - 6" :y="chartHeight - padding.bottom" fill="#64748b" font-size="9" text-anchor="end" font-family="monospace">
                0
              </text>

              <path v-if="memoryChartPaths.area" :d="memoryChartPaths.area" fill="url(#memGradient)" />
              <path v-if="memoryChartPaths.line" :d="memoryChartPaths.line" fill="none" stroke="#818cf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />

              <g v-if="memoryChartPaths.coords">
                <circle v-for="(pt, idx) in memoryChartPaths.coords" :key="idx" :cx="pt.x" :cy="pt.y" r="2" fill="#818cf8" />
              </g>
            </svg>
          </div>

          <!-- Bottom Telemetry Stats -->
          <div class="grid grid-cols-3 gap-3 mt-4 pt-4 border-t border-slate-800 text-xs font-mono">
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">Usage</div>
              <div class="text-white font-bold text-sm mt-0.5">{{ currentMemMb }} MB</div>
            </div>
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">Utilization</div>
              <div :class="memUsagePercent > 85 ? 'text-rose-400 font-bold' : 'text-slate-300'" class="text-sm font-bold mt-0.5">
                {{ memUsagePercent }}%
              </div>
            </div>
            <div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">oom_kill events</div>
              <div :class="currentContainer.cgroups.oomKillEvents > 0 ? 'text-rose-400 font-bold' : 'text-slate-400'" class="text-sm font-bold mt-0.5">
                {{ currentContainer.cgroups.oomKillEvents }}
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Memory Allocation Stress Action -->
        <div class="mt-5 pt-4 border-t border-slate-800 flex items-center justify-between">
          <span class="text-xs text-slate-400">Step up RSS to test OOM-Killer:</span>
          <div class="flex items-center gap-2">
            <button 
              @click="handleStressMemory(16)"
              :disabled="currentContainer.status !== 'running' || isAllocatingMem"
              class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-indigo-600/80 hover:bg-indigo-600 text-white transition-all cursor-pointer disabled:opacity-50 shadow-sm"
            >
              +16MB
            </button>
            <button 
              @click="handleStressMemory(32)"
              :disabled="currentContainer.status !== 'running' || isAllocatingMem"
              class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-rose-600 hover:bg-rose-500 text-white transition-all cursor-pointer disabled:opacity-50 shadow-sm"
            >
              +32MB (OOM test)
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- LIVE CGROUPS V2 TUNING WORKBENCH & KERNEL PSEUDO-FILES -->
    <div v-if="currentContainer" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Cgroups Tuner Controls -->
      <div class="bg-slate-900/80 border border-slate-800 rounded-xl p-5 lg:col-span-1 space-y-4 shadow-lg">
        <div class="flex items-center gap-2.5 border-b border-slate-800 pb-3">
          <Sliders class="w-4 h-4 text-sky-400" />
          <h3 class="text-sm font-bold text-white tracking-tight">Dynamic cgroups v2 Tuner</h3>
        </div>
        <p class="text-xs text-slate-400">
          Modify the resource parameters in real time. The engine writes these changes directly into the cgroup hierarchy.
        </p>

        <!-- CPU Quota Slider -->
        <div class="space-y-2 pt-1">
          <div class="flex items-center justify-between text-xs font-mono">
            <span class="text-slate-300">CPU CFS Quota:</span>
            <span class="text-cyan-400 font-bold">{{ targetCpuCores.toFixed(2) }} cores ({{ (targetCpuCores * 100).toFixed(0) }}%)</span>
          </div>
          <input 
            type="range" 
            v-model.number="targetCpuCores" 
            min="0.1" 
            max="2.0" 
            step="0.05"
            class="w-full accent-cyan-400 cursor-pointer h-2 bg-slate-950 rounded-lg appearance-none border border-slate-800"
          />
          <div class="flex justify-between text-[10px] text-slate-500 font-mono">
            <span>0.10 (10%)</span>
            <span>1.0 (Core)</span>
            <span>2.0 (Multi)</span>
          </div>
        </div>

        <!-- Memory Limit Slider -->
        <div class="space-y-2 pt-2">
          <div class="flex items-center justify-between text-xs font-mono">
            <span class="text-slate-300">Memory Ceiling (memory.max):</span>
            <span class="text-indigo-400 font-bold">{{ targetMemMb }} MB</span>
          </div>
          <input 
            type="range" 
            v-model.number="targetMemMb" 
            min="16" 
            max="512" 
            step="16"
            class="w-full accent-indigo-400 cursor-pointer h-2 bg-slate-950 rounded-lg appearance-none border border-slate-800"
          />
          <div class="flex justify-between text-[10px] text-slate-500 font-mono">
            <span>16 MB</span>
            <span>128 MB</span>
            <span>512 MB</span>
          </div>
        </div>

        <button 
          @click="handleApplyCgroups"
          :disabled="isUpdatingCgroups"
          class="w-full py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-semibold flex items-center justify-center gap-2 transition-all cursor-pointer disabled:opacity-50 mt-4 shadow-lg shadow-blue-950/40"
        >
          <CheckCircle2 class="w-4 h-4" />
          {{ isUpdatingCgroups ? 'Updating Kernel Cgroup...' : 'Apply Cgroup Limits' }}
        </button>
      </div>

      <!-- Real-Time Kernel File Inspector -->
      <div class="bg-slate-900/80 border border-slate-800 rounded-xl p-5 lg:col-span-2 space-y-4 shadow-lg">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <div class="flex items-center gap-2.5">
            <FileCode2 class="w-4 h-4 text-emerald-400" />
            <h3 class="text-sm font-bold text-white tracking-tight">Simulated /sys/fs/cgroup Hierarchy</h3>
          </div>
          <span class="text-[11px] font-mono text-slate-500 bg-slate-950 px-2.5 py-1 rounded-md border border-slate-800">
            cgroups v2 unified mode
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5 text-xs font-mono">
          <!-- cpu.max -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl p-3">
            <div class="text-slate-400 font-semibold mb-1.5 flex items-center justify-between">
              <span>cpu.max</span>
              <span class="text-[10px] text-slate-500">quota period</span>
            </div>
            <div class="text-cyan-300 bg-slate-900 p-2 rounded-lg text-[11px] border border-slate-800">
              {{ currentContainer.cgroups.cpuQuotaUs }} {{ currentContainer.cgroups.cpuPeriodUs }}
            </div>
            <div class="text-[10px] text-slate-500 mt-1.5">
              Limits CFS execution time to {{ (currentContainer.cgroups.cpuQuotaUs / 1000).toFixed(0) }}ms every {{ (currentContainer.cgroups.cpuPeriodUs / 1000).toFixed(0) }}ms.
            </div>
          </div>

          <!-- memory.max -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl p-3">
            <div class="text-slate-400 font-semibold mb-1.5 flex items-center justify-between">
              <span>memory.max</span>
              <span class="text-[10px] text-slate-500">bytes</span>
            </div>
            <div class="text-indigo-300 bg-slate-900 p-2 rounded-lg text-[11px] border border-slate-800">
              {{ currentContainer.cgroups.memoryMaxBytes }}
            </div>
            <div class="text-[10px] text-slate-500 mt-1.5">
              Hard limit: {{ Math.round(currentContainer.cgroups.memoryMaxBytes / (1024 * 1024)) }}MB (Exceeding triggers kernel OOM-Killer).
            </div>
          </div>

          <!-- cpu.stat -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl p-3">
            <div class="text-slate-400 font-semibold mb-1.5 flex items-center justify-between">
              <span>cpu.stat</span>
              <span class="text-[10px] text-slate-500">accounting</span>
            </div>
            <div class="text-slate-300 bg-slate-900 p-2 rounded-lg text-[10px] leading-relaxed space-y-1 border border-slate-800">
              <div>nr_periods {{ Math.floor(Date.now() / 10000) % 500 }}</div>
              <div class="text-amber-400 font-medium">nr_throttled {{ currentContainer.cgroups.throttlePeriods }}</div>
              <div class="text-amber-300 font-medium">throttled_usec {{ currentContainer.cgroups.throttledTimeUs }}</div>
            </div>
          </div>

          <!-- memory.events -->
          <div class="bg-slate-950 border border-slate-800/80 rounded-xl p-3">
            <div class="text-slate-400 font-semibold mb-1.5 flex items-center justify-between">
              <span>memory.events</span>
              <span class="text-[10px] text-slate-500">faults & kills</span>
            </div>
            <div class="text-slate-300 bg-slate-900 p-2 rounded-lg text-[10px] leading-relaxed space-y-1 border border-slate-800">
              <div>low 0</div>
              <div>high {{ memUsagePercent > 80 ? 1 : 0 }}</div>
              <div>max {{ currentContainer.cgroups.oomKillEvents > 0 ? 1 : 0 }}</div>
              <div :class="currentContainer.cgroups.oomKillEvents > 0 ? 'text-rose-400 font-bold' : ''">
                oom_kill {{ currentContainer.cgroups.oomKillEvents }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ALL CONTAINERS LIVE FLEET MONITOR -->
    <div class="space-y-3.5">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-bold text-white tracking-tight flex items-center gap-2">
          <span>All Containers Resource Fleet</span>
          <span class="text-xs font-mono text-slate-500 font-normal">({{ containers.length }} total)</span>
        </h3>
        <span class="text-xs text-slate-400 font-mono">Click card to switch primary graphs</span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div 
          v-for="c in containers" 
          :key="c.id"
          @click="selectedContainerId = c.id"
          class="bg-slate-900/80 border rounded-xl p-4 cursor-pointer transition-all hover:border-blue-500/50 hover:shadow-lg select-none group"
          :class="c.id === selectedContainerId ? 'border-blue-500 ring-2 ring-blue-500/20 bg-slate-900 shadow-lg' : 'border-slate-800'"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2.5 truncate">
              <span 
                class="w-2.5 h-2.5 rounded-full shrink-0 shadow-sm"
                :class="c.status === 'running' ? 'bg-emerald-400 animate-pulse' : c.status === 'oom_killed' ? 'bg-rose-400' : 'bg-slate-500'"
              ></span>
              <span class="font-bold text-sm text-white font-mono truncate group-hover:text-blue-400 transition-colors">{{ c.name }}</span>
            </div>
            <span class="text-[10px] font-mono text-slate-400 px-2 py-0.5 rounded bg-slate-950 border border-slate-800">
              {{ c.status === 'running' ? 'Active' : c.status }}
            </span>
          </div>

          <div class="text-xs text-slate-400 font-mono space-y-2">
            <!-- CPU Progress -->
            <div>
              <div class="flex justify-between text-[11px] mb-1">
                <span>CPU: <strong class="text-slate-200">{{ c.status === 'running' ? c.cgroups.cpuPercent : 0 }}%</strong></span>
                <span class="text-slate-500">Cap: {{ ((c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs) * 100).toFixed(0) }}%</span>
              </div>
              <div class="w-full h-1.5 bg-slate-950 rounded-full overflow-hidden border border-slate-800/60">
                <div 
                  class="h-full bg-cyan-400 transition-all duration-300 rounded-full"
                  :style="{ width: `${Math.min(100, (c.status === 'running' ? c.cgroups.cpuPercent : 0))}%` }"
                ></div>
              </div>
            </div>

            <!-- RAM Progress -->
            <div>
              <div class="flex justify-between text-[11px] mb-1">
                <span>RAM: <strong class="text-slate-200">{{ Math.round(c.cgroups.memoryUsageBytes / (1024 * 1024)) }}MB</strong></span>
                <span class="text-slate-500">Max: {{ Math.round(c.cgroups.memoryMaxBytes / (1024 * 1024)) }}MB</span>
              </div>
              <div class="w-full h-1.5 bg-slate-950 rounded-full overflow-hidden border border-slate-800/60">
                <div 
                  class="h-full transition-all duration-300 rounded-full"
                  :class="c.cgroups.memoryUsageBytes > (c.cgroups.memoryMaxBytes * 0.85) ? 'bg-rose-500' : 'bg-indigo-400'"
                  :style="{ width: `${Math.min(100, (c.cgroups.memoryUsageBytes / c.cgroups.memoryMaxBytes) * 100)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>