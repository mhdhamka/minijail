<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Container } from '../types/container';
import { stressMemory, stressCpu } from '../api';
import { 
  X, 
  Activity, 
  HardDrive, 
  Cpu, 
  Skull, 
  Zap, 
  Flame, 
  AlertTriangle,
  FileCode,
  Layers
} from 'lucide-vue-next';

interface Props {
  container: Container;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'updated', container: Container): void;
}>();

const loading = ref(false);
const alertMessage = ref<string | null>(null);

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
  return (props.container.cgroups.cpuQuotaUs / props.container.cgroups.cpuPeriodUs).toFixed(2);
});

async function handleStressMem(deltaMb: number) {
  loading.value = true;
  alertMessage.value = null;
  try {
    const res = await stressMemory(props.container.id, deltaMb);
    emit('updated', res.container);
    if (res.oomKilled) {
      alertMessage.value = `[KERNEL OOM-KILLER TRIGGERED] Process exceeded ${memLimitMB.value}MB memory.max! Kernel sent SIGKILL (Exit code 137). Container halted.`;
    } else {
      alertMessage.value = `Allocated +${deltaMb}MB. Current usage: ${(res.container.cgroups.memoryUsageBytes / (1024*1024)).toFixed(1)}MB / ${memLimitMB.value}MB.`;
    }
  } catch (err: any) {
    alertMessage.value = `Error: ${err.message}`;
  } finally {
    loading.value = false;
  }
}

async function handleForceOom() {
  const remainingBytes = props.container.cgroups.memoryMaxBytes - props.container.cgroups.memoryUsageBytes;
  const deltaMb = Math.max(Math.ceil(remainingBytes / (1024 * 1024)) + 10, 30);
  await handleStressMem(deltaMb);
}

async function handleStressCpu() {
  loading.value = true;
  alertMessage.value = null;
  try {
    const updated = await stressCpu(props.container.id);
    emit('updated', updated);
    alertMessage.value = `CPU stressed to quota limit! CFS scheduler throttled process. Total throttled periods: ${updated.cgroups.throttlePeriods}.`;
  } catch (err: any) {
    alertMessage.value = `Error: ${err.message}`;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-sm">
    <div class="bg-slate-900 border border-slate-700/80 rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="p-5 border-b border-slate-800 flex items-center justify-between bg-slate-950/50">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-amber-950/80 border border-amber-800/80 text-amber-400">
            <Activity class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white flex items-center gap-2 font-mono">
              <span>cgroups v2 Controller: {{ container.name }}</span>
            </h2>
            <p class="text-xs text-slate-400 font-mono">
              Mounted at <span class="text-amber-400">{{ container.cgroups.cgroupPath }}</span>
            </p>
          </div>
        </div>
        <button 
          @click="emit('close')"
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Alert Notification Banner -->
      <div v-if="alertMessage" class="p-3 bg-rose-950/80 border-b border-rose-800/80 text-rose-200 text-xs font-mono flex items-start gap-2">
        <AlertTriangle class="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
        <div>{{ alertMessage }}</div>
      </div>

      <!-- Content Body -->
      <div class="p-5 overflow-y-auto space-y-6 text-xs font-mono">
        <!-- Memory cgroup section -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-slate-300 font-bold flex items-center gap-1.5 text-sm">
              <HardDrive class="w-4 h-4 text-cyan-400" />
              <span>Memory Controller (memory.max)</span>
            </span>
            <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-300">
              OOM Kill Count: <strong class="text-rose-400">{{ container.cgroups.oomKillEvents }}</strong>
            </span>
          </div>

          <!-- Progress bar -->
          <div>
            <div class="flex justify-between text-xs mb-1">
              <span class="text-slate-400">Usage vs Hard Limit:</span>
              <span :class="memPercent > 80 ? 'text-rose-400 font-bold' : 'text-slate-200'">
                {{ memUsedMB }} MB / {{ memLimitMB }} MB ({{ memPercent }}%)
              </span>
            </div>
            <div class="w-full h-3 rounded-full bg-slate-800 overflow-hidden">
              <div 
                class="h-full rounded-full transition-all duration-300"
                :class="memPercent > 85 ? 'bg-rose-500' : memPercent > 60 ? 'bg-amber-500' : 'bg-cyan-500'"
                :style="{ width: `${memPercent}%` }"
              ></div>
            </div>
          </div>

          <!-- Memory stress testing buttons -->
          <div class="pt-2 border-t border-slate-800/80 flex items-center gap-2 flex-wrap">
            <span class="text-[11px] text-slate-400">Stress Allocator:</span>
            <button
              @click="handleStressMem(16)"
              :disabled="loading || container.status !== 'running'"
              class="px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 text-cyan-300 border border-slate-700 transition-colors cursor-pointer text-xs flex items-center gap-1"
            >
              <Zap class="w-3.5 h-3.5" />
              <span>+16 MB RSS</span>
            </button>

            <button
              @click="handleStressMem(32)"
              :disabled="loading || container.status !== 'running'"
              class="px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 text-cyan-300 border border-slate-700 transition-colors cursor-pointer text-xs flex items-center gap-1"
            >
              <Zap class="w-3.5 h-3.5" />
              <span>+32 MB RSS</span>
            </button>

            <button
              @click="handleForceOom"
              :disabled="loading || container.status !== 'running'"
              class="px-3 py-1.5 rounded-lg bg-rose-950 hover:bg-rose-900 disabled:opacity-40 text-rose-300 border border-rose-800 transition-colors cursor-pointer text-xs font-bold flex items-center gap-1.5 shadow-md shadow-rose-950"
              title="Force container process to exceed memory.max and trigger kernel OOM Killer"
            >
              <Skull class="w-4 h-4 text-rose-400" />
              <span>Trigger Kernel OOM-Killer</span>
            </button>
          </div>
        </div>

        <!-- CPU cgroup section -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-slate-300 font-bold flex items-center gap-1.5 text-sm">
              <Cpu class="w-4 h-4 text-indigo-400" />
              <span>CFS Bandwidth Control (cpu.max)</span>
            </span>
            <span class="text-xs px-2 py-0.5 rounded bg-slate-800 text-indigo-300">
              Quota: {{ container.cgroups.cpuQuotaUs }}us / {{ container.cgroups.cpuPeriodUs }}us
            </span>
          </div>

          <div class="grid grid-cols-2 gap-3 text-xs">
            <div class="p-2.5 rounded-lg bg-slate-900 border border-slate-800">
              <div class="text-[10px] text-slate-500">Allocated CPU Core Share</div>
              <div class="text-sm font-bold text-slate-100 mt-0.5">{{ cpuQuotaCores }} Cores</div>
              <div class="text-[10px] text-slate-400">CFS quota = {{ (parseFloat(cpuQuotaCores) * 100).toFixed(0) }}%</div>
            </div>

            <div class="p-2.5 rounded-lg bg-slate-900 border border-slate-800">
              <div class="text-[10px] text-slate-500">CFS Throttled Periods</div>
              <div class="text-sm font-bold text-amber-400 mt-0.5">{{ container.cgroups.throttlePeriods }} events</div>
              <div class="text-[10px] text-slate-400">{{ (container.cgroups.throttledTimeUs / 1000).toFixed(1) }}ms delayed</div>
            </div>
          </div>

          <div class="pt-2 border-t border-slate-800/80 flex items-center justify-between">
            <span class="text-[11px] text-slate-400">CFS Scheduler Stress:</span>
            <button
              @click="handleStressCpu"
              :disabled="loading || container.status !== 'running'"
              class="px-3 py-1.5 rounded-lg bg-indigo-950 hover:bg-indigo-900 disabled:opacity-40 text-indigo-200 border border-indigo-800 transition-colors cursor-pointer text-xs flex items-center gap-1.5"
            >
              <Flame class="w-3.5 h-3.5 text-indigo-400" />
              <span>Stress CPU (Trigger CFS Throttle)</span>
            </button>
          </div>
        </div>

        <!-- Kernel Cgroup v2 pseudo-filesystem view -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-2">
          <div class="flex items-center gap-1.5 text-xs text-slate-300 font-bold">
            <FileCode class="w-4 h-4 text-cyan-400" />
            <span>Under The Hood: Linux /sys/fs/cgroup Entries</span>
          </div>

          <div class="space-y-1 text-[11px] text-slate-400 bg-slate-950 p-3 rounded-lg border border-slate-800/80">
            <div><strong class="text-cyan-400">cgroup.procs:</strong> {{ container.hostPid > 0 ? container.hostPid : '(idle)' }}</div>
            <div><strong class="text-cyan-400">memory.max:</strong> {{ container.cgroups.memoryMaxBytes }} bytes ({{ memLimitMB }} MB)</div>
            <div><strong class="text-cyan-400">memory.current:</strong> {{ container.cgroups.memoryUsageBytes }} bytes ({{ memUsedMB }} MB)</div>
            <div><strong class="text-cyan-400">cpu.max:</strong> {{ container.cgroups.cpuQuotaUs }} {{ container.cgroups.cpuPeriodUs }}</div>
            <div><strong class="text-cyan-400">cpu.stat:</strong> nr_periods=120 nr_throttled={{ container.cgroups.throttlePeriods }} throttled_usec={{ container.cgroups.throttledTimeUs }}</div>
            <div><strong class="text-cyan-400">memory.events:</strong> low=0 high=0 max={{ container.cgroups.oomKillEvents }} oom={{ container.cgroups.oomKillEvents }} oom_kill={{ container.cgroups.oomKillEvents }}</div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-800 bg-slate-950/60 flex justify-end">
        <button 
          @click="emit('close')"
          class="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 bg-slate-800 hover:bg-slate-700 transition-colors cursor-pointer"
        >
          Close Monitor
        </button>
      </div>
    </div>
  </div>
</template>
