<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Package, 
  Play, 
  Search, 
} from 'lucide-vue-next';

interface ImageItem {
  id: string;
  repository: string;
  tag: string;
  size: string;
  created: string;
  desc: string;
  inUseCount: number;
}

const images = ref<ImageItem[]>([
  {
    id: 'sha256:05455a08881e',
    repository: 'alpine',
    tag: '3.19',
    size: '7.38 MB',
    created: '2 weeks ago',
    desc: 'Ultra-lightweight Linux distribution based on musl libc and busybox',
    inUseCount: 1,
  },
  {
    id: 'sha256:a416a98b71e2',
    repository: 'busybox',
    tag: '1.36',
    size: '4.26 MB',
    created: '1 month ago',
    desc: 'The Swiss Army Knife of Embedded Linux — tiny static binaries',
    inUseCount: 1,
  },
  {
    id: 'sha256:35a88802559d',
    repository: 'ubuntu',
    tag: '22.04',
    size: '77.8 MB',
    created: '3 weeks ago',
    desc: 'Ubuntu LTS root filesystem with apt package manager and GNU utilities',
    inUseCount: 0,
  },
  {
    id: 'sha256:8b45cd2a5f10',
    repository: 'nginx',
    tag: 'alpine',
    size: '23.4 MB',
    created: '5 days ago',
    desc: 'High-performance HTTP server and reverse proxy on Alpine Linux',
    inUseCount: 1,
  },
]);

const searchQuery = ref('');

const filteredImages = computed(() => {
  if (!searchQuery.value) return images.value;
  const q = searchQuery.value.toLowerCase();
  return images.value.filter(
    img => img.repository.toLowerCase().includes(q) || img.tag.toLowerCase().includes(q)
  );
});

const emit = defineEmits<{
  (e: 'run-image', image: string): void;
}>();
</script>

<template>
  <div class="space-y-4 select-none">
    <!-- Top Bar: Title & Search -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h2 class="text-xl font-bold text-white flex items-center gap-2 font-sans">
          <span>Images</span>
          <span class="text-xs px-2 py-0.5 rounded-full bg-[#1E2633] text-slate-400 border border-[#2B3545] font-mono">
            {{ filteredImages.length }} local
          </span>
        </h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Local Docker rootfs images and cached base layers for container instantiation.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <div class="relative w-64">
          <Search class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Search images..."
            class="w-full pl-8 pr-3 py-1.5 rounded-lg bg-[#161B22] border border-[#2B3545] text-xs font-mono text-slate-100 placeholder:text-slate-500 focus:outline-none focus:border-[#1D63ED]"
          />
        </div>
      </div>
    </div>

    <!-- Images Table (Authentic Docker Desktop Images View) -->
    <div class="rounded-xl border border-[#232A35] bg-[#161B22] overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs font-mono">
          <thead class="bg-[#11161D] border-b border-[#232A35] text-slate-400">
            <tr>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Image</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Tag</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Image ID</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Created</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Size</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">In Use</th>
              <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px] text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#202733]">
            <tr 
              v-for="img in filteredImages" 
              :key="img.id"
              class="hover:bg-[#1A222D] transition-colors group"
            >
              <!-- Image Repository & Description -->
              <td class="px-4 py-3">
                <div class="flex items-center gap-2.5">
                  <div class="w-7 h-7 rounded-lg bg-[#1D63ED]/10 border border-[#1D63ED]/30 flex items-center justify-center text-[#0db7ed]">
                    <Package class="w-4 h-4" />
                  </div>
                  <div>
                    <div class="font-bold text-white text-sm group-hover:text-[#0db7ed] transition-colors">
                      {{ img.repository }}
                    </div>
                    <div class="text-[10px] text-slate-400 font-sans max-w-sm truncate">
                      {{ img.desc }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- Tag -->
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded bg-[#1A222D] border border-[#2C3748] text-slate-200">
                  {{ img.tag }}
                </span>
              </td>

              <!-- ID -->
              <td class="px-4 py-3 text-slate-400">
                {{ img.id.slice(7, 19) }}
              </td>

              <!-- Created -->
              <td class="px-4 py-3 text-slate-400">
                {{ img.created }}
              </td>

              <!-- Size -->
              <td class="px-4 py-3 text-slate-300 font-semibold">
                {{ img.size }}
              </td>

              <!-- In Use -->
              <td class="px-4 py-3">
                <span 
                  class="px-2 py-0.5 rounded-full text-[10px] font-bold"
                  :class="img.inUseCount > 0 ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800' : 'bg-slate-800 text-slate-500'"
                >
                  {{ img.inUseCount > 0 ? `${img.inUseCount} container` : 'Unused' }}
                </span>
              </td>

              <!-- Run Button -->
              <td class="px-4 py-3 text-right">
                <button
                  @click="emit('run-image', `${img.repository}:${img.tag}`)"
                  class="px-3 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] text-white font-medium text-xs flex items-center gap-1.5 ml-auto transition-colors cursor-pointer"
                >
                  <Play class="w-3.5 h-3.5" />
                  <span>Run</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
