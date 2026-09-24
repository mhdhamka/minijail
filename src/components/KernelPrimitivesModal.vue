<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { KernelPrimitiveGuide } from '../types/container';
import { fetchKernelPrimitives } from '../api';
import { X, Layers, Code, Shield, Check, Copy } from 'lucide-vue-next';

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const primitives = ref<KernelPrimitiveGuide[]>([]);
const activeTab = ref(0);
const copied = ref(false);

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
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-sm">
    <div class="bg-slate-900 border border-slate-700/80 rounded-2xl w-full max-w-4xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="p-5 border-b border-slate-800 flex items-center justify-between bg-slate-950/50">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-cyan-950/80 border border-cyan-800/80 text-cyan-400">
            <Layers class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white font-mono">
              Docker Kernel Primitives in Go
            </h2>
            <p class="text-xs text-slate-400 font-mono">
              How Linux Namespaces, cgroups v2, and pivot_root work under the hood
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

      <!-- Tab selector -->
      <div class="flex items-center gap-1 px-5 pt-3 border-b border-slate-800 overflow-x-auto bg-slate-950/30">
        <button
          v-for="(item, idx) in primitives"
          :key="idx"
          @click="activeTab = idx"
          class="px-3.5 py-2 rounded-t-lg text-xs font-mono font-medium transition-all shrink-0 cursor-pointer border-t border-x"
          :class="activeTab === idx 
            ? 'bg-slate-900 text-cyan-300 border-slate-700 border-b-transparent' 
            : 'text-slate-400 hover:text-slate-200 border-transparent'"
        >
          {{ item.name.split(' ')[0] }}
        </button>
      </div>

      <!-- Body: Primitive Detail -->
      <div v-if="primitives[activeTab]" class="p-6 overflow-y-auto space-y-5 text-xs font-mono">
        <div>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <Shield class="w-4 h-4 text-cyan-400" />
              <span>{{ primitives[activeTab].name }}</span>
            </h3>
            <span class="px-2 py-1 rounded bg-slate-800 text-cyan-300 border border-slate-700 font-semibold text-[11px]">
              {{ primitives[activeTab].LinuxFlag || primitives[activeTab].linuxFlag }}
            </span>
          </div>
          <p class="text-slate-300 text-xs mt-2 leading-relaxed font-sans">
            {{ primitives[activeTab].explanation }}
          </p>
        </div>

        <div class="p-3 rounded-lg bg-slate-950/50 border border-slate-800 flex items-center justify-between text-[11px]">
          <span class="text-slate-400">Kernel Syscall / ProcFS Location:</span>
          <span class="text-emerald-400 font-bold">{{ primitives[activeTab].SyscallFile || primitives[activeTab].syscallFile }}</span>
        </div>

        <!-- Golang Implementation Code Snippet -->
        <div class="rounded-xl border border-slate-800 bg-slate-950 overflow-hidden">
          <div class="px-4 py-2 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
            <span class="text-[11px] text-slate-400 flex items-center gap-1.5">
              <Code class="w-3.5 h-3.5 text-cyan-400" />
              <span>Golang Implementation (minijail runtime)</span>
            </span>
            <button 
              @click="copyCode(primitives[activeTab].GoCode || primitives[activeTab].goCode)"
              class="px-2 py-1 rounded text-[10px] text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 transition-colors flex items-center gap-1 cursor-pointer"
            >
              <Check v-if="copied" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3" />
              <span>{{ copied ? 'Copied!' : 'Copy Code' }}</span>
            </button>
          </div>
          <pre class="p-4 text-cyan-200 text-xs overflow-x-auto leading-relaxed"><code>{{ primitives[activeTab].GoCode || primitives[activeTab].goCode }}</code></pre>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-800 bg-slate-950/60 flex justify-end">
        <button 
          @click="emit('close')"
          class="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 bg-slate-800 hover:bg-slate-700 transition-colors cursor-pointer"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
