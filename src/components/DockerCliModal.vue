<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { executeCli } from '../api';
import { X, Terminal, Send, Sparkles, HelpCircle } from 'lucide-vue-next';
import DockerLogo from './DockerLogo.vue';

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
}>();

const inputCommand = ref('docker ps');
const loading = ref(false);
const history = ref<string[]>([]);
const historyIndex = ref(-1);
const terminalLines = ref<{ command: string; output: string; error?: string; time: string }[]>([]);
const terminalBody = ref<HTMLElement | null>(null);

onMounted(async () => {
  // Run initial docker ps
  await runCli('docker ps');
});

function scrollToBottom() {
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight;
    }
  });
}

async function runCli(cmdToRun?: string) {
  const target = (cmdToRun || inputCommand.value).trim();
  if (!target || loading.value) return;

  history.value.push(target);
  historyIndex.value = history.value.length;
  inputCommand.value = '';
  loading.value = true;
  scrollToBottom();

  const now = new Date().toTimeString().slice(0, 8);

  try {
    const res = await executeCli(target);
    terminalLines.value.push({
      command: target,
      output: res.output,
      error: res.error,
      time: now,
    });
    emit('refresh');
  } catch (err: any) {
    terminalLines.value.push({
      command: target,
      output: '',
      error: err.message,
      time: now,
    });
  } finally {
    loading.value = false;
    scrollToBottom();
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowUp') {
    if (historyIndex.value > 0) {
      historyIndex.value--;
      inputCommand.value = history.value[historyIndex.value];
    }
  } else if (e.key === 'ArrowDown') {
    if (historyIndex.value < history.value.length - 1) {
      historyIndex.value++;
      inputCommand.value = history.value[historyIndex.value];
    } else {
      historyIndex.value = history.value.length;
      inputCommand.value = '';
    }
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm select-none">
    <div class="bg-[#161B22] border border-[#2D3848] rounded-xl w-full max-w-3xl overflow-hidden shadow-2xl flex flex-col h-[85vh] text-slate-200">
      <!-- Header -->
      <div class="px-5 py-3.5 border-b border-[#232A35] flex items-center justify-between bg-[#11161D]">
        <div class="flex items-center gap-3">
          <DockerLogo :size="24" />
          <div>
            <div class="text-xs font-bold text-white font-mono flex items-center gap-2">
              <span>Docker CLI Console</span>
              <span class="text-[10px] px-1.5 py-0.2 rounded bg-[#1D63ED]/20 text-[#0db7ed] border border-[#1D63ED]/30">
                Minijail-Go Engine
              </span>
            </div>
            <div class="text-[11px] text-slate-400 font-mono">
              Direct Go engine orchestration: docker ps, run, stats, stop, exec, inspect
            </div>
          </div>
        </div>

        <button 
          @click="emit('close')"
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-[#202836] transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Quick Command Buttons -->
      <div class="px-4 py-2 bg-[#0E1217] border-b border-[#232A35] flex items-center gap-2 overflow-x-auto text-[11px] font-mono">
        <span class="text-slate-500 shrink-0 flex items-center gap-1 text-[10px]">
          <Sparkles class="w-3 h-3 text-[#0db7ed]" /> Quick CLI:
        </span>
        <button 
          @click="runCli('docker ps -a')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-slate-300 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker ps -a
        </button>
        <button 
          @click="runCli('docker stats')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-cyan-300 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker stats
        </button>
        <button 
          @click="runCli('docker run -d --name redis -m 64m alpine:3.19 sh')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-emerald-300 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker run ...
        </button>
        <button 
          @click="runCli('docker images')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-slate-300 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker images
        </button>
        <button 
          @click="runCli('docker version')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-indigo-300 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker version
        </button>
        <button 
          @click="runCli('docker --help')"
          class="px-2 py-0.5 rounded bg-[#1A222D] hover:bg-[#243040] text-slate-400 border border-[#2D3848] transition-colors shrink-0 cursor-pointer"
        >
          docker --help
        </button>
      </div>

      <!-- Output Screen -->
      <div 
        ref="terminalBody"
        class="flex-1 p-4 overflow-y-auto bg-[#0A0D12] font-mono text-xs space-y-4 selection:bg-[#1D63ED]/30"
      >
        <div v-for="(entry, idx) in terminalLines" :key="idx" class="space-y-1">
          <!-- Command Prompt -->
          <div class="flex items-center gap-2 text-slate-400">
            <span class="text-slate-600 text-[10px]">{{ entry.time }}</span>
            <span class="text-[#0db7ed] font-bold">$</span>
            <span class="text-white font-semibold">{{ entry.command }}</span>
          </div>

          <!-- Output -->
          <div v-if="entry.output" class="bg-[#11161D] p-3 rounded-lg border border-[#232A35] text-slate-200 whitespace-pre font-mono text-[11px] overflow-x-auto leading-relaxed">
{{ entry.output }}
          </div>

          <!-- Error -->
          <div v-if="entry.error" class="bg-rose-950/40 p-2.5 rounded-lg border border-rose-800/60 text-rose-300 whitespace-pre-wrap font-mono text-[11px]">
Error: {{ entry.error }}
          </div>
        </div>

        <div v-if="loading" class="flex items-center gap-2 text-[#0db7ed] pt-2 text-[11px]">
          <span class="w-1.5 h-1.5 rounded-full bg-[#0db7ed] animate-ping"></span>
          <span>Interpreting in Go CLI engine...</span>
        </div>
      </div>

      <!-- Command Input -->
      <div class="p-3 border-t border-[#232A35] bg-[#11161D]">
        <form @submit.prevent="runCli()" class="flex items-center gap-2 font-mono text-xs">
          <span class="text-[#0db7ed] font-bold shrink-0 select-none">$</span>
          <input 
            v-model="inputCommand"
            @keydown="handleKeydown"
            type="text" 
            placeholder="Type docker command (e.g. docker ps -a, docker stats, docker run -d --name app -m 64m alpine:3.19 sh)..."
            class="flex-1 bg-transparent text-slate-100 placeholder:text-slate-600 focus:outline-none"
            :disabled="loading"
          />
          <button 
            type="submit" 
            :disabled="loading || !inputCommand.trim()"
            class="px-3.5 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-30 text-white font-medium flex items-center gap-1 transition-colors cursor-pointer"
          >
            <Send class="w-3.5 h-3.5" />
            <span>Run</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
