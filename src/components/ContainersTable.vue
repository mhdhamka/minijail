<script setup lang="ts">
import { computed } from 'vue';
import type { Container } from '../types/container';
import { 
  Play, 
  Square, 
  Pause, 
  Terminal, 
  Trash2, 
  Flame, 
  Box,
} from 'lucide-vue-next';

interface Props {
  containers: Container[];
  selectedIds: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'toggle-select', id: string): void;
  (e: 'select-all'): void;
  (e: 'open-details', container: Container, tab?: 'logs' | 'exec' | 'inspect' | 'files' | 'stats'): void;
  (e: 'start', id: string): void;
  (e: 'stop', id: string): void;
  (e: 'pause', id: string): void;
  (e: 'unpause', id: string): void;
  (e: 'kill', id: string): void;
  (e: 'delete', id: string): void;
}>();

const allSelected = computed(() => {
  return props.containers.length > 0 && props.selectedIds.length === props.containers.length;
});
</script>

<template>
  <div class="rounded-xl border border-[#232A35] bg-[#161B22] overflow-hidden select-none">
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs font-mono">
        <!-- Table Header (Exact Docker Desktop columns) -->
        <thead class="bg-[#11161D] border-b border-[#232A35] text-slate-400">
          <tr>
            <th class="w-10 px-3 py-3 text-center">
              <input 
                type="checkbox" 
                :checked="allSelected"
                @change="emit('select-all')"
                class="rounded bg-[#0E1217] border-[#2E3A4E] text-[#1D63ED] focus:ring-0 cursor-pointer"
              />
            </th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">Status</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">Name</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">Image</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">Port(s)</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">CPU</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px]">Memory</th>
            <th class="px-3 py-3 font-semibold uppercase tracking-wider text-[11px] text-right">Actions</th>
          </tr>
        </thead>

        <!-- Table Body -->
        <tbody class="divide-y divide-[#202733]">
          <tr 
            v-for="c in containers" 
            :key="c.id"
            class="hover:bg-[#1A222D] transition-colors group cursor-pointer"
          >
            <!-- Checkbox -->
            <td class="px-3 py-3 text-center" @click.stop>
              <input 
                type="checkbox" 
                :checked="selectedIds.includes(c.id)"
                @change="emit('toggle-select', c.id)"
                class="rounded bg-[#0E1217] border-[#2E3A4E] text-[#1D63ED] focus:ring-0 cursor-pointer"
              />
            </td>

            <!-- Status Indicator -->
            <td class="px-3 py-3" @click="emit('open-details', c, 'logs')">
              <div class="flex items-center gap-2">
                <span 
                  class="w-2.5 h-2.5 rounded-full shrink-0"
                  :class="[
                    c.status === 'running' ? 'bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.8)] animate-pulse' :
                    c.status === 'paused' ? 'bg-amber-500' :
                    c.status === 'oom_killed' ? 'bg-rose-500 shadow-[0_0_6px_rgba(244,63,94,0.8)] animate-pulse' :
                    'bg-slate-500'
                  ]"
                ></span>
                <span 
                  class="font-bold uppercase text-[11px]"
                  :class="[
                    c.status === 'running' ? 'text-emerald-400' :
                    c.status === 'paused' ? 'text-amber-400' :
                    c.status === 'oom_killed' ? 'text-rose-400' :
                    'text-slate-400'
                  ]"
                >
                  {{ c.status === 'oom_killed' ? 'OOM (137)' : c.status }}
                </span>
              </div>
            </td>

            <!-- Name -->
            <td class="px-3 py-3" @click="emit('open-details', c, 'logs')">
              <div class="flex items-center gap-2">
                <Box class="w-4 h-4 text-slate-400 group-hover:text-[#0db7ed] transition-colors shrink-0" />
                <div>
                  <div class="font-bold text-slate-100 group-hover:text-[#0db7ed] transition-colors flex items-center gap-1.5">
                    <span>{{ c.name }}</span>
                    <span class="text-[10px] text-slate-500 font-normal">({{ c.id.slice(0, 8) }})</span>
                  </div>
                  <div class="text-[10px] text-slate-500 truncate max-w-xs">$ {{ c.command }}</div>
                </div>
              </div>
            </td>

            <!-- Image -->
            <td class="px-3 py-3 text-slate-300" @click="emit('open-details', c, 'logs')">
              <span class="text-[#0db7ed] bg-[#0db7ed]/10 px-2 py-0.5 rounded border border-[#0db7ed]/20">
                {{ c.image }}
              </span>
            </td>

            <!-- Ports -->
            <td class="px-3 py-3 text-slate-300" @click="emit('open-details', c, 'logs')">
              <span v-if="c.portBindings.length > 0" class="text-cyan-300">
                {{ c.portBindings.join(', ') }}
              </span>
              <span v-else class="text-slate-600">—</span>
            </td>

            <!-- CPU -->
            <td class="px-3 py-3 text-slate-300" @click="emit('open-details', c, 'stats')">
              <div class="flex items-center gap-2">
                <span>{{ c.cgroups.cpuPercent.toFixed(1) }}%</span>
                <div class="w-12 h-1.5 rounded-full bg-[#0E1217] overflow-hidden border border-[#232A35]">
                  <div 
                    class="h-full bg-indigo-500" 
                    :style="{ width: `${Math.min(c.cgroups.cpuPercent, 100)}%` }"
                  ></div>
                </div>
              </div>
            </td>

            <!-- Memory -->
            <td class="px-3 py-3 text-slate-300" @click="emit('open-details', c, 'stats')">
              <div class="flex items-center gap-2">
                <span>
                  {{ (c.cgroups.memoryUsageBytes / (1024*1024)).toFixed(1) }} / {{ (c.cgroups.memoryMaxBytes / (1024*1024)).toFixed(0) }} MB
                </span>
                <div class="w-12 h-1.5 rounded-full bg-[#0E1217] overflow-hidden border border-[#232A35]">
                  <div 
                    class="h-full"
                    :class="[
                      (c.cgroups.memoryUsageBytes / c.cgroups.memoryMaxBytes) > 0.85 ? 'bg-rose-500' :
                      (c.cgroups.memoryUsageBytes / c.cgroups.memoryMaxBytes) > 0.6 ? 'bg-amber-500' :
                      'bg-[#1D63ED]'
                    ]"
                    :style="{ width: `${Math.min((c.cgroups.memoryUsageBytes / c.cgroups.memoryMaxBytes) * 100, 100)}%` }"
                  ></div>
                </div>
              </div>
            </td>

            <!-- Action buttons -->
            <td class="px-3 py-3 text-right" @click.stop>
              <div class="flex items-center justify-end gap-1">
                <!-- Lifecycle control -->
                <template v-if="c.status === 'running'">
                  <button 
                    @click="emit('pause', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-amber-400 transition-colors"
                    title="Pause"
                  >
                    <Pause class="w-3.5 h-3.5" />
                  </button>
                  <button 
                    @click="emit('stop', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-slate-300 hover:text-white transition-colors"
                    title="Stop"
                  >
                    <Square class="w-3.5 h-3.5" />
                  </button>
                  <button 
                    @click="emit('kill', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-rose-400 hover:text-rose-300 transition-colors"
                    title="Kill"
                  >
                    <Flame class="w-3.5 h-3.5" />
                  </button>
                </template>
                <template v-else-if="c.status === 'paused'">
                  <button 
                    @click="emit('unpause', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-emerald-400 transition-colors"
                    title="Resume"
                  >
                    <Play class="w-3.5 h-3.5" />
                  </button>
                  <button 
                    @click="emit('stop', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-slate-300 hover:text-white transition-colors"
                    title="Stop"
                  >
                    <Square class="w-3.5 h-3.5" />
                  </button>
                </template>
                <template v-else>
                  <button 
                    @click="emit('start', c.id)"
                    class="p-1.5 rounded hover:bg-[#202734] text-emerald-400 transition-colors"
                    title="Start"
                  >
                    <Play class="w-3.5 h-3.5" />
                  </button>
                </template>

                <!-- Quick Open Exec/Shell -->
                <button 
                  @click="emit('open-details', c, 'exec')"
                  class="p-1.5 rounded hover:bg-[#202734] text-emerald-400 hover:text-emerald-300 transition-colors"
                  title="Open Shell (CLI)"
                >
                  <Terminal class="w-3.5 h-3.5" />
                </button>

                <!-- Delete -->
                <button 
                  @click="emit('delete', c.id)"
                  :disabled="c.status === 'running'"
                  class="p-1.5 rounded hover:bg-[#202734] text-slate-500 hover:text-rose-400 disabled:opacity-30 disabled:hover:text-slate-500 transition-colors"
                  title="Delete"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-if="containers.length === 0">
            <td colspan="8" class="py-12 text-center text-slate-500">
              No containers match the current filter or search.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
