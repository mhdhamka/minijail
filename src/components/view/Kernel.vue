<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import type { KernelPrimitiveGuide, Container } from '../../types/container.ts';
import { fetchKernelPrimitives } from '../../api.ts';
import { 
  Shield, 
  Code, 
  Check, 
  Copy, 
  HardDrive, 
  Network,
  Workflow,
  ShieldCheck,
  Boxes,
  Layers3,
  Layers,
  ChevronDown
} from 'lucide-vue-next';
import DockerLogo from '../common/DockerLogo.vue';
import Namespace from '../kernel/Namespace.vue';
import OverlayFs from '../kernel/OverlayFs.vue';
import Lifecycle from '../kernel/Lifecycle.vue';
import SecurityBox from '../kernel/SecurityBox.vue';
import PacketTracer from '../kernel/PacketTracer.vue';
import ComposeTopo from '../kernel/ComposeTopo.vue';
import UnionFsStack from '../kernel/UnionFsStack.vue';

interface Props {
  containers?: Container[];
}

const props = withDefaults(defineProps<Props>(), {
  containers: () => [],
});

const primitives = ref<KernelPrimitiveGuide[]>([]);
const activeIndex = ref(0);
const copied = ref(false);
const dropdownOpen = ref(false);

export type KernelViewMode = 'unionfs' | 'namespaces' | 'lifecycle' | 'overlayfs' | 'security' | 'network' | 'compose' | 'code';
const viewMode = ref<KernelViewMode>('unionfs');

const viewLabels: Record<KernelViewMode, { label: string; icon: any; color: string }> = {
  unionfs: { label: 'UnionFS Stack', icon: Layers3, color: 'text-amber-400' },
  namespaces: { label: 'Namespaces', icon: Layers, color: 'text-blue-400' },
  lifecycle: { label: 'Lifecycle Stepper', icon: Workflow, color: 'text-cyan-400' },
  overlayfs: { label: 'OverlayFS CoW', icon: HardDrive, color: 'text-amber-400' },
  security: { label: 'Security & Seccomp', icon: ShieldCheck, color: 'text-emerald-400' },
  network: { label: 'Packet Tracer', icon: Network, color: 'text-indigo-400' },
  compose: { label: 'Compose Topology', icon: Boxes, color: 'text-pink-400' },
  code: { label: 'Go Syscalls', icon: Code, color: 'text-slate-300' },
};

function selectView(mode: KernelViewMode) {
  viewMode.value = mode;
  dropdownOpen.value = false;
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement;
  if (!target.closest('#simulator-dropdown')) {
    dropdownOpen.value = false;
  }
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside);
  try {
    primitives.value = await fetchKernelPrimitives();
  } catch (e) {
    console.error(e);
  }
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});

function copyCode(code: string) {
  navigator.clipboard.writeText(code);
  copied.value = true;
  setTimeout(() => { copied.value = false; }, 2000);
}
</script>

<template>
  <div class="space-y-4 select-none">
    <!-- Header with Right-Aligned Dropdown Switcher -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#21262d] pb-4">
      <div>
        <h2 class="text-xl font-bold text-white flex items-center gap-2 font-sans">
          <span>Kernel</span>
          <span class="text-xs px-2 py-0.5 rounded-full bg-[#1E2633] text-blue-400 border border-[#2B3545] font-mono">
            Visualizer Suite
          </span>
        </h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Explore how Linux kernel primitives assemble container runtimes interactively.
        </p>
      </div>

      <!-- Compact Dropdown Switcher -->
      <div id="simulator-dropdown" class="relative">
        <button 
          @click="dropdownOpen = !dropdownOpen"
          class="px-3 py-2 rounded-xl bg-[#161b22] hover:bg-[#1f242c] border border-[#21262d] hover:border-[#30363d] transition-all cursor-pointer flex items-center gap-2.5 text-xs font-mono text-slate-200 shadow-sm"
        >
          <component :is="viewLabels[viewMode].icon" class="w-3.5 h-3.5" :class="viewLabels[viewMode].color" />
          <span class="font-semibold">{{ viewLabels[viewMode].label }}</span>
          <ChevronDown class="w-3.5 h-3.5 text-slate-400 transition-transform duration-200" :class="{ 'rotate-180': dropdownOpen }" />
        </button>

        <!-- Dropdown Menu -->
        <transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="transform scale-95 opacity-0"
          enter-to-class="transform scale-100 opacity-100"
          leave-active-class="transition duration-75 ease-in"
          leave-from-class="transform scale-100 opacity-100"
          leave-to-class="transform scale-95 opacity-0"
        >
          <div 
            v-if="dropdownOpen" 
            class="absolute right-0 mt-2 w-56 rounded-xl bg-[#161b22] border border-[#30363d] shadow-xl py-1.5 z-50 font-mono text-xs"
          >
            <div class="px-3 py-1 text-[10px] uppercase font-bold text-slate-500 tracking-wider">
              Simulators
            </div>
            <button
              v-for="(meta, mode) in viewLabels"
              :key="mode"
              @click="selectView(mode as KernelViewMode)"
              class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors cursor-pointer"
              :class="viewMode === mode ? 'bg-blue-500/10 text-blue-400 font-semibold' : 'text-slate-300 hover:bg-[#1f242c] hover:text-slate-100'"
            >
              <component :is="meta.icon" class="w-3.5 h-3.5" :class="meta.color" />
              <span>{{ meta.label }}</span>
            </button>
          </div>
        </transition>
      </div>
    </div>

    <!-- Minimalist Concept Banner -->
    <div class="p-3 rounded-xl bg-[#161b22]/50 border border-[#21262d] flex items-center gap-3 text-xs">
      <div class="text-blue-400 shrink-0">
        <DockerLogo :size="20" />
      </div>
      <div class="text-slate-400 leading-relaxed font-sans">
        <strong class="text-slate-200">Containers are standard Linux processes:</strong> Isolated using namespaces, resource-bounded by cgroups, and layered through union filesystems.
      </div>
    </div>

    <!-- Active View Component Renderer -->
    <UnionFsStack v-if="viewMode === 'unionfs'" />
    <Namespace v-else-if="viewMode === 'namespaces'" :containers="containers" />
    <Lifecycle v-else-if="viewMode === 'lifecycle'" />
    <OverlayFs v-else-if="viewMode === 'overlayfs'" />
    <SecurityBox v-else-if="viewMode === 'security'" />
    <PacketTracer v-else-if="viewMode === 'network'" />
    <ComposeTopo v-else-if="viewMode === 'compose'" />

    <!-- Go Primitives & Code View -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-12 gap-4">
      <div class="lg:col-span-4 space-y-1.5">
        <div 
          v-for="(item, idx) in primitives"
          :key="idx"
          @click="activeIndex = idx"
          class="p-2.5 rounded-xl border transition-all cursor-pointer font-mono text-xs"
          :class="activeIndex === idx 
            ? 'bg-[#1f242c] border-blue-500/50 text-slate-100 shadow-sm' 
            : 'bg-[#161b22]/60 border-[#21262d] text-slate-400 hover:bg-[#161b22] hover:text-slate-200'"
        >
          <div class="flex items-center justify-between">
            <span class="font-semibold text-slate-200">
              {{ item.name }}
            </span>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-[#0d1117] text-slate-400 border border-[#21262d]">
              {{ item.LinuxFlag || item.linuxFlag }}
            </span>
          </div>
        </div>
      </div>

      <div class="lg:col-span-8 rounded-xl bg-[#161b22]/40 border border-[#21262d] p-4 space-y-3 font-mono text-xs" v-if="primitives[activeIndex]">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-sm font-bold text-slate-100 flex items-center gap-2">
              <Shield class="w-4 h-4 text-blue-400" />
              <span>{{ primitives[activeIndex].name }}</span>
            </h3>
            <p class="text-slate-400 text-xs mt-1 font-sans leading-relaxed">
              {{ primitives[activeIndex].explanation }}
            </p>
          </div>
          <span class="px-2 py-0.5 rounded bg-[#0d1117] text-blue-400 border border-[#21262d] font-bold text-[10px] shrink-0">
            {{ primitives[activeIndex].LinuxFlag || primitives[activeIndex].linuxFlag }}
          </span>
        </div>

        <div class="p-2.5 rounded-lg bg-[#0d1117] border border-[#21262d] flex items-center justify-between text-[11px]">
          <span class="text-slate-500">Syscall Location:</span>
          <span class="text-emerald-400 font-bold">{{ primitives[activeIndex].SyscallFile || primitives[activeIndex].syscallFile }}</span>
        </div>

        <div class="rounded-xl border border-[#21262d] bg-[#090d12] overflow-hidden">
          <div class="px-3 py-2 bg-[#11161d] border-b border-[#21262d] flex items-center justify-between">
            <span class="text-[11px] text-slate-400 flex items-center gap-1.5">
              <Code class="w-3.5 h-3.5 text-blue-400" />
              <span>minijail-go implementation</span>
            </span>
            <button 
              @click="copyCode(primitives[activeIndex].GoCode || primitives[activeIndex].goCode)"
              class="px-2 py-1 rounded text-[10px] text-slate-300 hover:text-white bg-[#161b22] hover:bg-[#21262d] border border-[#30363d] transition-colors flex items-center gap-1 cursor-pointer"
            >
              <Check v-if="copied" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3 text-slate-400" />
              <span>{{ copied ? 'Copied' : 'Copy' }}</span>
            </button>
          </div>
          <pre class="p-3 text-cyan-200/90 text-xs overflow-x-auto leading-relaxed"><code>{{ primitives[activeIndex].GoCode || primitives[activeIndex].goCode }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>