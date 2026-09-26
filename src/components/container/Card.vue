<script setup lang="ts">
import { computed } from 'vue';
import type { Container } from '../types/container';
import { 
  Play, 
  Square, 
  Pause, 
  Terminal, 
  Activity, 
  Trash2, 
  FolderTree, 
  Flame, 
  Skull,
  Shield,
  Network,
  ExternalLink,
  Code
} from 'lucide-vue-next';

interface Props {
  container: Container;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'start', id: string): void;
  (e: 'stop', id: string): void;
  (e: 'pause', id: string): void;
  (e: 'unpause', id: string): void;
  (e: 'kill', id: string): void;
  (e: 'delete', id: string): void;
  (e: 'open-details', container: Container, tab?: 'logs' | 'exec' | 'inspect' | 'files' | 'stats'): void;
}>();

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
</script>

<template>
  <div 
    :id="`container-card-${container.id.slice(0, 12)}`"
    class="rounded-xl bg-[#161B22] border transition-all duration-200 overflow-hidden flex flex-col justify-between hover:border-[#1D63ED]/60"
    :class="[
      container.status === 'running' ? 'border-[#232A35] shadow-md' : 
      container.status === 'oom_killed' ? 'border-rose-900 shadow-md shadow-rose-950/20' : 
      'border-[#232A35]'
    ]"
  >
    <!-- Top Bar: Name, Image, Status -->
    <div class="p-3.5 border-b border-[#232A35] bg-[#11161D] flex items-start justify-between gap-3">
      <div 
        @click="emit('open-details', container, 'logs')"
        class="cursor-pointer group flex-1"
      >
        <div class="flex items-center gap-2">
          <!-- Status Dot -->
          <span 
            class="w-2.5 h-2.5 rounded-full shrink-0"
            :class="[
              container.status === 'running' ? 'bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.8)] animate-pulse' :
              container.status === 'paused' ? 'bg-amber-500' :
              container.status === 'oom_killed' ? 'bg-rose-500 shadow-[0_0_6px_rgba(244,63,94,0.8)] animate-pulse' :
              'bg-slate-500'
            ]"
          ></span>
          <h3 class="text-sm font-bold text-white font-mono group-hover:text-[#0db7ed] transition-colors truncate">
            {{ container.name }}
          </h3>
          <span 
            class="text-[10px] font-mono px-1.5 py-0.2 rounded font-bold uppercase"
            :class="[
              container.status === 'running' ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/80' :
              container.status === 'paused' ? 'bg-amber-950/80 text-amber-400 border border-amber-800/80' :
              container.status === 'oom_killed' ? 'bg-rose-950 text-rose-300 border border-rose-800' :
              'bg-slate-800 text-slate-400 border border-slate-700'
            ]"
          >
            {{ container.status === 'oom_killed' ? 'OOM (137)' : container.status }}
          </span>
        </div>
        <div class="text-[11px] text-slate-400 font-mono flex items-center gap-2 mt-1">
          <span class="text-[#0db7ed] font-medium">{{ container.image }}</span>
          <span>•</span>
          <span class="text-slate-500">{{ container.id.slice(0, 8) }}</span>
          <span v-if="container.portBindings.length > 0">•</span>
          <span v-if="container.portBindings.length > 0" class="text-slate-300">{{ container.portBindings[0] }}</span>
        </div>
      </div>

      <!-- Quick lifecycle action buttons -->
      <div class="flex items-center gap-1">
        <template v-if="container.status === 'running'">
          <button 
            @click="emit('pause', container.id)"
            class="p-1.5 rounded-lg bg-[#202836] hover:bg-[#2B3648] text-amber-400 border border-[#2E3B4E] transition-colors cursor-pointer"
            title="Pause (cgroup.freeze = 1)"
          >
            <Pause class="w-3 h-3" />
          </button>
          <button 
            @click="emit('stop', container.id)"
            class="p-1.5 rounded-lg bg-[#202836] hover:bg-[#2B3648] text-slate-200 border border-[#2E3B4E] transition-colors cursor-pointer"
            title="Stop (SIGTERM)"
          >
            <Square class="w-3 h-3" />
          </button>
          <button 
            @click="emit('kill', container.id)"
            class="p-1.5 rounded-lg bg-rose-950/70 hover:bg-rose-900 text-rose-300 border border-rose-800/70 transition-colors cursor-pointer"
            title="Kill (SIGKILL exit 137)"
          >
            <Flame class="w-3 h-3" />
          </button>
        </template>
        <template v-else-if="container.status === 'paused'">
          <button 
            @click="emit('unpause', container.id)"
            class="p-1.5 rounded-lg bg-emerald-950 hover:bg-emerald-900 text-emerald-300 border border-emerald-800 transition-colors cursor-pointer"
            title="Resume"
          >
            <Play class="w-3 h-3" />
          </button>
          <button 
            @click="emit('stop', container.id)"
            class="p-1.5 rounded-lg bg-[#202836] hover:bg-[#2B3648] text-slate-200 border border-[#2E3B4E] transition-colors cursor-pointer"
            title="Stop"
          >
            <Square class="w-3 h-3" />
          </button>
        </template>
        <template v-else>
          <button 
            @click="emit('start', container.id)"
            class="p-1.5 rounded-lg bg-emerald-700 hover:bg-emerald-600 text-white transition-colors cursor-pointer"
            title="Start"
          >
            <Play class="w-3 h-3" />
          </button>
        </template>

        <button 
          @click="emit('delete', container.id)"
          class="p-1.5 rounded-lg bg-[#202836] hover:bg-rose-950/80 text-slate-400 hover:text-rose-400 border border-[#2E3B4E] hover:border-rose-800 transition-colors cursor-pointer"
          title="Delete container"
          :disabled="container.status === 'running'"
          :class="container.status === 'running' ? 'opacity-30 cursor-not-allowed' : ''"
        >
          <Trash2 class="w-3 h-3" />
        </button>
      </div>
    </div>

    <!-- Body: Kernel Telemetry & Resource Gauges -->
    <div class="p-3.5 space-y-3 font-mono text-xs">
      <!-- Command line display -->
      <div class="bg-[#0E1217] rounded-lg p-2 border border-[#232A35]">
        <div class="text-[9px] text-slate-500 uppercase tracking-wider mb-0.5 flex items-center justify-between">
          <span>Command</span>
          <span class="text-cyan-400">PID 1 (init)</span>
        </div>
        <div class="text-slate-300 text-[11px] truncate" :title="container.command">
          $ {{ container.command }}
        </div>
      </div>

      <!-- Namespace Isolation Indicators -->
      <div class="grid grid-cols-2 gap-2 text-xs">
        <div class="bg-[#0E1217] rounded-lg p-2 border border-[#232A35]">
          <div class="text-[9px] text-slate-500 flex items-center gap-1">
            <Shield class="w-3 h-3 text-cyan-400" />
            <span>Host PID</span>
          </div>
          <div class="text-xs font-bold text-slate-200 mt-0.5">
            {{ container.hostPid > 0 ? container.hostPid : '—' }}
          </div>
        </div>

        <div class="bg-[#0E1217] rounded-lg p-2 border border-[#232A35]">
          <div class="text-[9px] text-slate-500 flex items-center gap-1">
            <Shield class="w-3 h-3 text-emerald-400" />
            <span>Namespace PID</span>
          </div>
          <div class="text-xs font-bold text-emerald-400 mt-0.5">
            {{ container.containerPid > 0 ? container.containerPid : '—' }}
          </div>
        </div>
      </div>

      <!-- Cgroups Memory Gauge -->
      <div>
        <div class="flex items-center justify-between text-[11px] mb-1">
          <span class="text-slate-400 text-[10px] flex items-center gap-1">
            <span>memory.max</span>
            <span v-if="container.cgroups.oomKillEvents > 0" class="text-rose-400 font-bold flex items-center gap-0.5">
              <Skull class="w-3 h-3" /> OOM x{{ container.cgroups.oomKillEvents }}
            </span>
          </span>
          <span class="font-semibold" :class="memPercent > 80 ? 'text-rose-400' : 'text-slate-300'">
            {{ memUsedMB }} / {{ memLimitMB }} MB ({{ memPercent }}%)
          </span>
        </div>
        <div class="w-full h-1.5 rounded-full bg-[#0E1217] overflow-hidden border border-[#232A35]">
          <div 
            class="h-full rounded-full transition-all duration-300"
            :class="[
              memPercent > 85 ? 'bg-rose-500' : memPercent > 60 ? 'bg-amber-500' : 'bg-[#1D63ED]'
            ]"
            :style="{ width: `${memPercent}%` }"
          ></div>
        </div>
      </div>

      <!-- CPU cgroups -->
      <div>
        <div class="flex items-center justify-between text-[11px] mb-1">
          <span class="text-slate-400 text-[10px]">cpu.max ({{ cpuQuotaCores }} cores)</span>
          <span class="text-slate-300 font-semibold">{{ container.cgroups.cpuPercent.toFixed(1) }}%</span>
        </div>
        <div class="w-full h-1.5 rounded-full bg-[#0E1217] overflow-hidden border border-[#232A35]">
          <div 
            class="h-full rounded-full bg-indigo-500 transition-all duration-300"
            :style="{ width: `${Math.min(container.cgroups.cpuPercent, 100)}%` }"
          ></div>
        </div>
      </div>
    </div>

    <!-- Bottom Actions: Docker Desktop Tabs shortcuts -->
    <div class="p-2.5 border-t border-[#232A35] bg-[#11161D] grid grid-cols-4 gap-1.5 text-xs font-mono">
      <button
        @click="emit('open-details', container, 'logs')"
        class="py-1 rounded-md text-[11px] font-medium text-slate-300 bg-[#1A222D] hover:bg-[#243040] hover:text-white border border-[#263242] transition-all flex items-center justify-center gap-1 cursor-pointer"
        title="View logs"
      >
        <span>Logs</span>
      </button>

      <button
        @click="emit('open-details', container, 'exec')"
        class="py-1 rounded-md text-[11px] font-medium text-slate-300 bg-[#1A222D] hover:bg-[#243040] hover:text-emerald-300 border border-[#263242] transition-all flex items-center justify-center gap-1 cursor-pointer"
        title="Open container shell"
      >
        <span>Exec</span>
      </button>

      <button
        @click="emit('open-details', container, 'inspect')"
        class="py-1 rounded-md text-[11px] font-medium text-slate-300 bg-[#1A222D] hover:bg-[#243040] hover:text-amber-300 border border-[#263242] transition-all flex items-center justify-center gap-1 cursor-pointer"
        title="Inspect JSON"
      >
        <span>Inspect</span>
      </button>

      <button
        @click="emit('open-details', container, 'stats')"
        class="py-1 rounded-md text-[11px] font-medium text-slate-300 bg-[#1A222D] hover:bg-[#243040] hover:text-cyan-300 border border-[#263242] transition-all flex items-center justify-center gap-1 cursor-pointer"
        title="Cgroups & OOM lab"
      >
        <span>Stats</span>
      </button>
    </div>
  </div>
</template>
