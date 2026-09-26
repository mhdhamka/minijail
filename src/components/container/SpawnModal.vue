<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { CreateContainerPayload } from '../../types/container.ts';
import { X, Box, Sliders, Shield, Cpu, HardDrive, Terminal } from 'lucide-vue-next';
import DockerLogo from '../common/DockerLogo.vue';

interface Props {
  initialImage?: string;
}

const props = withDefaults(defineProps<Props>(), {
  initialImage: 'alpine:3.19'
});

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'deploy', payload: CreateContainerPayload): void;
}>();

const name = ref('');
const image = ref(props.initialImage || 'alpine:3.19');
const command = ref('sh -c "while true; do echo \\"[Container Alive] Handled packet\\"; sleep 3; done"');
const memoryLimitMb = ref(64);
const cpuQuotaCores = ref(0.5);
const hostname = ref('');
const portInput = ref('8080:80/tcp');
const envInput = ref('ENV=production');

const activePreset = ref<string>('alpine-web');

const enablePidNs = ref(true);
const enableUtsNs = ref(true);
const enableNetNs = ref(true);
const enableMntNs = ref(true);

watch(() => props.initialImage, (newImg) => {
  if (newImg) image.value = newImg;
});

// Computed live preview command for the mock terminal
const simulatedOutput = computed(() => {
  if (memoryLimitMb.value <= 32) return '[OOM-CANARY] WARNING: Low memory ceiling detected (32MB). GC pressure rising...';
  if (image.value.includes('nginx')) return `[nginx] 200 GET /index.html - Host: ${hostname.value || 'container-node'}`;
  return `[${image.value.split(':')[0]}] PID 1 active | cgroups v2 max: ${memoryLimitMb.value}MB`;
});

function applyPreset(presetType: string) {
  activePreset.value = presetType;
  if (presetType === 'alpine-web') {
    name.value = 'micro-web';
    image.value = 'alpine:3.19';
    command.value = 'sh -c "while true; do echo \\"[HTTP 200] GET /api/v1/health\\"; sleep 3; done"';
    memoryLimitMb.value = 64;
    cpuQuotaCores.value = 0.5;
    portInput.value = '8080:80/tcp';
  } else if (presetType === 'busybox-worker') {
    name.value = 'queue-worker';
    image.value = 'busybox:1.36';
    command.value = 'sh -c "while true; do echo \\"[Worker] Dequeued job payload\\"; sleep 5; done"';
    memoryLimitMb.value = 128;
    cpuQuotaCores.value = 1.0;
    portInput.value = '';
  } else if (presetType === 'oom-canary') {
    name.value = 'oom-canary';
    image.value = 'alpine:3.19';
    command.value = 'sh -c "while true; do echo \\"[Canary] Monitoring memory pressure\\"; sleep 4; done"';
    memoryLimitMb.value = 32;
    cpuQuotaCores.value = 0.25;
    portInput.value = '';
  } else if (presetType === 'ubuntu-minimal') {
    name.value = 'ubuntu-box';
    image.value = 'ubuntu:22.04';
    command.value = 'sh -c "while true; do echo \\"[Ubuntu Daemon] Active\\"; sleep 6; done"';
    memoryLimitMb.value = 256;
    cpuQuotaCores.value = 1.5;
    portInput.value = '';
  }
}

function handleDeploy() {
  const ports = portInput.value ? portInput.value.split(',').map(s => s.trim()).filter(Boolean) : [];
  const envs = envInput.value ? envInput.value.split(',').map(s => s.trim()).filter(Boolean) : [];

  emit('deploy', {
    name: name.value.trim() || `docker-${Date.now().toString(36).slice(-4)}`,
    image: image.value,
    command: command.value,
    memoryLimitMb: memoryLimitMb.value,
    cpuQuotaCores: cpuQuotaCores.value,
    enablePidNs: enablePidNs.value,
    enableUtsNs: enableUtsNs.value,
    enableNetNs: enableNetNs.value,
    enableMntNs: enableMntNs.value,
    hostname: hostname.value.trim() || name.value.trim(),
    portBindings: ports,
    env: envs,
  });
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md select-none animate-fadeIn">
    <div class="bg-[#0b0f17] border border-[#1e293b]/80 rounded-2xl w-full max-w-2xl overflow-hidden shadow-[0_0_50px_rgba(0,0,0,0.8)] flex flex-col max-h-[92vh] text-slate-200 ring-1 ring-white/10">
      
      <!-- Header -->
      <div class="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-gradient-to-r from-[#0f172a] via-[#0b0f17] to-[#0f172a]">
        <div class="flex items-center gap-3.5">
          <div class="relative p-2 rounded-xl bg-blue-500/10 border border-blue-500/20 shadow-inner">
            <DockerLogo :size="26" />
          </div>
          <div>
            <h2 class="text-sm font-bold tracking-wide text-white flex items-center gap-2">
              Run Container
            </h2>
            <p class="text-[11px] text-slate-400 font-mono">Linux namespaces isolation & cgroups v2 resource quotas</p>
          </div>
        </div>
        <button 
          @click="emit('close')"
          class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/5 transition-all duration-200 cursor-pointer border border-transparent hover:border-white/10"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Presets Interactive Cards Bar (Supercharged UI) -->
      <div class="px-5 py-3.5 bg-gradient-to-r from-[#070a0f] via-[#0d131f] to-[#070a0f] border-b border-white/10 flex flex-col gap-2">
        <div class="flex items-center justify-between">
          <span class="text-slate-400 flex items-center gap-1.5 font-bold text-[10px] uppercase tracking-wider">
            Presets:
          </span>
          <span class="text-[10px] font-mono text-slate-300">Click card to apply runtime profile</span>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
          
          <!-- Preset 1: Alpine Web -->
          <button 
            type="button" 
            @click="applyPreset('alpine-web')"
            :class="[
              'relative p-2.5 rounded-xl text-left font-mono transition-all duration-300 cursor-pointer border group overflow-hidden flex flex-col justify-between',
              activePreset === 'alpine-web' 
                ? 'bg-gradient-to-br from-cyan-950/50 via-[#111c2e] to-cyan-900/30 border-cyan-500/60 shadow-[0_0_20px_rgba(6,182,212,0.25)] ring-1 ring-cyan-500/40 scale-[1.02]' 
                : 'bg-[#111827]/70 border-white/5 hover:border-cyan-500/30 hover:bg-[#151f32] text-slate-400 hover:text-slate-200'
            ]"
          >
            <div class="absolute top-0 right-0 w-12 h-12 bg-cyan-500/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-cyan-500/10"></div>
            <div class="flex items-center justify-between mb-1">
              <span class="text-[11px] font-bold text-slate-200 flex items-center gap-1.5">
                <span :class="['w-2 h-2 rounded-full transition-all duration-300', activePreset === 'alpine-web' ? 'bg-cyan-400 shadow-[0_0_8px_#22d3ee] scale-125' : 'bg-slate-600']"></span>
                Alpine Web
              </span>
              <span v-if="activePreset === 'alpine-web'" class="text-[9px] bg-cyan-500/20 text-cyan-300 px-1.5 py-0.2 rounded font-bold animate-pulse">ACTIVE</span>
            </div>
            <div class="text-[9px] text-slate-400 flex items-center justify-between mt-1">
              <span>64MB RAM</span>
              <span class="text-cyan-400 font-semibold">0.5 CPU</span>
            </div>
          </button>

          <!-- Preset 2: OOM Canary -->
          <button 
            type="button" 
            @click="applyPreset('oom-canary')"
            :class="[
              'relative p-2.5 rounded-xl text-left font-mono transition-all duration-300 cursor-pointer border group overflow-hidden flex flex-col justify-between',
              activePreset === 'oom-canary' 
                ? 'bg-gradient-to-br from-rose-950/50 via-[#221318] to-rose-900/30 border-rose-500/60 shadow-[0_0_20px_rgba(244,63,94,0.25)] ring-1 ring-rose-500/40 scale-[1.02]' 
                : 'bg-[#111827]/70 border-white/5 hover:border-rose-500/30 hover:bg-[#20151a] text-slate-400 hover:text-slate-200'
            ]"
          >
            <div class="absolute top-0 right-0 w-12 h-12 bg-rose-500/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-rose-500/10"></div>
            <div class="flex items-center justify-between mb-1">
              <span class="text-[11px] font-bold text-slate-200 flex items-center gap-1.5">
                <span :class="['w-2 h-2 rounded-full transition-all duration-300', activePreset === 'oom-canary' ? 'bg-rose-400 shadow-[0_0_8px_#fb7185] scale-125' : 'bg-slate-600']"></span>
                OOM Canary
              </span>
              <span v-if="activePreset === 'oom-canary'" class="text-[9px] bg-rose-500/20 text-rose-300 px-1.5 py-0.2 rounded font-bold animate-pulse">ACTIVE</span>
            </div>
            <div class="text-[9px] text-slate-400 flex items-center justify-between mt-1">
              <span class="text-rose-300 font-bold">32MB RAM</span>
              <span class="text-rose-400 font-semibold">0.25 CPU</span>
            </div>
          </button>

          <!-- Preset 3: Worker Node -->
          <button 
            type="button" 
            @click="applyPreset('busybox-worker')"
            :class="[
              'relative p-2.5 rounded-xl text-left font-mono transition-all duration-300 cursor-pointer border group overflow-hidden flex flex-col justify-between',
              activePreset === 'busybox-worker' 
                ? 'bg-gradient-to-br from-amber-950/50 via-[#221c11] to-amber-900/30 border-amber-500/60 shadow-[0_0_20px_rgba(245,158,11,0.25)] ring-1 ring-amber-500/40 scale-[1.02]' 
                : 'bg-[#111827]/70 border-white/5 hover:border-amber-500/30 hover:bg-[#201c15] text-slate-400 hover:text-slate-200'
            ]"
          >
            <div class="absolute top-0 right-0 w-12 h-12 bg-amber-500/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-amber-500/10"></div>
            <div class="flex items-center justify-between mb-1">
              <span class="text-[11px] font-bold text-slate-200 flex items-center gap-1.5">
                <span :class="['w-2 h-2 rounded-full transition-all duration-300', activePreset === 'busybox-worker' ? 'bg-amber-400 shadow-[0_0_8px_#fbbf24] scale-125' : 'bg-slate-600']"></span>
                Worker Node
              </span>
              <span v-if="activePreset === 'busybox-worker'" class="text-[9px] bg-amber-500/20 text-amber-300 px-1.5 py-0.2 rounded font-bold animate-pulse">ACTIVE</span>
            </div>
            <div class="text-[9px] text-slate-400 flex items-center justify-between mt-1">
              <span>128MB RAM</span>
              <span class="text-amber-400 font-semibold">1.0 CPU</span>
            </div>
          </button>

          <!-- Preset 4: Ubuntu Base -->
          <button 
            type="button" 
            @click="applyPreset('ubuntu-minimal')"
            :class="[
              'relative p-2.5 rounded-xl text-left font-mono transition-all duration-300 cursor-pointer border group overflow-hidden flex flex-col justify-between',
              activePreset === 'ubuntu-minimal' 
                ? 'bg-gradient-to-br from-indigo-950/50 via-[#18182f] to-indigo-900/30 border-indigo-500/60 shadow-[0_0_20px_rgba(99,102,241,0.25)] ring-1 ring-indigo-500/40 scale-[1.02]' 
                : 'bg-[#111827]/70 border-white/5 hover:border-indigo-500/30 hover:bg-[#171732] text-slate-400 hover:text-slate-200'
            ]"
          >
            <div class="absolute top-0 right-0 w-12 h-12 bg-indigo-500/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-indigo-500/10"></div>
            <div class="flex items-center justify-between mb-1">
              <span class="text-[11px] font-bold text-slate-200 flex items-center gap-1.5">
                <span :class="['w-2 h-2 rounded-full transition-all duration-300', activePreset === 'ubuntu-minimal' ? 'bg-indigo-400 shadow-[0_0_8px_#818cf8] scale-125' : 'bg-slate-600']"></span>
                Ubuntu Base
              </span>
              <span v-if="activePreset === 'ubuntu-minimal'" class="text-[9px] bg-indigo-500/20 text-indigo-300 px-1.5 py-0.2 rounded font-bold animate-pulse">ACTIVE</span>
            </div>
            <div class="text-[9px] text-slate-400 flex items-center justify-between mt-1">
              <span>256MB RAM</span>
              <span class="text-indigo-400 font-semibold">1.5 CPU</span>
            </div>
          </button>

        </div>
      </div>

      <!-- Scrollable Form Body -->
      <div class="p-6 overflow-y-auto space-y-5 text-xs font-mono custom-scrollbar">
        
        <!-- Live Terminal Stream Preview Box -->
        <div class="p-3.5 rounded-xl bg-black/60 border border-white/10 relative overflow-hidden group">
          <div class="absolute top-0 left-0 w-1 h-full bg-blue-500"></div>
          <div class="flex items-center justify-between mb-1.5 text-[10px] text-slate-400 border-b border-white/5 pb-1">
            <span class="flex items-center gap-1.5 text-blue-400 font-bold">
              <Terminal class="w-3.5 h-3.5" /> LIVE STREAM SIMULATOR
            </span>
            <span class="flex items-center gap-1 text-emerald-400">
              TTY Ready
            </span>
          </div>
          <p class="text-[11px] text-cyan-300 font-mono tracking-tight truncate">
            $ docker run --name {{ name || 'pending...' }} -m {{ memoryLimitMb }}m --cpus={{ cpuQuotaCores }} {{ image }}
          </p>
          <p class="text-[10px] text-slate-400 mt-1 italic">
            > Output: {{ simulatedOutput }}
          </p>
        </div>

        <!-- Container Name & Base Image -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-slate-300 font-semibold mb-1.5 tracking-wide">Container Name</label>
            <div class="relative">
              <input 
                v-model="name"
                type="text" 
                placeholder="e.g. web-gateway-01"
                class="w-full px-3.5 py-2.5 rounded-xl bg-[#111827] border border-white/10 text-slate-100 text-xs focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
              />
            </div>
          </div>
          <div>
            <label class="block text-slate-300 font-semibold mb-1.5 tracking-wide">Base Image Engine</label>
            <select 
              v-model="image"
              class="w-full px-3.5 py-2.5 rounded-xl bg-[#111827] border border-white/10 text-slate-100 text-xs focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all cursor-pointer"
            >
              <option value="alpine:3.19">alpine:3.19 (Ultra-light 7.38 MB)</option>
              <option value="busybox:1.36">busybox:1.36 (Minimal 4.26 MB)</option>
              <option value="ubuntu:22.04">ubuntu:22.04 (Full GNU 77.8 MB)</option>
              <option value="nginx:alpine">nginx:alpine (Web Server 23.4 MB)</option>
            </select>
          </div>
        </div>

        <!-- Command (Entrypoint) -->
        <div>
          <label class="block text-slate-300 font-semibold mb-1.5 tracking-wide">Entrypoint / Exec Command</label>
          <input 
            v-model="command"
            type="text" 
            placeholder="sh -c 'echo Hello; sleep 10'"
            class="w-full px-3.5 py-2.5 rounded-xl bg-[#111827] border border-white/10 text-slate-100 text-xs focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
          />
        </div>

        <!-- Cgroups v2 Hardware Limits Panel -->
        <div class="p-4 rounded-xl bg-[#070a0f] border border-white/10 space-y-4 shadow-inner">
          <div class="flex items-center justify-between text-cyan-400 font-semibold border-b border-white/5 pb-2">
            <span class="flex items-center gap-2">
              <Sliders class="w-4 h-4 text-cyan-400 animate-pulse" />
              <span>cgroups v2 Hardware Isolation</span>
            </span>
            <span class="text-[10px] text-slate-500 font-mono bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20 text-cyan-300">Kernel cgroups v2</span>
          </div>

          <!-- Memory Limit Slider -->
          <div class="space-y-2">
            <div class="flex justify-between items-center text-xs">
              <span class="text-slate-300 flex items-center gap-1.5 font-medium">
                <HardDrive class="w-3.5 h-3.5 text-cyan-400" />
                memory.max <span class="text-slate-500 text-[10px]">(RAM Ceiling)</span>
              </span>
              <span class="text-cyan-300 font-bold font-mono bg-cyan-500/15 px-2.5 py-0.5 rounded-md border border-cyan-500/30 shadow-[0_0_10px_rgba(6,182,212,0.2)]">
                {{ memoryLimitMb }} MB
              </span>
            </div>
            <div class="relative flex items-center">
              <input 
                v-model.number="memoryLimitMb"
                type="range" 
                min="16" 
                max="512" 
                step="16"
                class="w-full accent-cyan-400 cursor-pointer bg-slate-800/80 rounded-lg h-2.5 transition-all focus:outline-none focus:ring-1 focus:ring-cyan-400/50"
              />
            </div>
            <div class="flex justify-between text-[10px] text-slate-500 font-mono">
              <span :class="memoryLimitMb <= 32 ? 'text-rose-400 font-bold' : ''">16 MB (Canary)</span>
              <span>256 MB</span>
              <span>512 MB (Max)</span>
            </div>
          </div>

          <!-- CPU CFS Quota Slider -->
          <div class="space-y-2 pt-2 border-t border-white/5">
            <div class="flex justify-between items-center text-xs">
              <span class="text-slate-300 flex items-center gap-1.5 font-medium">
                <Cpu class="w-3.5 h-3.5 text-indigo-400" />
                cpu.max <span class="text-slate-500 text-[10px]">(CFS Quota)</span>
              </span>
              <span class="text-indigo-300 font-bold font-mono bg-indigo-500/15 px-2.5 py-0.5 rounded-md border border-indigo-500/30 shadow-[0_0_10px_rgba(99,102,241,0.2)]">
                {{ cpuQuotaCores.toFixed(2) }} Cores ({{ Math.round(cpuQuotaCores * 100) }}%)
              </span>
            </div>
            <div class="relative flex items-center">
              <input 
                v-model.number="cpuQuotaCores"
                type="range" 
                min="0.1" 
                max="2.0" 
                step="0.1"
                class="w-full accent-indigo-400 cursor-pointer bg-slate-800/80 rounded-lg h-2.5 transition-all focus:outline-none focus:ring-1 focus:ring-indigo-400/50"
              />
            </div>
            <div class="flex justify-between text-[10px] text-slate-500 font-mono">
              <span>0.10 Cores</span>
              <span>1.00 Core</span>
              <span>2.00 Cores</span>
            </div>
          </div>
        </div>

        <!-- Linux Namespaces Cloned Panel -->
        <div class="p-4 rounded-xl bg-[#070a0f] border border-white/10 space-y-3">
          <div class="flex items-center justify-between text-emerald-400 font-semibold border-b border-white/5 pb-2">
            <span class="flex items-center gap-2">
              <Shield class="w-4 h-4 text-emerald-400 animate-pulse" />
              <span>Linux Namespaces (CLONE Flags)</span>
            </span>
            <span class="text-[10px] font-mono text-emerald-300/80 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
              {{ [enablePidNs, enableUtsNs, enableNetNs, enableMntNs].filter(Boolean).length }}/4 Active
            </span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5 text-xs">
            <!-- PID Checkbox Card -->
            <label :class="[
              'flex items-center justify-between p-2.5 rounded-xl border transition-all duration-200 cursor-pointer',
              enablePidNs ? 'bg-blue-600/10 border-blue-500/50 shadow-[0_0_12px_rgba(37,99,235,0.15)] text-blue-200' : 'bg-[#111827] border-white/5 text-slate-400 hover:border-white/15'
            ]">
              <div class="flex items-center gap-2.5">
                <input type="checkbox" v-model="enablePidNs" class="rounded accent-blue-500 w-4 h-4 cursor-pointer" />
                <span class="font-mono text-[11px] font-semibold">CLONE_NEWPID</span>
              </div>
              <span class="text-[10px] text-slate-500">PID 1 Sandbox</span>
            </label>

            <!-- UTS Checkbox Card -->
            <label :class="[
              'flex items-center justify-between p-2.5 rounded-xl border transition-all duration-200 cursor-pointer',
              enableUtsNs ? 'bg-blue-600/10 border-blue-500/50 shadow-[0_0_12px_rgba(37,99,235,0.15)] text-blue-200' : 'bg-[#111827] border-white/5 text-slate-400 hover:border-white/15'
            ]">
              <div class="flex items-center gap-2.5">
                <input type="checkbox" v-model="enableUtsNs" class="rounded accent-blue-500 w-4 h-4 cursor-pointer" />
                <span class="font-mono text-[11px] font-semibold">CLONE_NEWUTS</span>
              </div>
              <span class="text-[10px] text-slate-500">Hostname</span>
            </label>

            <!-- NET Checkbox Card -->
            <label :class="[
              'flex items-center justify-between p-2.5 rounded-xl border transition-all duration-200 cursor-pointer',
              enableNetNs ? 'bg-blue-600/10 border-blue-500/50 shadow-[0_0_12px_rgba(37,99,235,0.15)] text-blue-200' : 'bg-[#111827] border-white/5 text-slate-400 hover:border-white/15'
            ]">
              <div class="flex items-center gap-2.5">
                <input type="checkbox" v-model="enableNetNs" class="rounded accent-blue-500 w-4 h-4 cursor-pointer" />
                <span class="font-mono text-[11px] font-semibold">CLONE_NEWNET</span>
              </div>
              <span class="text-[10px] text-slate-500">Veth pair</span>
            </label>

            <!-- MNT Checkbox Card -->
            <label :class="[
              'flex items-center justify-between p-2.5 rounded-xl border transition-all duration-200 cursor-pointer',
              enableMntNs ? 'bg-blue-600/10 border-blue-500/50 shadow-[0_0_12px_rgba(37,99,235,0.15)] text-blue-200' : 'bg-[#111827] border-white/5 text-slate-400 hover:border-white/15'
            ]">
              <div class="flex items-center gap-2.5">
                <input type="checkbox" v-model="enableMntNs" class="rounded accent-blue-500 w-4 h-4 cursor-pointer" />
                <span class="font-mono text-[11px] font-semibold">CLONE_NEWNS</span>
              </div>
              <span class="text-[10px] text-slate-500">Mount table</span>
            </label>
          </div>
        </div>

        <!-- Optional: Network & Ports -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-slate-300 font-semibold mb-1.5 tracking-wide">Port Bindings (-p)</label>
            <input 
              v-model="portInput"
              type="text" 
              placeholder="8080:80/tcp"
              class="w-full px-3.5 py-2.5 rounded-xl bg-[#111827] border border-white/10 text-slate-100 text-xs focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
            />
          </div>
          <div>
            <label class="block text-slate-300 font-semibold mb-1.5 tracking-wide">Environment Envs (-e)</label>
            <input 
              v-model="envInput"
              type="text" 
              placeholder="KEY=VALUE, NODE_ENV=prod"
              class="w-full px-3.5 py-2.5 rounded-xl bg-[#111827] border border-white/10 text-slate-100 text-xs focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
            />
          </div>
        </div>
      </div>

      <!-- Footer Action -->
      <div class="px-6 py-4 border-t border-white/5 bg-[#070a0f] flex items-center justify-between">
        <div class="flex items-center gap-2 text-[11px] text-slate-400">
          <span>Ready to execute runtime daemon</span>
        </div>
        <div class="flex items-center gap-3">
          <button 
            type="button" 
            @click="emit('close')"
            class="px-4 py-2.5 rounded-xl text-xs font-semibold text-slate-400 hover:text-white bg-white/5 hover:bg-white/10 transition-all duration-200 cursor-pointer border border-white/5"
          >
            Cancel
          </button>
          <button 
            type="button" 
            @click="handleDeploy"
            class="px-6 py-2.5 rounded-xl text-xs font-bold text-white bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 shadow-[0_0_20px_rgba(37,99,235,0.4)] hover:shadow-[0_0_25px_rgba(37,99,235,0.6)] transition-all duration-200 cursor-pointer flex items-center gap-2 border border-blue-400/30 active:scale-95"
          >
            <Box class="w-4 h-4" />
            <span>Run Container</span>
          </button>
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.97); }
  to { opacity: 1; transform: scale(1); }
}
.animate-fadeIn {
  animation: fadeIn 0.18s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>