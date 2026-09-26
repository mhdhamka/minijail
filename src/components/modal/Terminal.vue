<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import type { Container } from '../../types/container';
import { execCommand } from '../../api';
import { X, Terminal as TerminalIcon, Send, Sparkles, RefreshCw, Trash2, ArrowUp, ArrowDown } from 'lucide-vue-next';

interface Props {
  container: Container;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
}>();

const inputCommand = ref('');
const loading = ref(false);
const terminalLines = ref<{ stream: 'stdout' | 'stderr' | 'stdin' | 'system'; text: string; time: string }[]>([]);
const terminalBody = ref<HTMLElement | null>(null);
const commandInputRef = ref<HTMLInputElement | null>(null);

// Command history state
const commandHistory = ref<string[]>([]);
const historyIndex = ref(-1);

onMounted(() => {
  // Load initial logs
  props.container.logs.forEach(log => {
    terminalLines.value.push({
      stream: log.stream,
      text: log.message,
      time: log.timestamp,
    });
  });
  scrollToBottom();
  
  // Auto-focus input on mount
  nextTick(() => {
    commandInputRef.value?.focus();
  });
});

function scrollToBottom() {
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight;
    }
  });
}

function clearTerminal() {
  terminalLines.value = [];
}

function handleKeydown(e: KeyboardEvent) {
  if (commandHistory.value.length === 0) return;

  if (e.key === 'ArrowUp') {
    e.preventDefault();
    if (historyIndex.value < commandHistory.value.length - 1) {
      historyIndex.value++;
      inputCommand.value = commandHistory.value[commandHistory.value.length - 1 - historyIndex.value];
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    if (historyIndex.value > 0) {
      historyIndex.value--;
      inputCommand.value = commandHistory.value[commandHistory.value.length - 1 - historyIndex.value];
    } else if (historyIndex.value === 0) {
      historyIndex.value = -1;
      inputCommand.value = '';
    }
  }
}

async function runCommand(cmd?: string) {
  const target = (cmd || inputCommand.value).trim();
  if (!target || loading.value) return;

  // Push to history
  commandHistory.value.push(target);
  historyIndex.value = -1;

  const now = new Date().toTimeString().slice(0, 8);
  terminalLines.value.push({
    stream: 'stdin',
    text: `$ ${target}`,
    time: now,
  });

  inputCommand.value = '';
  loading.value = true;
  scrollToBottom();

  try {
    const res = await execCommand(props.container.id, target);
    if (res.stdout) {
      terminalLines.value.push({
        stream: 'stdout',
        text: res.stdout.trimEnd(),
        time: now,
      });
    }
    if (res.stderr) {
      terminalLines.value.push({
        stream: 'stderr',
        text: res.stderr.trimEnd(),
        time: now,
      });
    }
    emit('refresh');
  } catch (err: any) {
    terminalLines.value.push({
      stream: 'stderr',
      text: `exec error: ${err.message}`,
      time: now,
    });
  } finally {
    loading.value = false;
    scrollToBottom();
    commandInputRef.value?.focus();
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fadeIn">
    <div class="bg-slate-950/95 border border-slate-700/60 rounded-2xl w-full max-w-4xl overflow-hidden shadow-2xl shadow-cyan-950/20 flex flex-col h-[82vh]">
      
      <!-- Modern Window Header -->
      <div class="px-5 py-3.5 border-b border-slate-800/80 flex items-center justify-between bg-gradient-to-r from-slate-900 via-slate-900/90 to-slate-950">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 shadow-inner">
            <TerminalIcon class="w-4 h-4" />
          </div>
          <div>
            <div class="text-xs font-bold text-slate-100 font-mono flex items-center gap-2">
              <span class="text-cyan-400">exec -it</span> 
              <span class="text-white">{{ container.name }}</span> 
              <span class="text-slate-400">sh</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-cyan-950/80 border border-cyan-800/50 text-cyan-300 font-normal">
                PID 1 : {{ container.namespaces.pidInode }}
              </span>
            </div>
            <div class="text-[10px] text-slate-400 font-mono mt-0.5 flex items-center gap-2">
              <span class="text-slate-400">UTS: <strong class="text-slate-300">{{ container.namespaces.hostname }}</strong></span>
              <span>•</span>
              <span class="text-slate-400">IP: <strong class="text-slate-300">{{ container.namespaces.virtualIp }}</strong></span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-1.5">
          <button 
            @click="clearTerminal"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
            title="Clear terminal buffer"
          >
            <Trash2 class="w-4 h-4" />
          </button>
          <button 
            @click="emit('refresh')"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors cursor-pointer"
            title="Refresh stream"
          >
            <RefreshCw class="w-4 h-4" />
          </button>
          <div class="h-4 w-[1px] bg-slate-800 mx-1"></div>
          <button 
            @click="emit('close')"
            class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
            title="Close terminal"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Quick Command Chips -->
      <div class="px-4 py-2.5 bg-slate-900/40 border-b border-slate-800/60 flex items-center gap-2 overflow-x-auto text-[11px] font-mono scrollbar-none">
        <span class="text-slate-500 shrink-0 flex items-center gap-1 text-[10px] font-sans font-medium uppercase tracking-wider">
          Quick Actions:
        </span>
        <button 
          @click="runCommand('uname -a')"
          class="px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 text-cyan-300 border border-slate-800 hover:border-cyan-500/40 transition-all shrink-0 cursor-pointer shadow-sm"
        >
          uname -a
        </button>
        <button 
          @click="runCommand('ps aux')"
          class="px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 text-emerald-300 border border-slate-800 hover:border-emerald-500/40 transition-all shrink-0 cursor-pointer shadow-sm"
        >
          ps aux
        </button>
        <button 
          @click="runCommand('ip a')"
          class="px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 text-blue-300 border border-slate-800 hover:border-blue-500/40 transition-all shrink-0 cursor-pointer shadow-sm"
        >
          ip a
        </button>
        <button 
          @click="runCommand('cat /sys/fs/cgroup/memory.max')"
          class="px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 text-amber-300 border border-slate-800 hover:border-amber-500/40 transition-all shrink-0 cursor-pointer shadow-sm"
        >
          cat memory.max
        </button>
        <button 
          @click="runCommand('df -h')"
          class="px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 text-purple-300 border border-slate-800 hover:border-purple-500/40 transition-all shrink-0 cursor-pointer shadow-sm"
        >
          df -h (OverlayFS)
        </button>
      </div>

      <!-- Terminal Output Screen -->
      <div 
        ref="terminalBody"
        class="flex-1 p-5 overflow-y-auto bg-slate-950 font-mono text-xs space-y-2 selection:bg-cyan-500/30"
      >
        <div 
          v-for="(line, idx) in terminalLines" 
          :key="idx"
          class="flex items-start gap-3 leading-relaxed group"
        >
          <span class="text-slate-600 text-[10px] select-none shrink-0 w-14 pt-0.5 opacity-60 group-hover:opacity-100 transition-opacity">
            {{ line.time }}
          </span>

          <!-- Stdin command -->
          <template v-if="line.stream === 'stdin'">
            <div class="flex-1 flex items-start gap-2 break-all">
              <span class="text-cyan-400 font-semibold select-none">root@{{ container.namespaces.hostname }}:/#</span>
              <span class="text-slate-100 font-medium whitespace-pre-wrap">{{ line.text.replace('$ ', '') }}</span>
            </div>
          </template>

          <!-- System event -->
          <template v-else-if="line.stream === 'system'">
            <div class="flex-1 text-indigo-300/90 bg-indigo-950/30 border border-indigo-900/40 px-3 py-1.5 rounded-lg">
              <span class="text-indigo-400 font-semibold mr-2">[sandbox-kernel]</span>
              <span class="whitespace-pre-wrap break-all">{{ line.text }}</span>
            </div>
          </template>

          <!-- Stderr -->
          <template v-else-if="line.stream === 'stderr'">
            <div class="flex-1 text-rose-300 bg-rose-950/20 border border-rose-900/30 px-3 py-1.5 rounded-lg break-all">
              <span class="text-rose-400 font-semibold mr-2">[stderr]</span>
              <span class="whitespace-pre-wrap">{{ line.text }}</span>
            </div>
          </template>

          <!-- Stdout -->
          <template v-else>
            <div class="flex-1 text-slate-300 whitespace-pre-wrap break-all">
              {{ line.text }}
            </div>
          </template>
        </div>

        <div v-if="loading" class="flex items-center gap-2.5 text-cyan-400 pt-2 text-[11px] bg-cyan-950/20 px-3 py-2 rounded-lg border border-cyan-900/30 w-fit">
          <span class="w-2 h-2 rounded-full bg-cyan-400 animate-ping"></span>
          <span>Executing command inside isolated container namespace...</span>
        </div>
      </div>

      <!-- Interactive Input Bar -->
      <div class="p-3.5 border-t border-slate-800/80 bg-slate-900/90">
        <form @submit.prevent="runCommand()" class="flex items-center gap-3 font-mono text-xs">
          <span class="text-cyan-400 font-semibold shrink-0 select-none flex items-center gap-1">
            root@{{ container.namespaces.hostname }}:/#
          </span>
          <div class="relative flex-1 flex items-center">
            <input 
              ref="commandInputRef"
              v-model="inputCommand"
              @keydown="handleKeydown"
              type="text" 
              placeholder="Type command (e.g. uname -a, ps aux, touch /tmp/test)..."
              class="w-full bg-slate-950/80 border border-slate-800 focus:border-cyan-500/50 rounded-xl px-3.5 py-2 text-slate-100 placeholder:text-slate-600 focus:outline-none transition-all shadow-inner"
              :disabled="loading || container.status !== 'running'"
            />
            <div class="absolute right-3 flex items-center gap-1 text-[10px] text-slate-500 pointer-events-none hidden sm:flex">
              <span class="px-1 py-0.5 rounded bg-slate-800/80 border border-slate-700/50">↑↓ history</span>
            </div>
          </div>
          <button 
            type="submit" 
            :disabled="loading || !inputCommand.trim() || container.status !== 'running'"
            class="px-4 py-2 rounded-xl bg-gradient-to-r from-cyan-600 to-cyan-500 hover:from-cyan-500 hover:to-cyan-400 disabled:opacity-30 disabled:hover:from-cyan-600 text-white font-medium flex items-center gap-1.5 transition-all shadow-lg shadow-cyan-950 cursor-pointer"
          >
            <Send class="w-3.5 h-3.5" />
            <span>Run</span>
          </button>
        </form>
        <div v-if="container.status !== 'running'" class="text-[11px] text-amber-400/90 mt-2 font-mono flex items-center gap-1.5 bg-amber-950/30 border border-amber-900/40 px-3 py-1.5 rounded-lg">
          <span>⚠️ Container is currently <strong class="capitalize">{{ container.status }}</strong>. Start the container to execute live commands.</span>
        </div>
      </div>
    </div>
  </div>
</template>