<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Layers, 
  HardDrive, 
  FileText, 
  Trash2, 
  Edit3, 
  Plus, 
  Check, 
  RefreshCw, 
  Eye, 
  FileCode, 
  ArrowUp, 
  Folder, 
  Info,
  Database,
  Box,
  Layers3,
  Cpu,
  ShieldCheck,
  ChevronDown,
  Sparkles,
  Zap,
  ArrowRight
} from 'lucide-vue-next';

interface LayerMeta {
  id: string;
  name: string;
  type: 'writable' | 'readonly';
  technology: 'AUFS / Overlay2' | 'Overlay2 lowerdir';
  dockerDirective: string;
  size: string;
  badgeColor: string;
  borderColor: string;
  bgColor: string;
  description: string;
  files: {
    name: string;
    path: string;
    status: 'added' | 'modified' | 'deleted' | 'base';
    content: string;
  }[];
}

// Preset Dockerfile architectures to demonstrate
type ImagePreset = 'nginx' | 'node' | 'python';
const selectedPreset = ref<ImagePreset>('nginx');

// Active selected layer in visual stack
const selectedLayerId = ref<string>('layer-writable');

// Stack model representing AUFS / Overlay2 layers from top (writable) to bottom (base kernel/rootfs)
const stackLayers = ref<LayerMeta[]>([
  {
    id: 'layer-writable',
    name: 'Container Layer (RW)',
    type: 'writable',
    technology: 'AUFS / Overlay2',
    dockerDirective: 'Runtime Container Diff (upperdir)',
    size: '1.4 MB',
    badgeColor: 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40',
    borderColor: 'border-emerald-500',
    bgColor: 'bg-emerald-950/20',
    description: 'The ephemeral read-write container layer created on `docker run`. Captures runtime changes via Copy-on-Write (CoW). Deleted when container is removed unless committed.',
    files: [
      {
        name: 'nginx.conf',
        path: '/etc/nginx/nginx.conf',
        status: 'modified',
        content: 'worker_processes auto;\nevents { worker_connections 2048; }\n# MODIFIED in container layer!\nhttp { server { listen 80; } }'
      },
      {
        name: 'app.log',
        path: '/var/log/nginx/app.log',
        status: 'added',
        content: '[2026-09-24 12:35:00] Worker process initialized inside container namespace.'
      },
      {
        name: '.wh.ping',
        path: '/bin/.wh.ping',
        status: 'deleted',
        content: 'AUFS / Overlay2 whiteout device file (0:0). Hides /bin/ping in read-only lower layers.'
      }
    ]
  },
  {
    id: 'layer-app',
    name: 'Application Assets & Code',
    type: 'readonly',
    technology: 'Overlay2 lowerdir',
    dockerDirective: 'COPY . /usr/share/nginx/html',
    size: '4.8 MB',
    badgeColor: 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/40',
    borderColor: 'border-cyan-500',
    bgColor: 'bg-cyan-950/20',
    description: 'Image layer containing built static assets and web application code. Immutable sha256 blob cached on host disk.',
    files: [
      {
        name: 'index.html',
        path: '/usr/share/nginx/html/index.html',
        status: 'base',
        content: '<!DOCTYPE html>\n<html>\n<head><title>Docker Union FS</title></head>\n<body><h1>Hello from Layer 3!</h1></body>\n</html>'
      },
      {
        name: 'bundle.js',
        path: '/usr/share/nginx/html/bundle.js',
        status: 'base',
        content: 'console.log("App loaded from read-only image layer sha256:7b92...");'
      }
    ]
  },
  {
    id: 'layer-package',
    name: 'Installed Binaries (Nginx Engine)',
    type: 'readonly',
    technology: 'Overlay2 lowerdir',
    dockerDirective: 'RUN apk add --no-cache nginx',
    size: '18.2 MB',
    badgeColor: 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/40',
    borderColor: 'border-indigo-500',
    bgColor: 'bg-indigo-950/20',
    description: 'Package manager layer containing binaries, shared libraries (.so), and default configuration files created during image build.',
    files: [
      {
        name: 'nginx',
        path: '/usr/sbin/nginx',
        status: 'base',
        content: 'ELF 64-bit LSB pie executable, x86-64, dynamically linked.'
      },
      {
        name: 'nginx.conf.default',
        path: '/etc/nginx/nginx.conf',
        status: 'base',
        content: '# Default upstream template installed by apk'
      }
    ]
  },
  {
    id: 'layer-base',
    name: 'Base OS RootFS (Alpine Linux 3.19)',
    type: 'readonly',
    technology: 'Overlay2 lowerdir',
    dockerDirective: 'FROM alpine:3.19',
    size: '7.3 MB',
    badgeColor: 'bg-purple-500/20 text-purple-400 border border-purple-500/40',
    borderColor: 'border-purple-500',
    bgColor: 'bg-purple-950/20',
    description: 'The foundation layer containing the minimal userland rootfs (busybox, musl libc, /etc, /bin, /dev, /proc mount points). Shared read-only by multiple containers.',
    files: [
      {
        name: 'os-release',
        path: '/etc/os-release',
        status: 'base',
        content: 'NAME="Alpine Linux"\nID=alpine\nVERSION_ID=3.19.1\nPRETTY_NAME="Alpine Linux v3.19"'
      },
      {
        name: 'sh',
        path: '/bin/sh',
        status: 'base',
        content: 'BusyBox v1.36.1 multi-call binary.'
      },
      {
        name: 'ping',
        path: '/bin/ping',
        status: 'base',
        content: 'BusyBox ping utility (hidden in top container view via AUFS/Overlay2 whiteout).'
      }
    ]
  }
]);

// Currently selected layer
const activeLayer = computed(() => {
  return stackLayers.value.find(l => l.id === selectedLayerId.value) || stackLayers.value[0];
});

// Union filesystem engine selection
const unionDriver = ref<'overlay2' | 'aufs'>('overlay2');

// Interactive file view modal/selection
const inspectingFile = ref<any | null>(stackLayers.value[0].files[0]);

// Simulation of Copy-on-Write action
const isPerformingCow = ref(false);
const cowNotification = ref<string | null>(null);

function triggerCopyOnWrite(file: { name: string; path: string; content: string }) {
  isPerformingCow.value = true;
  cowNotification.value = `[Copy-on-Write Triggered] File '${file.path}' requested for edit. Linux VFS duplicate-copied bytes from lowerdir to upperdir!`;
  
  setTimeout(() => {
    // Check if already in writable layer
    const writable = stackLayers.value[0];
    const exists = writable.files.find(f => f.path === file.path);
    if (!exists) {
      writable.files.unshift({
        name: file.name,
        path: file.path,
        status: 'modified',
        content: `# [COW DUPLICATE] Copied from read-only lowerdir at ${new Date().toLocaleTimeString()}\n` + file.content
      });
    }
    selectedLayerId.value = 'layer-writable';
    inspectingFile.value = writable.files[0];
    isPerformingCow.value = false;
  }, 600);
}

function triggerWhiteout(file: { name: string; path: string }) {
  cowNotification.value = `[AUFS/Overlay2 Whiteout] Created character device 0:0 (.wh.${file.name}) in upperdir. Underlying file '${file.path}' is completely masked from container view!`;
  const writable = stackLayers.value[0];
  writable.files.push({
    name: `.wh.${file.name}`,
    path: file.path,
    status: 'deleted',
    content: `Character Device 0:0 (.wh.${file.name}) masking file from lower base image.`
  });
  selectedLayerId.value = 'layer-writable';
  inspectingFile.value = writable.files[writable.files.length - 1];
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Layers3 class="w-5 h-5 text-amber-400" />
          <h2 class="text-xl font-bold text-white font-sans">
            Docker Union Filesystem (AUFS & Overlay2) Visual Stack
          </h2>
          <span class="text-xs font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30">
            CoW Storage Driver
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Explore how Docker stacks multiple read-only image layers (<code class="text-cyan-400">lowerdir</code>) under a single thin writable container layer (<code class="text-emerald-400">upperdir</code>) to form a unified merged view at <code class="text-white">/</code>.
        </p>
      </div>

      <!-- Union Engine Toggle -->
      <div class="flex items-center gap-2">
        <label class="text-xs font-mono text-slate-400">Storage Driver:</label>
        <div class="flex items-center p-1 bg-[#141A22] border border-[#232A35] rounded-lg text-xs font-mono">
          <button 
            @click="unionDriver = 'overlay2'"
            class="px-2.5 py-1 rounded transition-colors cursor-pointer"
            :class="unionDriver === 'overlay2' ? 'bg-[#1D63ED] text-white font-bold' : 'text-slate-400 hover:text-white'"
          >
            Overlay2 (Default)
          </button>
          <button 
            @click="unionDriver = 'aufs'"
            class="px-2.5 py-1 rounded transition-colors cursor-pointer"
            :class="unionDriver === 'aufs' ? 'bg-[#1D63ED] text-white font-bold' : 'text-slate-400 hover:text-white'"
          >
            AUFS (Legacy)
          </button>
        </div>
      </div>
    </div>

    <!-- Notification Banner for CoW Actions -->
    <div v-if="cowNotification" class="p-3 rounded-lg bg-[#141F2D] border border-cyan-800/80 text-xs font-mono text-cyan-300 flex items-center justify-between">
      <div class="flex items-center gap-2 truncate">
        <Zap class="w-4 h-4 text-amber-400 shrink-0 animate-bounce" />
        <span class="truncate">{{ cowNotification }}</span>
      </div>
      <button 
        @click="cowNotification = null"
        class="text-slate-400 hover:text-white text-xs px-2 py-0.5 rounded bg-[#1C2738] cursor-pointer shrink-0 ml-2"
      >
        Dismiss
      </button>
    </div>

    <!-- MAIN TWO-COLUMN VISUAL STACK VIEW -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: ISOMETRIC / VERTICAL LAYER CAKE -->
      <div class="lg:col-span-7 space-y-4">
        
        <!-- Unified Merged Mount Output Banner (Top of Stack) -->
        <div class="p-3.5 rounded-xl bg-[#121822] border-2 border-[#1D63ED] shadow-md flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <Box class="w-5 h-5 text-[#0db7ed]" />
            <div>
              <div class="text-xs font-bold text-white font-sans flex items-center gap-2">
                <span>Unified Container View: <code class="text-[#0db7ed]">/merged</code> (pivot_root)</span>
                <span class="text-[10px] px-1.5 py-0.2 rounded bg-[#1D63ED]/20 text-[#0db7ed] font-mono">
                  Container Mount Point
                </span>
              </div>
              <div class="text-[11px] text-slate-400 font-mono mt-0.5">
                VFS resolves lookups top-to-bottom: Upper layer shadows lower layers.
              </div>
            </div>
          </div>
          <ArrowRight class="w-4 h-4 text-[#0db7ed] shrink-0" />
        </div>

        <!-- STACK OF LAYERS (Top = Writable, Bottom = Base Image) -->
        <div class="space-y-3 relative">
          
          <div 
            v-for="(layer, idx) in stackLayers" 
            :key="layer.id"
            @click="selectedLayerId = layer.id; inspectingFile = layer.files[0]"
            class="p-4 rounded-xl border-2 transition-all cursor-pointer relative overflow-hidden select-none"
            :class="[
              layer.id === selectedLayerId 
                ? `${layer.borderColor} ring-2 ring-opacity-50 shadow-lg ${layer.bgColor}` 
                : 'border-[#232A35] bg-[#141A22] hover:border-slate-500'
            ]"
          >
            <!-- Layer Type Badge on right -->
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <span 
                  class="w-6 h-6 rounded-lg flex items-center justify-center font-mono text-xs font-bold"
                  :class="layer.type === 'writable' ? 'bg-emerald-500 text-black' : 'bg-[#1E2633] text-slate-300'"
                >
                  {{ stackLayers.length - idx }}
                </span>
                <div>
                  <div class="text-sm font-bold text-white font-sans flex items-center gap-2">
                    <span>{{ layer.name }}</span>
                    <span 
                      class="text-[10px] px-2 py-0.5 rounded font-mono font-medium"
                      :class="layer.badgeColor"
                    >
                      {{ layer.type === 'writable' ? 'READ-WRITE' : 'READ-ONLY' }}
                    </span>
                  </div>
                  <div class="text-xs font-mono text-slate-400 mt-0.5">
                    Directive: <span class="text-cyan-300">{{ layer.dockerDirective }}</span>
                  </div>
                </div>
              </div>

              <!-- Size & Driver Inode -->
              <div class="text-right font-mono text-xs">
                <div class="text-white font-bold">{{ layer.size }}</div>
                <div class="text-[10px] text-slate-500">{{ layer.technology }}</div>
              </div>
            </div>

            <!-- Description -->
            <p class="text-xs text-slate-400 font-sans leading-relaxed mb-3">
              {{ layer.description }}
            </p>

            <!-- Layer Files Preview Bar -->
            <div class="flex flex-wrap items-center gap-1.5 pt-2 border-t border-[#232A35]">
              <span class="text-[10px] text-slate-500 font-mono uppercase mr-1">Layer Contents:</span>
              <span 
                v-for="file in layer.files" 
                :key="file.path"
                class="px-2 py-0.5 rounded text-[11px] font-mono flex items-center gap-1"
                :class="file.status === 'modified' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40' : file.status === 'deleted' ? 'bg-rose-500/20 text-rose-300 line-through border border-rose-500/40' : file.status === 'added' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40' : 'bg-[#1A222D] text-slate-300 border border-[#2D3848]'"
              >
                <FileText class="w-3 h-3" />
                <span>{{ file.name }}</span>
              </span>
            </div>
          </div>

        </div>

        <!-- Union Mount Syscall Behind the Stack -->
        <div class="p-3 rounded-xl bg-[#0E1217] border border-[#232A35] font-mono text-xs space-y-1">
          <div class="text-[10px] text-slate-500 uppercase font-bold flex items-center justify-between">
            <span>Linux Kernel Union Mount Syscall Behind This Stack:</span>
            <span class="text-emerald-400 font-bold">VFS Overlay2</span>
          </div>
          <code class="text-cyan-300 block break-all text-[11px]">
            mount -t overlay overlay -o lowerdir=/layer4:/layer3:/layer2,upperdir=/layer1/diff,workdir=/layer1/work /merged
          </code>
        </div>
      </div>

      <!-- RIGHT 5 COLS: DEEP LAYER INSPECTION & COPY-ON-WRITE TESTER -->
      <div class="lg:col-span-5 space-y-4">
        
        <!-- Active Layer Details Card -->
        <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-5 space-y-4">
          <div class="border-b border-[#232A35] pb-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-mono uppercase text-slate-400 font-bold">Inspecting Layer</span>
              <span 
                class="text-[10px] font-mono px-2 py-0.5 rounded font-bold"
                :class="activeLayer.badgeColor"
              >
                {{ activeLayer.type.toUpperCase() }}
              </span>
            </div>
            <h3 class="text-base font-bold text-white font-sans mt-1">{{ activeLayer.name }}</h3>
            <div class="text-xs font-mono text-cyan-300 mt-0.5">{{ activeLayer.dockerDirective }}</div>
          </div>

          <!-- Files in this layer with interactive CoW buttons -->
          <div class="space-y-2">
            <div class="text-xs font-mono text-slate-400 uppercase font-bold">Files In This Layer:</div>
            
            <div 
              v-for="file in activeLayer.files" 
              :key="file.path"
              class="p-2.5 rounded-lg bg-[#0E1217] border border-[#232A35] space-y-2"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-1.5 truncate">
                  <FileCode class="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                  <span class="text-xs font-mono font-bold text-white truncate">{{ file.path }}</span>
                </div>
                <span 
                  class="text-[9px] px-1.5 py-0.2 rounded font-mono font-bold shrink-0 ml-1"
                  :class="file.status === 'modified' ? 'bg-amber-500/20 text-amber-300' : file.status === 'deleted' ? 'bg-rose-500/20 text-rose-300' : file.status === 'added' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-slate-800 text-slate-400'"
                >
                  {{ file.status.toUpperCase() }}
                </span>
              </div>

              <!-- Content Preview -->
              <pre class="p-2 rounded bg-[#0A0D12] text-[10px] font-mono text-slate-300 overflow-x-auto max-h-24 leading-relaxed border border-[#1A222D]"><code>{{ file.content }}</code></pre>

              <!-- Interactive CoW Action Triggers -->
              <div v-if="activeLayer.type === 'readonly'" class="flex items-center justify-end gap-2 pt-1">
                <button 
                  @click="triggerCopyOnWrite(file)"
                  class="px-2 py-1 rounded text-[10px] font-mono bg-amber-950/60 hover:bg-amber-900 text-amber-300 border border-amber-800/80 cursor-pointer flex items-center gap-1 transition-colors"
                  title="Simulate editing this file (Triggers Copy-on-Write duplicate to upperdir)"
                >
                  <Edit3 class="w-3 h-3" />
                  <span>Modify (CoW)</span>
                </button>
                <button 
                  @click="triggerWhiteout(file)"
                  class="px-2 py-1 rounded text-[10px] font-mono bg-rose-950/60 hover:bg-rose-900 text-rose-300 border border-rose-800/80 cursor-pointer flex items-center gap-1 transition-colors"
                  title="Simulate deleting this file (Creates Overlay2 whiteout device file)"
                >
                  <Trash2 class="w-3 h-3" />
                  <span>Delete (Whiteout)</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- How AUFS vs Overlay2 Compares -->
        <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4 space-y-2 text-xs font-sans">
          <div class="flex items-center gap-2 font-bold text-white font-mono">
            <Info class="w-4 h-4 text-cyan-400" />
            <span>AUFS vs. Overlay2 Architecture</span>
          </div>
          <div class="text-slate-400 leading-relaxed text-[11px] space-y-1">
            <p>
              • <strong class="text-white">AUFS (AnotherUnionFS)</strong>: Used in Docker 1.x. Handled layers via branch filesystems. Limited by kernel patches and high inode consumption.
            </p>
            <p>
              • <strong class="text-white">Overlay2</strong>: Linux kernel standard (3.18+). Uses native kernel VFS primitives with direct dentry caching, faster page cache sharing, and up to 128 stacked lower layers without performance degradation.
            </p>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>
