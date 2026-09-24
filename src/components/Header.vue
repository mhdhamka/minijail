<script setup lang="ts">
import { ref } from 'vue';
import { 
  Search, 
  Terminal, 
  Plus, 
  HelpCircle, 
  Settings, 
  ExternalLink,
  ShieldCheck,
  RefreshCw,
  Bell,
  Sun,
  Moon
} from 'lucide-vue-next';
import DockerLogo from './DockerLogo.vue';
import { useTheme } from '../composables/useTheme';

interface Props {
  connected: boolean;
  activeCount: number;
  totalCount: number;
  searchQuery: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'update:searchQuery', val: string): void;
  (e: 'open-spawner'): void;
  (e: 'open-cli'): void;
  (e: 'open-primitives'): void;
  (e: 'refresh'): void;
}>();

const { theme, toggleTheme } = useTheme();
</script>

<template>
  <header id="docker-header" class="h-14 bg-[#161B22] border-b border-[#232A35] flex items-center justify-between px-4 sm:px-6 sticky top-0 z-30 select-none">
    <!-- Left: Search Box (Global Docker Desktop search) -->
    <div class="flex items-center gap-3 flex-1 max-w-lg">
      <div class="relative w-full">
        <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
        <input 
          :value="searchQuery"
          @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          type="text" 
          placeholder="Search containers, images, ports (e.g. alpine, 8080, running)..."
          class="w-full pl-9 pr-8 py-1.5 rounded-lg bg-[#0E1217] border border-[#2B3545] text-xs font-mono text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-[#1D63ED] transition-colors"
        />
        <div class="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] text-slate-500 font-mono border border-slate-700 rounded px-1">
          /
        </div>
      </div>
    </div>

    <!-- Right: Quick Actions and Docker Desktop Controls -->
    <div class="flex items-center gap-2 sm:gap-3">
      <!-- Dark / Light Theme Toggle Button -->
      <button 
        id="btn-theme-toggle"
        @click="toggleTheme"
        class="p-2 rounded-lg text-slate-400 hover:text-amber-300 hover:bg-[#202734] transition-colors cursor-pointer flex items-center gap-1.5"
        :title="theme === 'dark' ? 'Switch to Light Theme' : 'Switch to Dark Theme'"
      >
        <Sun v-if="theme === 'dark'" class="w-4 h-4 text-amber-400" />
        <Moon v-else class="w-4 h-4 text-indigo-500" />
        <span class="text-xs font-mono hidden md:inline">{{ theme === 'dark' ? 'Light' : 'Dark' }}</span>
      </button>

      <!-- Refresh -->
      <button 
        @click="emit('refresh')"
        class="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-[#202734] transition-colors cursor-pointer"
        title="Refresh Engine State"
      >
        <RefreshCw class="w-4 h-4" />
      </button>

      <!-- Kernel Primitives Shortcut -->
      <button
        id="btn-open-primitives"
        @click="emit('open-primitives')"
        class="hidden sm:flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-slate-300 bg-[#1E2633] hover:bg-[#273244] border border-[#2E3B4E] transition-colors cursor-pointer"
        title="Inspect Go kernel code: namespaces, cgroups, pivot_root"
      >
        <ShieldCheck class="w-3.5 h-3.5 text-cyan-400" />
        <span>Kernel Internals</span>
      </button>

      <!-- Docker CLI quick launch -->
      <button
        id="btn-open-cli"
        @click="emit('open-cli')"
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-mono font-medium text-emerald-300 bg-emerald-950/40 hover:bg-emerald-900/50 border border-emerald-800/60 transition-colors cursor-pointer"
        title="Open Docker CLI terminal"
      >
        <Terminal class="w-3.5 h-3.5 text-emerald-400" />
        <span>docker cli</span>
      </button>

      <!-- + Run Container (Docker primary action button) -->
      <button
        id="btn-deploy-container"
        @click="emit('open-spawner')"
        class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold text-white bg-[#1D63ED] hover:bg-[#1A57D0] shadow-sm transition-all cursor-pointer"
      >
        <Plus class="w-4 h-4" />
        <span>+ Add Container</span>
      </button>

      <!-- User avatar / Docker ID -->
      <div class="h-6 w-px bg-[#26303F] mx-1"></div>
      <div class="flex items-center gap-2 pl-1">
        <div class="w-7 h-7 rounded-full bg-[#1D63ED]/20 border border-[#1D63ED]/40 flex items-center justify-center text-[#0db7ed] font-bold text-xs font-mono cursor-pointer" title="Docker ID: developer">
          D
        </div>
      </div>
    </div>
  </header>
</template>
