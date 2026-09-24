<script setup lang="ts">
import { 
  Box, 
  Layers, 
  HardDrive, 
  Cpu, 
  Terminal, 
  ShieldCheck, 
  HelpCircle, 
  Settings, 
  FolderTree, 
  Package, 
  Activity,
  CheckCircle2,
  ExternalLink,
  Sliders
} from 'lucide-vue-next';
import DockerLogo from './DockerLogo.vue';

export type ActiveTab = 'workbench' | 'containers' | 'images' | 'volumes' | 'kernel' | 'metrics' | 'cli';

interface Props {
  activeTab: ActiveTab;
  runningContainersCount: number;
  totalContainersCount: number;
  connected: boolean;
  totalMemoryMB: number;
  totalCpuCores: number;
}

defineProps<Props>();

const emit = defineEmits<{
  (e: 'change-tab', tab: ActiveTab): void;
  (e: 'open-cli'): void;
  (e: 'open-spawner'): void;
}>();
</script>

<template>
  <aside 
    id="docker-sidebar" 
    class="w-64 bg-[#11161D] border-r border-[#232A35] flex flex-col justify-between shrink-0 h-screen sticky top-0 select-none text-slate-300"
  >
    <!-- Top Branding -->
    <div>
      <div class="p-4 border-b border-[#232A35] flex items-center justify-between">
        <div class="flex items-center gap-3">
          <DockerLogo :size="30" />
          <div>
            <div class="flex items-center gap-1.5">
              <span class="font-bold text-white tracking-wide text-sm font-sans">Docker</span>
              <span class="text-xs px-1.5 py-0.2 rounded bg-[#1D63ED]/20 text-[#0db7ed] font-medium border border-[#1D63ED]/30">
                Desktop
              </span>
            </div>
            <div class="text-[11px] text-slate-400 font-mono">
              Engine: <span class="text-slate-300">Minijail-Go</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Navigation Menu (Exact Docker Desktop sidebar items) -->
      <nav class="p-3 space-y-1 text-xs font-medium">
        <!-- Workbench (Split Terminal & Ops) -->
        <button
          id="nav-workbench"
          @click="emit('change-tab', 'workbench')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'workbench' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <Sliders class="w-4 h-4" />
            <span>Workbench</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono font-medium"
            :class="activeTab === 'workbench' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#1E2633] text-[#0db7ed] border border-[#2D3848]'"
          >
            Terminal + Ops
          </span>
        </button>

        <!-- Containers -->
        <button
          id="nav-containers"
          @click="emit('change-tab', 'containers')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'containers' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <Box class="w-4 h-4" />
            <span>Containers</span>
          </div>
          <span 
            class="text-[11px] px-2 py-0.5 rounded-full font-mono font-bold"
            :class="activeTab === 'containers' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#1E2633] text-slate-400 border border-[#2D3848]'"
          >
            {{ runningContainersCount }}/{{ totalContainersCount }}
          </span>
        </button>

        <!-- Images -->
        <button
          id="nav-images"
          @click="emit('change-tab', 'images')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'images' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <Package class="w-4 h-4" />
            <span>Images</span>
          </div>
          <span 
            class="text-[11px] px-2 py-0.5 rounded-full font-mono"
            :class="activeTab === 'images' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#1E2633] text-slate-400 border border-[#2D3848]'"
          >
            4
          </span>
        </button>

        <!-- Volumes -->
        <button
          id="nav-volumes"
          @click="emit('change-tab', 'volumes')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'volumes' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <FolderTree class="w-4 h-4" />
            <span>Volumes & Mounts</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono"
            :class="activeTab === 'volumes' ? 'bg-white/20 text-white' : 'text-slate-500'"
          >
            OverlayFS
          </span>
        </button>

        <!-- Learning Center / Kernel Primitives -->
        <button
          id="nav-kernel"
          @click="emit('change-tab', 'kernel')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'kernel' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <ShieldCheck class="w-4 h-4" />
            <span>Under The Hood</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono"
            :class="activeTab === 'kernel' ? 'bg-white/20 text-white' : 'text-emerald-400 bg-emerald-950/60 border border-emerald-800/60'"
          >
            Go / Linux
          </span>
        </button>

        <!-- Resource Monitor (Real-Time Cgroups & Throttling) -->
        <button
          id="nav-metrics"
          @click="emit('change-tab', 'metrics')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left"
          :class="activeTab === 'metrics' 
            ? 'bg-[#1D63ED] text-white font-semibold shadow-sm' 
            : 'text-slate-300 hover:bg-[#1A222D] hover:text-white'"
        >
          <div class="flex items-center gap-2.5">
            <Activity class="w-4 h-4 text-cyan-400" />
            <span>Resource Monitor</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono"
            :class="activeTab === 'metrics' ? 'bg-white/20 text-white' : 'text-cyan-400 bg-cyan-950/60 border border-cyan-800/60'"
          >
            Live Charts
          </span>
        </button>

        <div class="pt-3 pb-1 px-3 text-[10px] uppercase font-bold tracking-wider text-slate-500">
          Tools & Console
        </div>

        <!-- Docker CLI -->
        <button
          id="nav-cli"
          @click="emit('open-cli')"
          class="w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-colors cursor-pointer text-left text-slate-300 hover:bg-[#1A222D] hover:text-white"
        >
          <div class="flex items-center gap-2.5">
            <Terminal class="w-4 h-4 text-emerald-400" />
            <span class="font-mono">docker cli</span>
          </div>
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-[#1E2633] text-slate-400 border border-[#2D3848]">
            CLI
          </span>
        </button>
      </nav>
    </div>

    <!-- Bottom Engine Status & System Allocation (Just like Docker Desktop) -->
    <div class="p-3 border-t border-[#232A35] bg-[#0E1218] space-y-2.5 text-xs">
      <!-- Engine status badge -->
      <div class="flex items-center justify-between px-2 py-1.5 rounded-md bg-[#161B22] border border-[#232A35]">
        <div class="flex items-center gap-2">
          <span 
            class="w-2.5 h-2.5 rounded-full" 
            :class="connected ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)] animate-pulse' : 'bg-rose-500'"
          ></span>
          <span class="text-slate-200 font-medium text-[11px]">
            {{ connected ? 'Engine running' : 'Engine starting...' }}
          </span>
        </div>
        <span class="text-[10px] text-slate-400 font-mono">v27.0.3</span>
      </div>

      <!-- Cgroups & CPU telemetry -->
      <div class="grid grid-cols-2 gap-1.5 text-[11px] font-mono">
        <div class="px-2 py-1.5 rounded bg-[#161B22] border border-[#232A35]">
          <div class="text-[10px] text-slate-500 flex items-center gap-1">
            <Cpu class="w-3 h-3 text-cyan-400" />
            <span>CPU</span>
          </div>
          <div class="text-slate-200 font-semibold mt-0.5">
            {{ totalCpuCores.toFixed(1) }} Cores
          </div>
        </div>
        <div class="px-2 py-1.5 rounded bg-[#161B22] border border-[#232A35]">
          <div class="text-[10px] text-slate-500 flex items-center gap-1">
            <HardDrive class="w-3 h-3 text-indigo-400" />
            <span>Memory</span>
          </div>
          <div class="text-slate-200 font-semibold mt-0.5">
            {{ totalMemoryMB }} MB
          </div>
        </div>
      </div>

      <!-- Docker Desktop Footer -->
      <div class="flex items-center justify-between text-[11px] text-slate-500 pt-1 px-1">
        <div class="flex items-center gap-2">
          <button 
            @click="emit('open-spawner')"
            class="text-[#0db7ed] hover:underline cursor-pointer flex items-center gap-0.5"
          >
            + Run
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>
