<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { 
  Search, 
  Plus, 
  ShieldCheck,
  Sun,
  Moon,
  Command,
  X
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
const searchInputRef = ref<HTMLInputElement | null>(null);

// Global keyboard shortcut to focus search bar (Press '/' or 'Cmd/Ctrl + K')
function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.key === '/' && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') ||
      ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k')) {
    e.preventDefault();
    searchInputRef.value?.focus();
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown);
});

function clearSearch() {
  emit('update:searchQuery', '');
  searchInputRef.value?.focus();
}
</script>

<template>
  <header id="docker-header" class="h-16 bg-[#12161D]/90 backdrop-blur-md border-b border-[#232A35] flex items-center justify-between px-5 sm:px-6 sticky top-0 z-30 select-none transition-all shadow-sm">
    
    <!-- Left: Modernized Interactive Search Bar -->
    <div class="flex items-center gap-3 flex-1 max-w-lg">
      <div class="relative w-full group">
        <Search class="w-4 h-4 text-slate-400 group-focus-within:text-[#0db7ed] absolute left-3.5 top-1/2 -translate-y-1/2 transition-colors" />
        <input 
          ref="searchInputRef"
          :value="searchQuery"
          @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          type="text" 
          placeholder="Search containers, images, ports, or IDs..."
          class="w-full pl-10 pr-16 py-2 rounded-xl bg-[#080B10] border border-[#232A35] text-xs font-mono text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-[#1D63ED] focus:ring-2 focus:ring-[#1D63ED]/20 transition-all shadow-inner"
        />
        
        <!-- Clear search button (appears when typing) -->
        <button 
          v-if="searchQuery"
          @click="clearSearch"
          class="absolute right-7 top-1/2 -translate-y-1/2 p-0.5 rounded-md text-slate-400 hover:text-white hover:bg-[#1E2633] transition-colors cursor-pointer"
          title="Clear search"
        >
          <X class="w-3.5 h-3.5" />
        </button>

        <!-- Keyboard Shortcut Badge -->
        <div class="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center gap-0.5 text-[10px] text-slate-500 font-mono pointer-events-none bg-[#161B22] border border-[#2B3545] rounded px-1.5 py-0.5 shadow-sm">
          <Command class="w-2.5 h-2.5" />
          <span>K</span>
        </div>
      </div>
    </div>

    <!-- Right: Minimalist Controls & Actions -->
    <div class="flex items-center gap-2 sm:gap-2.5">
      
      <!-- Theme Toggle -->
      <button 
        id="btn-theme-toggle"
        @click="toggleTheme"
        class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-[#1C2430] border border-transparent hover:border-[#2B3545] transition-all cursor-pointer"
        :title="theme === 'dark' ? 'Switch to Light Theme' : 'Switch to Dark Theme'"
      >
        <Sun v-if="theme === 'dark'" class="w-4 h-4 text-amber-400" />
        <Moon v-else class="w-4 h-4 text-slate-300" />
      </button>

      <!-- Kernel Internals -->
      <button
        id="btn-open-primitives"
        @click="emit('open-primitives')"
        class="hidden sm:flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-medium text-slate-200 bg-[#18202A] hover:bg-[#222C3A] border border-[#2B3545] hover:border-cyan-500/40 transition-all cursor-pointer shadow-sm active:scale-95"
        title="Inspect Go kernel code: namespaces, cgroups, pivot_root"
      >
        <ShieldCheck class="w-4 h-4 text-cyan-400" />
        <span>Kernel Suite</span>
      </button>

      <!-- Add Container -->
      <button
        id="btn-deploy-container"
        @click="emit('open-spawner')"
        class="flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-semibold text-white bg-[#1D63ED] hover:bg-[#1A57D0] shadow-md shadow-blue-600/20 border border-blue-500/30 transition-all cursor-pointer active:scale-95"
      >
        <Plus class="w-4 h-4" />
        <span>Deploy Container</span>
      </button>
    </div>
  </header>
</template>