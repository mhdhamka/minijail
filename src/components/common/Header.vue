<script setup lang="ts">
import { 
  Search, 
  Plus, 
  ShieldCheck,
  Sun,
  Moon
} from 'lucide-vue-next';
import { useTheme } from '../../composables/useTheme';

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
    <!-- Left: Sleek Search Bar -->
    <div class="flex items-center gap-3 flex-1 max-w-md">
      <div class="relative w-full">
        <Search class="w-3.5 h-3.5 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
        <input 
          :value="searchQuery"
          @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          type="text" 
          placeholder="Search containers, images, ports..."
          class="w-full pl-8 pr-8 py-1 rounded-md bg-[#0D1117] border border-[#21262D] text-xs font-mono text-slate-200 placeholder:text-slate-600 focus:outline-none focus:border-blue-500 transition-all shadow-inner"
        />
        <div class="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] text-slate-600 font-mono border border-slate-800 bg-[#161B22] rounded px-1">
          /
        </div>
      </div>
    </div>

    <!-- Right: Minimalist Controls & Actions -->
    <div class="flex items-center gap-1.5 sm:gap-2">
      <!-- Theme Toggle -->
      <button 
        id="btn-theme-toggle"
        @click="toggleTheme"
        class="p-2 rounded-md text-slate-400 hover:text-slate-200 hover:bg-[#21262D] transition-colors cursor-pointer"
        :title="theme === 'dark' ? 'Switch to Light Theme' : 'Switch to Dark Theme'"
      >
        <Sun v-if="theme === 'dark'" class="w-4 h-4 text-amber-400" />
        <Moon v-else class="w-4 h-4 text-slate-400" />
      </button>

      <!-- Kernel Internals -->
      <button
        id="btn-open-primitives"
        @click="emit('open-primitives')"
        class="hidden sm:flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium text-slate-300 bg-[#21262D] hover:bg-[#30363D] border border-[#30363D] transition-colors cursor-pointer"
        title="Inspect Go kernel code: namespaces, cgroups, pivot_root"
      >
        <ShieldCheck class="w-3.5 h-3.5 text-cyan-400" />
        <span>Kernel</span>
      </button>

      <!-- Add Container -->
      <button
        id="btn-deploy-container"
        @click="emit('open-spawner')"
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium text-white bg-blue-600 hover:bg-blue-500 shadow-sm transition-all cursor-pointer"
      >
        <Plus class="w-4 h-4" />
        <span>Container</span>
      </button>
    </div>
  </header>
</template>