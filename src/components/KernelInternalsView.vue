<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { KernelPrimitiveGuide, Container } from '../types/container';
import { fetchKernelPrimitives } from '../api';
import { 
  Shield, 
  Code, 
  Check, 
  Copy, 
  Layers, 
  ExternalLink, 
  Cpu, 
  HardDrive, 
  FolderTree, 
  Network,
  Zap,
  ShieldCheck,
  Boxes,
  FileCode2,
  Layers3
} from 'lucide-vue-next';
import DockerLogo from './DockerLogo.vue';
import NamespaceDiagram from './NamespaceDiagram.vue';
import OverlayFsPlayground from './OverlayFsPlayground.vue';
import LifecycleMachine from './LifecycleMachine.vue';
import SecurityCapabilitiesSandbox from './SecurityCapabilitiesSandbox.vue';
import NetworkPacketTracer from './NetworkPacketTracer.vue';
import ComposeTopologyVisualizer from './ComposeTopologyVisualizer.vue';
import UnionFsStackDiagram from './UnionFsStackDiagram.vue';

interface Props {
  containers?: Container[];
}

const props = withDefaults(defineProps<Props>(), {
  containers: () => [],
});

const primitives = ref<KernelPrimitiveGuide[]>([]);
const activeIndex = ref(0);
const copied = ref(false);

export type KernelViewMode = 'unionfs' | 'namespaces' | 'lifecycle' | 'overlayfs' | 'security' | 'network' | 'compose' | 'code';
const viewMode = ref<KernelViewMode>('unionfs');

onMounted(async () => {
  try {
    primitives.value = await fetchKernelPrimitives();
  } catch (e) {
    console.error(e);
  }
});

function copyCode(code: string) {
  navigator.clipboard.writeText(code);
  copied.value = true;
  setTimeout(() => { copied.value = false; }, 2000);
}
</script>

<template>
  <div class="space-y-5 select-none">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 border-b border-[#232A35] pb-4">
      <div>
        <h2 class="text-xl font-bold text-white flex items-center gap-2 font-sans">
          <span>Under The Hood</span>
          <span class="text-xs px-2 py-0.5 rounded-full bg-[#1E2633] text-[#0db7ed] border border-[#2B3545] font-mono">
            Interactive Container Visualizer Suite
          </span>
        </h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Select any interactive simulator below to visually explore how Linux kernel primitives assemble Docker containers.
        </p>
      </div>

      <!-- Simulator Tabs Switcher -->
      <div class="flex flex-wrap items-center gap-1.5 p-1 bg-[#141A22] border border-[#232A35] rounded-xl font-mono text-xs">
        <button 
          @click="viewMode = 'unionfs'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'unionfs' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Layered Visual Stack Diagram (AUFS / Overlay2)"
        >
          <Layers3 class="w-3.5 h-3.5 text-amber-400" />
          <span>UnionFS Stack</span>
        </button>

        <button 
          @click="viewMode = 'namespaces'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'namespaces' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Namespaces Isolation Boundaries"
        >
          <Layers class="w-3.5 h-3.5" />
          <span>Namespaces</span>
        </button>

        <button 
          @click="viewMode = 'lifecycle'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'lifecycle' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Container Lifecycle Machine: from runc to app"
        >
          <Zap class="w-3.5 h-3.5 text-cyan-400" />
          <span>Lifecycle Stepper</span>
        </button>

        <button 
          @click="viewMode = 'overlayfs'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'overlayfs' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="OverlayFS Copy-on-Write Layering"
        >
          <HardDrive class="w-3.5 h-3.5 text-amber-400" />
          <span>OverlayFS CoW</span>
        </button>

        <button 
          @click="viewMode = 'security'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'security' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Linux Capabilities & Seccomp Sandbox"
        >
          <ShieldCheck class="w-3.5 h-3.5 text-emerald-400" />
          <span>Capabilities & Seccomp</span>
        </button>

        <button 
          @click="viewMode = 'network'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'network' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Virtual veth and iptables NAT Packet Tracer"
        >
          <Network class="w-3.5 h-3.5 text-indigo-400" />
          <span>Packet Tracer</span>
        </button>

        <button 
          @click="viewMode = 'compose'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'compose' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Compose Multi-Container Microservices Topology"
        >
          <Boxes class="w-3.5 h-3.5 text-pink-400" />
          <span>Compose Topology</span>
        </button>

        <button 
          @click="viewMode = 'code'"
          class="px-2.5 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="viewMode === 'code' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Underlying Go Engine Code"
        >
          <Code class="w-3.5 h-3.5 text-slate-300" />
          <span>Go Syscalls</span>
        </button>
      </div>
    </div>

    <!-- Educational Concept Banner -->
    <div class="p-3.5 rounded-xl bg-[#161B22] border border-[#232A35] flex items-start gap-3 text-xs">
      <div class="w-9 h-9 rounded-lg bg-[#1D63ED]/20 text-[#0db7ed] flex items-center justify-center shrink-0 border border-[#1D63ED]/30">
        <DockerLogo :size="24" />
      </div>
      <div>
        <div class="font-bold text-slate-200">The Secret: Containers are Just Ordinary Linux Processes!</div>
        <div class="text-slate-400 mt-0.5 leading-relaxed font-sans">
          Unlike Virtual Machines (which emulate physical hardware via hypervisors), a Docker container is simply a standard Linux process run with unshared namespaces (<code class="text-cyan-400">CLONE_NEWPID</code>, <code class="text-indigo-400">CLONE_NEWNET</code>, <code class="text-amber-400">CLONE_NEWNS</code>), bounded by cgroups (<code class="text-cyan-400">/sys/fs/cgroup/memory.max</code>), restricted by capabilities (<code class="text-emerald-400">capset</code>), and jailed in an AUFS/Overlay2 union filesystem root via <code class="text-cyan-400">pivot_root</code>.
        </div>
      </div>
    </div>

    <!-- VIEW 0: LAYERED VISUAL STACK DIAGRAM (AUFS / OVERLAY2) -->
    <UnionFsStackDiagram 
      v-if="viewMode === 'unionfs'" 
    />

    <!-- VIEW 1: NAMESPACE BOUNDARY DIAGRAM -->
    <NamespaceDiagram 
      v-else-if="viewMode === 'namespaces'" 
      :containers="containers" 
    />

    <!-- VIEW 2: STEP-BY-STEP LIFECYCLE MACHINE -->
    <LifecycleMachine 
      v-else-if="viewMode === 'lifecycle'" 
    />

    <!-- VIEW 3: OVERLAYFS LAYER STACKER (COPY-ON-WRITE PLAYGROUND) -->
    <OverlayFsPlayground 
      v-else-if="viewMode === 'overlayfs'" 
    />

    <!-- VIEW 4: LINUX CAPABILITIES & SECCOMP SANDBOX -->
    <SecurityCapabilitiesSandbox 
      v-else-if="viewMode === 'security'" 
    />

    <!-- VIEW 5: VIRTUAL NETWORKING & PACKET TRACER -->
    <NetworkPacketTracer 
      v-else-if="viewMode === 'network'" 
    />

    <!-- VIEW 6: COMPOSE MULTI-CONTAINER TOPOLOGY -->
    <ComposeTopologyVisualizer 
      v-else-if="viewMode === 'compose'" 
    />

    <!-- VIEW 7: GO PRIMITIVES & CODE -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-12 gap-4">
      <!-- Left Column: Primitive Items List -->
      <div class="lg:col-span-4 space-y-2">
        <div 
          v-for="(item, idx) in primitives"
          :key="idx"
          @click="activeIndex = idx"
          class="p-3 rounded-xl border transition-all cursor-pointer font-mono text-xs"
          :class="activeIndex === idx 
            ? 'bg-[#1D63ED]/15 border-[#1D63ED] text-white shadow-sm' 
            : 'bg-[#161B22] border-[#232A35] text-slate-400 hover:bg-[#1A222D] hover:text-slate-200'"
        >
          <div class="flex items-center justify-between">
            <span class="font-bold" :class="activeIndex === idx ? 'text-[#0db7ed]' : 'text-slate-300'">
              {{ item.name }}
            </span>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-[#0E1217] border border-[#2B3545]">
              {{ item.LinuxFlag || item.linuxFlag }}
            </span>
          </div>
          <div class="text-[11px] text-slate-400 mt-1 font-sans line-clamp-2">
            {{ item.explanation }}
          </div>
        </div>
      </div>

      <!-- Right Column: Detail & Go Code -->
      <div class="lg:col-span-8 rounded-xl bg-[#161B22] border border-[#232A35] p-5 space-y-4 font-mono text-xs" v-if="primitives[activeIndex]">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <Shield class="w-4 h-4 text-[#0db7ed]" />
              <span>{{ primitives[activeIndex].name }}</span>
            </h3>
            <p class="text-slate-300 text-xs mt-1.5 font-sans leading-relaxed">
              {{ primitives[activeIndex].explanation }}
            </p>
          </div>
          <span class="px-2.5 py-1 rounded bg-[#0E1217] text-[#0db7ed] border border-[#2B3545] font-bold text-[11px] shrink-0">
            {{ primitives[activeIndex].LinuxFlag || primitives[activeIndex].linuxFlag }}
          </span>
        </div>

        <div class="p-3 rounded-lg bg-[#0E1217] border border-[#232A35] flex items-center justify-between text-[11px]">
          <span class="text-slate-400">Syscall / Kernel Location:</span>
          <span class="text-emerald-400 font-bold">{{ primitives[activeIndex].SyscallFile || primitives[activeIndex].syscallFile }}</span>
        </div>

        <!-- Golang Implementation Code -->
        <div class="rounded-xl border border-[#232A35] bg-[#0A0D12] overflow-hidden">
          <div class="px-4 py-2 bg-[#11161D] border-b border-[#232A35] flex items-center justify-between">
            <span class="text-[11px] text-slate-400 flex items-center gap-1.5">
              <Code class="w-3.5 h-3.5 text-[#0db7ed]" />
              <span>Go Engine Implementation (minijail-go)</span>
            </span>
            <button 
              @click="copyCode(primitives[activeIndex].GoCode || primitives[activeIndex].goCode)"
              class="px-2 py-1 rounded text-[10px] text-slate-300 hover:text-white bg-[#1A222D] hover:bg-[#243040] border border-[#2D3848] transition-colors flex items-center gap-1 cursor-pointer"
            >
              <Check v-if="copied" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3 text-slate-400" />
              <span>{{ copied ? 'Copied' : 'Copy Code' }}</span>
            </button>
          </div>
          <pre class="p-4 text-cyan-200 text-xs overflow-x-auto leading-relaxed"><code>{{ primitives[activeIndex].GoCode || primitives[activeIndex].goCode }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>
