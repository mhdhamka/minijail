<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import type { Container } from '../types/container';
import { execCommand } from '../api';
import { X, Terminal as TerminalIcon, Send, Sparkles, RefreshCw } from 'lucide-vue-next';

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
});

function scrollToBottom() {
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight;
    }
  });
}

async function runCommand(cmd?: string) {
  const target = (cmd || inputCommand.value).trim();
  if (!target || loading.value) return;

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
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-sm">
    <div class="bg-slate-950 border border-slate-700/80 rounded-2xl w-full max-w-3xl overflow-hidden shadow-2xl flex flex-col h-[80vh]">
      <!-- Header -->
      <div class="px-5 py-3.5 border-b border-slate-800 flex items-center justify-between bg-slate-900/80">
        <div class="flex items-center gap-2.5">
          <div class="p-1.5 rounded-lg bg-cyan-950 border border-cyan-800 text-cyan-400">
            <TerminalIcon class="w-4 h-4" />
          </div>
          <div>
            <div class="text-xs font-bold text-white font-mono flex items-center gap-2">
              <span>exec -it {{ container.name }} sh</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-300">
                PID 1 in {{ container.namespaces.pidInode }}
              </span>
            </div>
            <div class="text-[10px] text-slate-400 font-mono">
              UTS: {{ container.namespaces.hostname }} • IP: {{ container.namespaces.virtualIp }}
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button 
            @click="emit('refresh')"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
            title="Refresh stream"
          >
            <RefreshCw class="w-4 h-4" />
          </button>
          <button 
            @click="emit('close')"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Quick Command Chips -->
      <div class="px-4 py-2 bg-slate-900/40 border-b border-slate-800/80 flex items-center gap-2 overflow-x-auto text-[11px] font-mono">
        <span class="text-slate-500 shrink-0 flex items-center gap-1 text-[10px]">
          <Sparkles class="w-3 h-3 text-cyan-400" /> Quick Exec:
        </span>
        <button 
          @click="runCommand('uname -a')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-cyan-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          uname -a
        </button>
        <button 
          @click="runCommand('hostname')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-cyan-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          hostname
        </button>
        <button 
          @click="runCommand('ps aux')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-emerald-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          ps aux
        </button>
        <button 
          @click="runCommand('cat /etc/os-release')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          cat /etc/os-release
        </button>
        <button 
          @click="runCommand('ip a')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-blue-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          ip a
        </button>
        <button 
          @click="runCommand('cat /sys/fs/cgroup/memory.max')"
          class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-amber-300 border border-slate-700 transition-colors shrink-0 cursor-pointer"
        >
          cat memory.max
        </button>
      </div>

      <!-- Terminal Output Window -->
      <div 
        ref="terminalBody"
        class="flex-1 p-4 overflow-y-auto bg-slate-950 font-mono text-xs space-y-1.5 selection:bg-cyan-500/30"
      >
        <div 
          v-for="(line, idx) in terminalLines" 
          :key="idx"
          class="flex items-start gap-2 leading-relaxed"
        >
          <span class="text-slate-600 text-[10px] select-none shrink-0 w-14">
            {{ line.time }}
          </span>

          <!-- Stdin command -->
          <template v-if="line.stream === 'stdin'">
            <span class="text-cyan-400 font-semibold select-none">root@{{ container.namespaces.hostname }}:/#</span>
            <span class="text-white font-medium whitespace-pre-wrap break-all">{{ line.text.replace('$ ', '') }}</span>
          </template>

          <!-- System runtime event -->
          <template v-else-if="line.stream === 'system'">
            <span class="text-indigo-400 text-[11px]">[minijail-kernel]</span>
            <span class="text-indigo-200/90 whitespace-pre-wrap break-all">{{ line.text }}</span>
          </template>

          <!-- Stderr / Error -->
          <template v-else-if="line.stream === 'stderr'">
            <span class="text-rose-400 select-none">[stderr]</span>
            <span class="text-rose-300 whitespace-pre-wrap break-all">{{ line.text }}</span>
          </template>

          <!-- Stdout -->
          <template v-else>
            <span class="text-slate-300 whitespace-pre-wrap break-all">{{ line.text }}</span>
          </template>
        </div>

        <div v-if="loading" class="flex items-center gap-2 text-cyan-400 pt-2 text-[11px]">
          <span class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-ping"></span>
          <span>Executing in Go container namespace...</span>
        </div>
      </div>

      <!-- Interactive Input Prompt -->
      <div class="p-3 border-t border-slate-800 bg-slate-900/90">
        <form @submit.prevent="runCommand()" class="flex items-center gap-2 font-mono text-xs">
          <span class="text-cyan-400 font-semibold shrink-0 select-none">
            root@{{ container.namespaces.hostname }}:/#
          </span>
          <input 
            v-model="inputCommand"
            type="text" 
            placeholder="Type command (e.g. uname -a, ps aux, touch /tmp/test, cat /etc/hosts)..."
            class="flex-1 bg-transparent text-slate-100 placeholder:text-slate-600 focus:outline-none"
            :disabled="loading || container.status !== 'running'"
          />
          <button 
            type="submit" 
            :disabled="loading || !inputCommand.trim() || container.status !== 'running'"
            class="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 disabled:opacity-30 text-white font-medium flex items-center gap-1 transition-colors cursor-pointer"
          >
            <Send class="w-3.5 h-3.5" />
            <span>Send</span>
          </button>
        </form>
        <div v-if="container.status !== 'running'" class="text-[10px] text-amber-400 mt-1 font-mono">
          Container is currently {{ container.status }}. Start the container to execute live commands.
        </div>
      </div>
    </div>
  </div>
</template>
