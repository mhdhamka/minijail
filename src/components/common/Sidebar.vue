<script setup lang="ts">
import { 
  Box, 
  HardDrive, 
  Cpu, 
  Terminal, 
  ShieldCheck, 
  FolderTree, 
  Package, 
  Activity,
  Sliders,
  ChevronRight
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
    class="w-64 bg-[#0d1117] border-r border-[#21262d] flex flex-col justify-between shrink-0 h-screen sticky top-0 select-none text-slate-300 font-sans"
  >
    <!-- Top Branding -->
    <div>
      <div class="p-4 border-b border-[#21262d] flex items-center justify-between bg-[#0a0e14]/50">
        <div class="flex items-center gap-3">
          <div class="p-1 rounded-lg bg-[#161b22] border border-[#30363d] shadow-inner">
            <DockerLogo :size="26" />
          </div>
          <div>
            <div class="flex items-center gap-1.5">
              <span class="font-bold text-slate-100 tracking-tight text-sm">Docker</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded-md bg-blue-500/10 text-blue-400 font-semibold border border-blue-500/20">
                Desktop
              </span>
            </div>
            <div class="text-[11px] text-slate-400 font-mono mt-0.5 flex items-center gap-1">
              <span>Minijail-Go</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Navigation Menu -->
      <nav class="p-3 space-y-1 text-xs font-medium">
        <!-- Workbench -->
        <button
          id="nav-workbench"
          @click="emit('change-tab', 'workbench')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'workbench' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <Sliders class="w-4 h-4 transition-transform group-hover:scale-110" :class="activeTab === 'workbench' ? 'text-white' : 'text-slate-400'" />
            <span>Workbench</span>
          </div>
          <span 
            class="text-[10px] px-2 py-0.5 rounded-md font-mono font-medium transition-colors"
            :class="activeTab === 'workbench' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#161b22] text-blue-400 border border-[#30363d]'"
          >
            Ops + CLI
          </span>
        </button>

        <!-- Containers -->
        <button
          id="nav-containers"
          @click="emit('change-tab', 'containers')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'containers' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <Box class="w-4 h-4 transition-transform group-hover:scale-110" :class="activeTab === 'containers' ? 'text-white' : 'text-slate-400'" />
            <span>Containers</span>
          </div>
          <span 
            class="text-[11px] px-2 py-0.5 rounded-full font-mono font-bold transition-colors"
            :class="activeTab === 'containers' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#161b22] text-slate-300 border border-[#30363d]'"
          >
            {{ runningContainersCount }}/{{ totalContainersCount }}
          </span>
        </button>

        <!-- Images -->
        <button
          id="nav-images"
          @click="emit('change-tab', 'images')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'images' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <Package class="w-4 h-4 transition-transform group-hover:scale-110" :class="activeTab === 'images' ? 'text-white' : 'text-slate-400'" />
            <span>Images</span>
          </div>
          <span 
            class="text-[11px] px-2.5 py-0.5 rounded-full font-mono transition-colors"
            :class="activeTab === 'images' 
              ? 'bg-white/20 text-white' 
              : 'bg-[#161b22] text-slate-300 border border-[#30363d]'"
          >
            4
          </span>
        </button>

        <!-- Volumes -->
        <button
          id="nav-volumes"
          @click="emit('change-tab', 'volumes')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'volumes' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <FolderTree class="w-4 h-4 transition-transform group-hover:scale-110" :class="activeTab === 'volumes' ? 'text-white' : 'text-slate-400'" />
            <span>Volumes & Mounts</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono transition-colors"
            :class="activeTab === 'volumes' ? 'bg-white/20 text-white' : 'text-slate-500'"
          >
            OverlayFS
          </span>
        </button>

        <!-- Kernel / Under The Hood -->
        <button
          id="nav-kernel"
          @click="emit('change-tab', 'kernel')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'kernel' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <ShieldCheck class="w-4 h-4 transition-transform group-hover:scale-110" :class="activeTab === 'kernel' ? 'text-white' : 'text-slate-400'" />
            <span>Kernel</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono transition-colors"
            :class="activeTab === 'kernel' ? 'bg-white/20 text-white' : 'text-emerald-400 bg-emerald-950/40 border border-emerald-900/50'"
          >
            Go / Linux
          </span>
        </button>

        <!-- Resource Monitor -->
        <button
          id="nav-metrics"
          @click="emit('change-tab', 'metrics')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left"
          :class="activeTab === 'metrics' 
            ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white font-semibold shadow-lg shadow-blue-500/20' 
            : 'text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center gap-2.5">
            <Activity class="w-4 h-4 transition-transform group-hover:scale-110 text-cyan-400" />
            <span>Resource Monitor</span>
          </div>
          <span 
            class="text-[10px] px-1.5 py-0.5 rounded font-mono transition-colors"
            :class="activeTab === 'metrics' ? 'bg-white/20 text-white' : 'text-cyan-400 bg-cyan-950/40 border border-cyan-900/50'"
          >
            Live
          </span>
        </button>

        <div class="pt-4 pb-1 px-3 text-[10px] uppercase font-bold tracking-wider text-slate-500">
          Tools & Console
        </div>

        <!-- Docker CLI Launcher -->
        <button
          id="nav-cli"
          @click="emit('open-cli')"
          class="group w-full flex items-center justify-between px-3 py-2.5 rounded-xl transition-all duration-150 cursor-pointer text-left text-slate-400 hover:bg-[#161b22] hover:text-slate-200"
        >
          <div class="flex items-center gap-2.5">
            <Terminal class="w-4 h-4 text-emerald-400 transition-transform group-hover:scale-110" />
            <span class="font-mono">CLI Shell</span>
          </div>
          <ChevronRight class="w-3.5 h-3.5 text-slate-600 transition-transform group-hover:translate-x-0.5" />
        </button>
      </nav>
    </div>

    <!-- Bottom Engine Status & System Allocation -->
    <div class="p-3.5 border-t border-[#21262d] bg-[#090d12] space-y-2.5 text-xs">
      <!-- Engine Status Badge -->
      <div class="flex items-center justify-between px-2.5 py-2 rounded-xl bg-[#11161d] border border-[#21262d]">
        <div class="flex items-center gap-2">
          <span 
            class="w-2.5 h-2.5 rounded-full transition-all duration-300" 
            :class="connected ? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]' : 'bg-rose-500'"
          ></span>
          <span class="text-slate-200 font-medium text-[11px]">
            {{ connected ? 'Engine running' : 'Connecting...' }}
          </span>
        </div>
        <span class="text-[10px] text-slate-500 font-mono">v1.0.0</span>
      </div>

      <!-- Cgroups & Telemetry Cards -->
      <div class="grid grid-cols-2 gap-2 text-[11px] font-mono">
        <div class="px-2.5 py-2 rounded-xl bg-[#11161d] border border-[#21262d] transition-colors hover:border-slate-700">
          <div class="text-[10px] text-slate-500 flex items-center gap-1.5">
            <Cpu class="w-3 h-3 text-cyan-400" />
            <span>CPU</span>
          </div>
          <div class="text-slate-200 font-semibold mt-1">
            {{ totalCpuCores.toFixed(1) }} Cores
          </div>
        </div>
        <div class="px-2.5 py-2 rounded-xl bg-[#11161d] border border-[#21262d] transition-colors hover:border-slate-700">
          <div class="text-[10px] text-slate-500 flex items-center gap-1.5">
            <HardDrive class="w-3 h-3 text-indigo-400" />
            <span>Memory</span>
          </div>
          <div class="text-slate-200 font-semibold mt-1">
            {{ totalMemoryMB }} MB
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>