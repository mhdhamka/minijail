<script setup lang="ts">
import { ref, watch } from 'vue';
import type { CreateContainerPayload } from '../types/container';
import { X, Box, Sliders, Shield, Cpu, HardDrive, Sparkles } from 'lucide-vue-next';
import DockerLogo from './common/DockerLogo.vue';

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

const enablePidNs = ref(true);
const enableUtsNs = ref(true);
const enableNetNs = ref(true);
const enableMntNs = ref(true);

watch(() => props.initialImage, (newImg) => {
  if (newImg) image.value = newImg;
});

function applyPreset(presetType: string) {
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
    memoryLimitMb.value = 32; // Low memory to easily demo OOM-killer!
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
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm select-none">
    <div class="bg-[#161B22] border border-[#2D3848] rounded-xl w-full max-w-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh] text-slate-200">
      <!-- Header (Docker Desktop style) -->
      <div class="px-5 py-3.5 border-b border-[#232A35] flex items-center justify-between bg-[#11161D]">
        <div class="flex items-center gap-3">
          <DockerLogo :size="24" />
          <div>
            <h2 class="text-sm font-bold text-white">Run a new container</h2>
            <p class="text-xs text-slate-400 font-mono">Linux namespaces isolation & cgroups v2 resource limits</p>
          </div>
        </div>
        <button 
          @click="emit('close')"
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-[#202836] transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Presets Banner -->
      <div class="px-5 py-2.5 bg-[#0E1217] border-b border-[#232A35] flex items-center gap-2 flex-wrap text-xs">
        <span class="text-slate-400 flex items-center gap-1 font-medium text-[11px]">
          <Sparkles class="w-3.5 h-3.5 text-[#0db7ed]" /> Quick Presets:
        </span>
        <button 
          type="button" 
          @click="applyPreset('alpine-web')"
          class="px-2.5 py-1 rounded-md text-xs font-mono bg-[#1A222D] hover:bg-[#243040] text-cyan-300 border border-[#2D3848] transition-colors cursor-pointer"
        >
          Alpine Web (64MB)
        </button>
        <button 
          type="button" 
          @click="applyPreset('oom-canary')"
          class="px-2.5 py-1 rounded-md text-xs font-mono bg-rose-950/60 hover:bg-rose-900/80 text-rose-300 border border-rose-800/60 transition-colors cursor-pointer"
        >
          OOM Canary (32MB)
        </button>
        <button 
          type="button" 
          @click="applyPreset('busybox-worker')"
          class="px-2.5 py-1 rounded-md text-xs font-mono bg-[#1A222D] hover:bg-[#243040] text-slate-300 border border-[#2D3848] transition-colors cursor-pointer"
        >
          Busybox Worker (128MB)
        </button>
        <button 
          type="button" 
          @click="applyPreset('ubuntu-minimal')"
          class="px-2.5 py-1 rounded-md text-xs font-mono bg-[#1A222D] hover:bg-[#243040] text-slate-300 border border-[#2D3848] transition-colors cursor-pointer"
        >
          Ubuntu (256MB)
        </button>
      </div>

      <!-- Scrollable Form Body -->
      <div class="p-5 overflow-y-auto space-y-4 text-xs font-mono">
        <!-- Container Name & Base Image -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-slate-300 font-medium mb-1.5">Container Name</label>
            <input 
              v-model="name"
              type="text" 
              placeholder="e.g. web-gateway"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            />
          </div>
          <div>
            <label class="block text-slate-300 font-medium mb-1.5">Image</label>
            <select 
              v-model="image"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            >
              <option value="alpine:3.19">alpine:3.19 (7.38 MB)</option>
              <option value="busybox:1.36">busybox:1.36 (4.26 MB)</option>
              <option value="ubuntu:22.04">ubuntu:22.04 (77.8 MB)</option>
              <option value="nginx:alpine">nginx:alpine (23.4 MB)</option>
            </select>
          </div>
        </div>

        <!-- Command (Entrypoint) -->
        <div>
          <label class="block text-slate-300 font-medium mb-1.5">Entrypoint / Command</label>
          <input 
            v-model="command"
            type="text" 
            placeholder="sh -c 'echo Hello; sleep 10'"
            class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
          />
        </div>

        <!-- Cgroups v2 Hardware Limits -->
        <div class="p-4 rounded-xl bg-[#0E1217] border border-[#232A35] space-y-4">
          <div class="flex items-center gap-2 text-[#0db7ed] font-semibold">
            <Sliders class="w-4 h-4" />
            <span>Resource Limits (cgroups v2)</span>
          </div>

          <!-- Memory Limit -->
          <div>
            <div class="flex justify-between mb-1.5">
              <span class="text-slate-400 flex items-center gap-1">
                <HardDrive class="w-3.5 h-3.5 text-cyan-400" />
                memory.max
              </span>
              <span class="text-[#0db7ed] font-bold">{{ memoryLimitMb }} MB</span>
            </div>
            <input 
              v-model.number="memoryLimitMb"
              type="range" 
              min="16" 
              max="512" 
              step="16"
              class="w-full accent-[#1D63ED] cursor-pointer"
            />
            <div class="flex justify-between text-[10px] text-slate-500 mt-1">
              <span>16 MB (Ultra-restricted)</span>
              <span>128 MB</span>
              <span>512 MB</span>
            </div>
          </div>

          <!-- CPU CFS Quota -->
          <div>
            <div class="flex justify-between mb-1.5">
              <span class="text-slate-400 flex items-center gap-1">
                <Cpu class="w-3.5 h-3.5 text-indigo-400" />
                cpu.max (CFS Quota)
              </span>
              <span class="text-indigo-400 font-bold">{{ cpuQuotaCores.toFixed(2) }} Cores ({{ Math.round(cpuQuotaCores * 100) }}%)</span>
            </div>
            <input 
              v-model.number="cpuQuotaCores"
              type="range" 
              min="0.1" 
              max="2.0" 
              step="0.1"
              class="w-full accent-indigo-500 cursor-pointer"
            />
            <div class="flex justify-between text-[10px] text-slate-500 mt-1">
              <span>0.10 Cores (Throttled)</span>
              <span>1.00 Core</span>
              <span>2.00 Cores</span>
            </div>
          </div>
        </div>

        <!-- Linux Namespaces Cloned -->
        <div class="p-4 rounded-xl bg-[#0E1217] border border-[#232A35]">
          <div class="flex items-center gap-2 text-emerald-400 font-semibold mb-3">
            <Shield class="w-4 h-4" />
            <span>Linux Namespaces (CLONE flags)</span>
          </div>

          <div class="grid grid-cols-2 gap-3 text-xs">
            <label class="flex items-center gap-2 text-slate-300 cursor-pointer">
              <input type="checkbox" v-model="enablePidNs" class="rounded accent-[#1D63ED]" />
              <span>CLONE_NEWPID (PID 1)</span>
            </label>
            <label class="flex items-center gap-2 text-slate-300 cursor-pointer">
              <input type="checkbox" v-model="enableUtsNs" class="rounded accent-[#1D63ED]" />
              <span>CLONE_NEWUTS (Hostname)</span>
            </label>
            <label class="flex items-center gap-2 text-slate-300 cursor-pointer">
              <input type="checkbox" v-model="enableNetNs" class="rounded accent-[#1D63ED]" />
              <span>CLONE_NEWNET (Veth pair)</span>
            </label>
            <label class="flex items-center gap-2 text-slate-300 cursor-pointer">
              <input type="checkbox" v-model="enableMntNs" class="rounded accent-[#1D63ED]" />
              <span>CLONE_NEWNS (pivot_root)</span>
            </label>
          </div>
        </div>

        <!-- Optional: Network & Ports -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-slate-300 font-medium mb-1.5">Port Mapping (-p)</label>
            <input 
              v-model="portInput"
              type="text" 
              placeholder="8080:80/tcp"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            />
          </div>
          <div>
            <label class="block text-slate-300 font-medium mb-1.5">Environment Variables (-e)</label>
            <input 
              v-model="envInput"
              type="text" 
              placeholder="KEY=VALUE, ENV=production"
              class="w-full px-3 py-2 rounded-lg bg-[#0E1217] border border-[#2B3545] text-slate-100 text-xs focus:outline-none focus:border-[#1D63ED]"
            />
          </div>
        </div>
      </div>

      <!-- Footer Action -->
      <div class="px-5 py-3.5 border-t border-[#232A35] bg-[#11161D] flex items-center justify-end gap-3">
        <button 
          type="button" 
          @click="emit('close')"
          class="px-4 py-2 rounded-lg text-xs font-semibold text-slate-400 hover:text-white bg-[#1A222D] hover:bg-[#243040] transition-colors cursor-pointer"
        >
          Cancel
        </button>
        <button 
          type="button" 
          @click="handleDeploy"
          class="px-5 py-2 rounded-lg text-xs font-semibold text-white bg-[#1D63ED] hover:bg-[#1A57D0] shadow-sm transition-all cursor-pointer flex items-center gap-1.5"
        >
          <Box class="w-4 h-4" />
          <span>Run Container</span>
        </button>
      </div>
    </div>
  </div>
</template>
