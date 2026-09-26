<script setup lang="ts">
import { ref } from 'vue';
import { 
  FolderTree, 
  HardDrive, 
  Layers, 
  Layers3
} from 'lucide-vue-next';
import OverlayFs from '../kernel/OverlayFs.vue';
import UnionFsStack from '../kernel/UnionFsStack.vue';

interface VolumeItem {
  name: string;
  driver: string;
  scope: string;
  containerId: string;
  containerName: string;
  size: string;
  created: string;
  upperDir: string;
  lowerDir: string;
}

const activeSubTab = ref<'stack' | 'playground' | 'table'>('stack');

const volumes = ref<VolumeItem[]>([
  {
    name: 'overlay2_alpine_web_rootfs',
    driver: 'overlay2',
    scope: 'local',
    containerId: 'c1a8f9024b11',
    containerName: 'alpine-web',
    size: '1.4 MB',
    created: '2 hours ago',
    upperDir: '/var/lib/minijail/overlay/c1a8f9024b11/upper',
    lowerDir: '/var/lib/minijail/images/alpine-3.19/rootfs',
  },
  {
    name: 'overlay2_busybox_worker_rootfs',
    driver: 'overlay2',
    scope: 'local',
    containerId: 'e499d0831a23',
    containerName: 'busybox-worker',
    size: '420 KB',
    created: '4 hours ago',
    upperDir: '/var/lib/minijail/overlay/e499d0831a23/upper',
    lowerDir: '/var/lib/minijail/images/busybox-1.36/rootfs',
  },
  {
    name: 'minijail_data_storage',
    driver: 'local (bind-mount)',
    scope: 'local',
    containerId: 'c1a8f9024b11',
    containerName: 'alpine-web',
    size: '12.8 MB',
    created: '1 day ago',
    upperDir: '/mnt/data',
    lowerDir: 'N/A (Host bind)',
  },
]);
</script>

<template>
  <div class="space-y-4 select-none">
    <!-- Top Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h2 class="text-xl font-bold text-white flex items-center gap-2 font-sans">
          <span>Volumes & UnionFS Architecture</span>
          <span class="text-xs px-2 py-0.5 rounded-full bg-[#1E2633] text-amber-400 border border-[#2B3545] font-mono">
            AUFS / Overlay2 + CoW
          </span>
        </h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Linux kernel union filesystems, layered base images, copy-on-write diff layers, and persistent host mounts.
        </p>
      </div>

      <!-- View Selector -->
      <div class="flex items-center gap-1.5 p-1 bg-[#141A22] border border-[#232A35] rounded-xl self-start sm:self-auto font-mono text-xs">
        <button 
          @click="activeSubTab = 'stack'"
          class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="activeSubTab === 'stack' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Layered Visual Stack Diagram (AUFS / Overlay2)"
        >
          <Layers3 class="w-3.5 h-3.5" />
          <span>Layered Stack Diagram</span>
        </button>
        <button 
          @click="activeSubTab = 'playground'"
          class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="activeSubTab === 'playground' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Interactive CoW Filesystem Playground"
        >
          <Layers class="w-3.5 h-3.5" />
          <span>CoW Playground</span>
        </button>
        <button 
          @click="activeSubTab = 'table'"
          class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5"
          :class="activeSubTab === 'table' ? 'bg-[#1D63ED] text-white font-semibold' : 'text-slate-400 hover:text-white'"
          title="Raw Volumes Table"
        >
          <HardDrive class="w-3.5 h-3.5" />
          <span>Mounted Volumes</span>
        </button>
      </div>
    </div>

    <!-- SUBVIEW 1: LAYERED VISUAL STACK DIAGRAM (AUFS / OVERLAY2) -->
    <div v-if="activeSubTab === 'stack'">
      <UnionFsStack />
    </div>

    <!-- SUBVIEW 2: INTERACTIVE OVERLAYFS PLAYGROUND -->
    <div v-else-if="activeSubTab === 'playground'">
      <OverlayFs />
    </div>

    <!-- SUBVIEW 3: RAW VOLUMES TABLE -->
    <div v-else class="space-y-4">
      <!-- Educational Callout banner (Docker OverlayFS architecture) -->
      <div class="p-4 rounded-xl bg-[#161B22] border border-[#232A35] flex items-start gap-3 text-xs">
        <div class="w-8 h-8 rounded-lg bg-[#1D63ED]/20 text-[#0db7ed] flex items-center justify-center shrink-0 border border-[#1D63ED]/30">
          <FolderTree class="w-4 h-4" />
        </div>
        <div>
          <div class="font-bold text-slate-200">How Docker Volumes & Storage Work</div>
          <div class="text-slate-400 mt-0.5 leading-relaxed font-sans">
            Docker uses <strong class="text-slate-200">OverlayFS (overlay2)</strong> to combine immutable base image layers (<code class="text-cyan-400">lowerdir</code>) with an ephemeral read-write container layer (<code class="text-cyan-400">upperdir</code>). File modifications are tracked via Copy-on-Write (COW). The container process only sees the unified <code class="text-cyan-400">merged</code> directory through Linux's <code class="text-cyan-400">pivot_root</code> syscall.
          </div>
        </div>
      </div>

      <!-- Volumes Table -->
      <div class="rounded-xl border border-[#232A35] bg-[#161B22] overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-mono">
            <thead class="bg-[#11161D] border-b border-[#232A35] text-slate-400">
              <tr>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Volume Name</th>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Driver</th>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Container</th>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">UpperDir (RW Layer)</th>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Size</th>
                <th class="px-4 py-3 font-semibold uppercase tracking-wider text-[11px]">Created</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#202733]">
              <tr 
                v-for="vol in volumes" 
                :key="vol.name"
                class="hover:bg-[#1A222D] transition-colors"
              >
                <td class="px-4 py-3 text-[#0db7ed] font-bold">
                  {{ vol.name }}
                </td>
                <td class="px-4 py-3 text-slate-300">
                  <span class="px-2 py-0.5 rounded bg-[#1A222D] border border-[#2C3748]">
                    {{ vol.driver }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-slate-200 font-semibold">{{ vol.containerName }}</span>
                </td>
                <td class="px-4 py-3 text-slate-400 truncate max-w-xs" :title="vol.upperDir">
                  {{ vol.upperDir }}
                </td>
                <td class="px-4 py-3 text-slate-300 font-bold">
                  {{ vol.size }}
                </td>
                <td class="px-4 py-3 text-slate-500">
                  {{ vol.created }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
