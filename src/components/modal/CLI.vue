<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { executeCli } from '../../api.ts';
import { X, Send, Trash2, Copy, Check, Terminal, ShieldAlert, HelpCircle } from 'lucide-vue-next';

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
}>();

const inputCommand = ref('docker ps');
const loading = ref(false);
const history = ref<string[]>([]);
const historyIndex = ref(-1);
const copiedIndex = ref<number | null>(null);

interface TerminalLine {
  command: string;
  output: string;
  error?: string;
  time: string;
}

const terminalLines = ref<TerminalLine[]>([]);
const terminalBody = ref<HTMLElement | null>(null);

onMounted(async () => {
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
  if (!cmdToRun) inputCommand.value = '';
  loading.value = true;
  scrollToBottom();

  const now = new Date().toTimeString().slice(0, 8);

  // Fallback handler if user types --help directly without 'docker' prefix
  let actualTarget = target;
  if (target === '--help' || target === '-h') {
    actualTarget = 'docker --help';
  }

  try {
    const res = await executeCli(actualTarget);
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
    e.preventDefault();
    if (historyIndex.value > 0) {
      historyIndex.value--;
      inputCommand.value = history.value[historyIndex.value];
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    if (historyIndex.value < history.value.length - 1) {
      historyIndex.value++;
      inputCommand.value = history.value[historyIndex.value];
    } else {
      historyIndex.value = history.value.length;
      inputCommand.value = '';
    }
  }
}

function clearConsole() {
  terminalLines.value = [];
}

function copyOutput(text: string, index: number) {
  navigator.clipboard.writeText(text);
  copiedIndex.value = index;
  setTimeout(() => {
    if (copiedIndex.value === index) copiedIndex.value = null;
  }, 2000);
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md select-none animate-in fade-in duration-200">
    <div class="bg-[#12161D] border border-[#2B3545] rounded-2xl w-full max-w-4xl overflow-hidden shadow-2xl flex flex-col h-[88vh] text-slate-200">
      
      <!-- Top Window Header -->
      <div class="px-5 py-4 border-b border-[#232A35] flex items-center justify-between bg-[#0E1217]">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-[#1D63ED]/10 border border-[#1D63ED]/30 text-[#0db7ed]">
            <Terminal class="w-5 h-5" />
          </div>
          <div>
            <div class="text-xs font-bold text-white font-mono flex items-center gap-2">
              <span>Docker CLI Engine</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-[#1D63ED]/20 text-[#0db7ed] border border-[#1D63ED]/30">
                Minijail-Go Engine
              </span>
            </div>
            <div class="text-[11px] text-slate-400 font-sans mt-0.5">
              Execute live container operations, namespaces setups, and inspect sandbox metrics.
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button 
            @click="clearConsole"
            class="px-2.5 py-1.5 rounded-lg bg-[#18202A] hover:bg-[#222C3A] border border-[#2B3545] text-xs font-mono text-slate-400 hover:text-white transition-colors flex items-center gap-1.5 cursor-pointer"
            title="Clear terminal history"
          >
            <Trash2 class="w-3.5 h-3.5" />
            <span>Clear</span>
          </button>
          <button 
            @click="emit('close')"
            class="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-[#202836] transition-colors cursor-pointer"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Command Presets Toolbar -->
      <div class="px-4 py-2.5 bg-[#090D12] border-b border-[#232A35] flex items-center gap-2 overflow-x-auto text-xs font-mono">
        <span class="text-slate-500 shrink-0 flex items-center gap-1 text-[11px] font-semibold">
          Presets:
        </span>
        <button 
          @click="runCli('docker --help')"
          class="px-2.5 py-1 rounded-lg bg-[#1D63ED]/20 hover:bg-[#1D63ED]/30 text-sky-300 border border-[#1D63ED]/40 transition-all shrink-0 cursor-pointer active:scale-95 flex items-center gap-1"
        >
          <HelpCircle class="w-3.5 h-3.5" /> docker --help
        </button>
        <button 
          @click="runCli('docker ps -a')"
          class="px-2.5 py-1 rounded-lg bg-[#161B22] hover:bg-[#1F2937] text-slate-300 border border-[#2B3545] transition-all shrink-0 cursor-pointer active:scale-95"
        >
          docker ps -a
        </button>
        <button 
          @click="runCli('docker stats')"
          class="px-2.5 py-1 rounded-lg bg-[#161B22] hover:bg-[#1F2937] text-cyan-300 border border-[#2B3545] transition-all shrink-0 cursor-pointer active:scale-95"
        >
          docker stats
        </button>
        <button 
          @click="runCli('docker run -d --name redis-cache -m 64m alpine:3.19 sh')"
          class="px-2.5 py-1 rounded-lg bg-[#161B22] hover:bg-[#1F2937] text-emerald-300 border border-[#2B3545] transition-all shrink-0 cursor-pointer active:scale-95"
        >
          docker run (Redis)
        </button>
        <button 
          @click="runCli('docker images')"
          class="px-2.5 py-1 rounded-lg bg-[#161B22] hover:bg-[#1F2937] text-purple-300 border border-[#2B3545] transition-all shrink-0 cursor-pointer active:scale-95"
        >
          docker images
        </button>
        <button 
          @click="runCli('docker version')"
          class="px-2.5 py-1 rounded-lg bg-[#161B22] hover:bg-[#1F2937] text-indigo-300 border border-[#2B3545] transition-all shrink-0 cursor-pointer active:scale-95"
        >
          docker version
        </button>
      </div>

      <!-- Terminal Screen Body -->
      <div 
        ref="terminalBody"
        class="flex-1 p-5 overflow-y-auto bg-[#070A0F] font-mono text-xs space-y-5 selection:bg-[#1D63ED]/40 scroll-smooth"
      >
        <div v-for="(entry, idx) in terminalLines" :key="idx" class="space-y-2 group">
          <!-- Command Prompt Line -->
          <div class="flex items-center justify-between text-slate-400 bg-[#10151D] px-3 py-1.5 rounded-lg border border-[#1E2633]">
            <div class="flex items-center gap-2">
              <span class="text-slate-500 text-[10px]">{{ entry.time }}</span>
              <span class="text-[#0db7ed] font-bold">root@minijail:~#</span>
              <span class="text-white font-semibold">{{ entry.command }}</span>
            </div>
            
            <button 
              @click="copyOutput(entry.output || entry.error || '', idx)"
              class="opacity-0 group-hover:opacity-100 transition-opacity px-2 py-0.5 rounded bg-[#1A222D] text-slate-300 hover:text-white border border-[#2B3545] flex items-center gap-1 text-[10px] cursor-pointer"
            >
              <Check v-if="copiedIndex === idx" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3 text-slate-400" />
              <span>{{ copiedIndex === idx ? 'Copied' : 'Copy' }}</span>
            </button>
          </div>

          <!-- Command Success Output -->
          <div v-if="entry.output" class="bg-[#0D1117] p-4 rounded-xl border border-[#212B38] text-slate-200 whitespace-pre font-mono text-[11px] overflow-x-auto leading-relaxed shadow-inner">
{{ entry.output }}
          </div>

          <!-- Command Error Output -->
          <div v-if="entry.error" class="bg-rose-950/30 p-3.5 rounded-xl border border-rose-900/50 text-rose-300 whitespace-pre-wrap font-mono text-[11px] flex items-start gap-2">
            <ShieldAlert class="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
            <div>
              <span class="font-bold">Execution Error:</span>
              <div class="mt-0.5">{{ entry.error }}</div>
            </div>
          </div>
        </div>

        <!-- Real-time Loading Indicator -->
        <div v-if="loading" class="flex items-center gap-3 text-[#0db7ed] py-2 text-[11px] font-mono animate-pulse">
          <span class="w-2 h-2 rounded-full bg-[#0db7ed] shadow-[0_0_8px_#0db7ed]"></span>
          <span>Interpreting Go system call & compiling namespaces...</span>
        </div>
      </div>

      <!-- Interactive Input Footer -->
      <div class="p-3.5 border-t border-[#232A35] bg-[#0E1217]">
        <form @submit.prevent="runCli()" class="flex items-center gap-3 font-mono text-xs bg-[#070A0F] px-4 py-2.5 rounded-xl border border-[#2B3545] focus-within:border-[#1D63ED] transition-colors">
          <span class="text-[#0db7ed] font-bold shrink-0 select-none">root@minijail:~#</span>
          <input 
            v-model="inputCommand"
            @keydown="handleKeydown"
            type="text" 
            placeholder="Type command (e.g. docker --help, docker ps, docker stats)..."
            class="flex-1 bg-transparent text-slate-100 placeholder:text-slate-600 focus:outline-none"
            :disabled="loading"
          />
          <button 
            type="submit" 
            :disabled="loading || !inputCommand.trim()"
            class="px-4 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] disabled:opacity-30 text-white font-medium flex items-center gap-1.5 transition-all cursor-pointer shadow-md active:scale-95"
          >
            <Send class="w-3.5 h-3.5" />
            <span>Execute</span>
          </button>
        </form>
        <div class="flex items-center justify-between text-[10px] text-slate-500 font-mono px-1 mt-2">
          <span>Tip: Use <kbd class="px-1 py-0.5 rounded bg-[#18202A] text-slate-300 border border-[#2B3545]">↑</kbd> <kbd class="px-1 py-0.5 rounded bg-[#18202A] text-slate-300 border border-[#2B3545]">↓</kbd> to cycle command history. Type <code class="text-sky-400">docker --help</code> for guidance.</span>
          <span class="text-emerald-400 font-semibold">Engine Connected</span>
        </div>
      </div>

    </div>
  </div>
</template>