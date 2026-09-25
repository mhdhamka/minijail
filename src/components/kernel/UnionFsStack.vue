<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Layers3, 
  FileText, 
  Trash2, 
  Edit3, 
  FileCode, 
  Box, 
  Info,
  Zap,
  ArrowRight,
} from 'lucide-vue-next';

interface LayerFile {
  name: string;
  path: string;
  status: 'added' | 'modified' | 'deleted' | 'base';
  content: string;
}

interface LayerMeta {
  id: string;
  name: string;
  type: 'writable' | 'readonly';
  technology: 'AUFS / Overlay2' | 'Overlay2 lowerdir';
  dockerDirective: string;
  size: string;
  badgeStyle: string;
  borderColor: string;
  bgColor: string;
  description: string;
  files: LayerFile[];
}

type ImagePreset = 'nginx' | 'node' | 'python';
const selectedPreset = ref<ImagePreset>('nginx');
const selectedLayerId = ref<string>('layer-writable');
const unionDriver = ref<'overlay2' | 'aufs'>('overlay2');
const cowNotification = ref<string | null>(null);

const stackLayers = ref<LayerMeta[]>([
  {
    id: 'layer-writable',
    name: 'Container Layer (RW)',
    type: 'writable',
    technology: 'AUFS / Overlay2',
    dockerDirective: 'Runtime Container Diff (upperdir)',
    size: '1.4 MB',
    badgeStyle: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
    borderColor: 'border-emerald-500/80',
    bgColor: 'bg-emerald-950/10',
    description: 'Ephemeral read-write container layer created on `docker run`. Captures runtime mutations via Copy-on-Write (CoW).',
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
        content: 'AUFS / Overlay2 whiteout device file (0:0). Masks /bin/ping in lower layers.'
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
    badgeStyle: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30',
    borderColor: 'border-cyan-500/40',
    bgColor: 'bg-cyan-950/10',
    description: 'Immutable image layer containing built static assets and compiled application code.',
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
    badgeStyle: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30',
    borderColor: 'border-indigo-500/40',
    bgColor: 'bg-indigo-950/10',
    description: 'System package layer containing compiled binaries, shared libraries (.so), and configs.',
    files: [
      {
        name: 'nginx',
        path: '/usr/sbin/nginx',
        status: 'base',
        content: 'ELF 64-bit LSB pie executable, x86-64, dynamically linked.'
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
    badgeStyle: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
    borderColor: 'border-purple-500/40',
    bgColor: 'bg-purple-950/10',
    description: 'Foundation userland rootfs (busybox, musl libc, core mount points) shared globally across containers.',
    files: [
      {
        name: 'os-release',
        path: '/etc/os-release',
        status: 'base',
        content: 'NAME="Alpine Linux"\nID=alpine\nVERSION_ID=3.19.1'
      },
      {
        name: 'ping',
        path: '/bin/ping',
        status: 'base',
        content: 'BusyBox ping utility.'
      }
    ]
  }
]);

const activeLayer = computed(() => {
  return stackLayers.value.find(l => l.id === selectedLayerId.value) || stackLayers.value[0];
});

function triggerCopyOnWrite(file: LayerFile) {
  cowNotification.value = `[CoW Triggered] Copied '${file.path}' from lowerdir to upperdir container layer.`;
  const writable = stackLayers.value[0];
  if (!writable.files.some(f => f.path === file.path)) {
    writable.files.unshift({
      name: file.name,
      path: file.path,
      status: 'modified',
      content: `# [COW DUPLICATED] Editable clone at ${new Date().toLocaleTimeString()}\n` + file.content
    });
  }
  selectedLayerId.value = 'layer-writable';
}

function triggerWhiteout(file: LayerFile) {
  cowNotification.value = `[Whiteout Created] Character device 0:0 (.wh.${file.name}) injected to shadow lower layer path.`;
  const writable = stackLayers.value[0];
  writable.files.push({
    name: `.wh.${file.name}`,
    path: file.path,
    status: 'deleted',
    content: `Whiteout marker masking base file from unified view.`
  });
  selectedLayerId.value = 'layer-writable';
}
</script>

<template>
  <div class="space-y-6 select-none font-sans text-slate-200">
    <!-- Header Component -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-5 rounded-2xl bg-[#0d1322]/80 border border-slate-800/80 backdrop-blur-xl shadow-xl">
      <div class="space-y-1.5">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
            <Layers3 class="w-5 h-5" />
          </div>
          <h2 class="text-base font-bold text-white tracking-tight">
            Docker Union Filesystem Visualizer
          </h2>
          <span class="text-[10px] px-2.5 py-0.5 rounded-full font-mono bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium">
            Storage Engine Simulator
          </span>
        </div>
        <p class="text-xs text-slate-400 max-w-2xl leading-relaxed">
          Examine how container engines stack immutable read-only image layers (<code class="text-cyan-400 font-mono">lowerdir</code>) beneath a lightweight volatile execution layer (<code class="text-emerald-400 font-mono">upperdir</code>).
        </p>
      </div>

      <!-- Driver Switcher -->
      <div class="flex items-center gap-2 bg-[#090d16] p-1.5 rounded-xl border border-slate-800 font-mono text-xs shrink-0">
        <button 
          @click="unionDriver = 'overlay2'"
          class="px-3 py-1.5 rounded-lg transition-all cursor-pointer font-medium"
          :class="unionDriver === 'overlay2' ? 'bg-blue-600 text-white shadow-md shadow-blue-900/30' : 'text-slate-400 hover:text-slate-200'"
        >
          Overlay2
        </button>
        <button 
          @click="unionDriver = 'aufs'"
          class="px-3 py-1.5 rounded-lg transition-all cursor-pointer font-medium"
          :class="unionDriver === 'aufs' ? 'bg-blue-600 text-white shadow-md shadow-blue-900/30' : 'text-slate-400 hover:text-slate-200'"
        >
          AUFS <span class="opacity-60">(Legacy)</span>
        </button>
      </div>
    </div>

    <!-- Notification Alert -->
    <div v-if="cowNotification" class="p-3.5 rounded-xl bg-blue-950/40 border border-blue-500/30 text-xs font-mono text-blue-200 flex items-center justify-between backdrop-blur-md shadow-lg animate-fade-in">
      <div class="flex items-center gap-2.5 truncate">
        <Zap class="w-4 h-4 text-amber-400 shrink-0" />
        <span class="truncate">{{ cowNotification }}</span>
      </div>
      <button 
        @click="cowNotification = null"
        class="text-slate-400 hover:text-white text-xs px-2 py-1 rounded-md bg-blue-900/40 cursor-pointer shrink-0 ml-3 transition-colors"
      >
        Dismiss
      </button>
    </div>

    <!-- Main Workspace Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- Left Column: Layer Stack Cake -->
      <div class="lg:col-span-7 space-y-4">
        
        <!-- Merged Viewpoint Header -->
        <div class="p-4 rounded-xl bg-gradient-to-r from-blue-950/40 via-[#0d1322] to-[#0d1322] border border-blue-500/40 shadow-md flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
              <Box class="w-4 h-4" />
            </div>
            <div>
              <div class="text-xs font-bold text-white flex items-center gap-2 font-mono">
                <span>Unified Mount: <code class="text-cyan-400">/merged</code></span>
                <span class="text-[10px] px-2 py-0.5 rounded bg-blue-500/10 text-blue-300 border border-blue-500/20">
                  Container Viewport
                </span>
              </div>
              <div class="text-[11px] text-slate-400 font-mono mt-0.5">
                VFS kernel file lookups resolve top-down. Upper layers shadow lower matches.
              </div>
            </div>
          </div>
          <ArrowRight class="w-4 h-4 text-blue-400 shrink-0" />
        </div>

        <!-- Stack Elements -->
        <div class="space-y-3">
          <div 
            v-for="(layer, idx) in stackLayers" 
            :key="layer.id"
            @click="selectedLayerId = layer.id"
            class="p-4 rounded-xl border transition-all cursor-pointer relative overflow-hidden backdrop-blur-sm group"
            :class="[
              layer.id === selectedLayerId 
                ? `${layer.borderColor} ring-1 ring-blue-500/50 shadow-lg shadow-black/40 ${layer.bgColor}` 
                : 'border-slate-800/80 bg-[#0d1322]/60 hover:border-slate-700 hover:bg-[#0d1322]'
            ]"
          >
            <div class="flex items-center justify-between mb-2.5">
              <div class="flex items-center gap-3">
                <span 
                  class="w-6 h-6 rounded-lg flex items-center justify-center font-mono text-xs font-bold shadow-inner"
                  :class="layer.type === 'writable' ? 'bg-emerald-500 text-slate-950' : 'bg-slate-800 text-slate-300'"
                >
                  {{ stackLayers.length - idx }}
                </span>
                <div>
                  <div class="text-sm font-bold text-white flex items-center gap-2">
                    <span>{{ layer.name }}</span>
                    <span class="text-[10px] px-2 py-0.5 rounded-md font-mono border font-medium" :class="layer.badgeStyle">
                      {{ layer.type === 'writable' ? 'READ-WRITE' : 'READ-ONLY' }}
                    </span>
                  </div>
                  <div class="text-xs font-mono text-slate-400 mt-0.5">
                    Directive: <span class="text-cyan-300 font-medium">{{ layer.dockerDirective }}</span>
                  </div>
                </div>
              </div>

              <div class="text-right font-mono">
                <div class="text-xs font-bold text-slate-200">{{ layer.size }}</div>
                <div class="text-[10px] text-slate-500">{{ layer.technology }}</div>
              </div>
            </div>

            <p class="text-xs text-slate-400 leading-relaxed mb-3">
              {{ layer.description }}
            </p>

            <div class="flex flex-wrap items-center gap-1.5 pt-2.5 border-t border-slate-800/60">
              <span class="text-[10px] text-slate-500 font-mono uppercase tracking-wider mr-1">Contents:</span>
              <span 
                v-for="file in layer.files" 
                :key="file.path"
                class="px-2 py-0.5 rounded text-[11px] font-mono flex items-center gap-1 border"
                :class="file.status === 'modified' ? 'bg-amber-500/10 text-amber-300 border-amber-500/30' : file.status === 'deleted' ? 'bg-rose-500/10 text-rose-300 line-through border-rose-500/30' : file.status === 'added' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-slate-900/60 text-slate-300 border-slate-800'"
              >
                <FileText class="w-3 h-3 opacity-70" />
                <span>{{ file.name }}</span>
              </span>
            </div>
          </div>
        </div>

        <!-- Kernel Mount Command Box -->
        <div class="p-3.5 rounded-xl bg-[#090d16] border border-slate-800 font-mono text-xs space-y-1.5 shadow-inner">
          <div class="text-[10px] text-slate-500 uppercase font-bold flex items-center justify-between">
            <span>Underlying Kernel VFS Mount Call:</span>
            <span class="text-emerald-400">Active</span>
          </div>
          <code class="text-cyan-300 block break-all text-[11px] leading-relaxed">
            mount -t overlay overlay -o lowerdir=/layer4:/layer3:/layer2,upperdir=/layer1/diff,workdir=/layer1/work /merged
          </code>
        </div>
      </div>

      <!-- Right Column: Layer Inspection & CoW Simulator -->
      <div class="lg:col-span-5 space-y-4">
        <div class="bg-[#0d1322]/80 border border-slate-800/80 rounded-2xl p-5 space-y-4 backdrop-blur-xl shadow-xl">
          <div class="border-b border-slate-800 pb-3.5">
            <div class="flex items-center justify-between">
              <span class="text-[10px] font-mono uppercase tracking-wider text-slate-400 font-bold">Active Inspection Focus</span>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded border font-bold" :class="activeLayer.badgeStyle">
                {{ activeLayer.type.toUpperCase() }}
              </span>
            </div>
            <h3 class="text-sm font-bold text-white mt-1">{{ activeLayer.name }}</h3>
            <div class="text-xs font-mono text-cyan-300 mt-0.5 truncate">{{ activeLayer.dockerDirective }}</div>
          </div>

          <div class="space-y-3">
            <div class="text-[11px] font-mono text-slate-400 uppercase tracking-wider font-bold">Layer File Nodes:</div>
            
            <div 
              v-for="file in activeLayer.files" 
              :key="file.path"
              class="p-3 rounded-xl bg-[#090d16] border border-slate-800/80 space-y-2.5 shadow-sm"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 truncate">
                  <FileCode class="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                  <span class="text-xs font-mono font-bold text-slate-200 truncate">{{ file.path }}</span>
                </div>
                <span 
                  class="text-[9px] px-2 py-0.5 rounded font-mono font-bold shrink-0 ml-2"
                  :class="file.status === 'modified' ? 'bg-amber-500/10 text-amber-300 border border-amber-500/20' : file.status === 'deleted' ? 'bg-rose-500/10 text-rose-300 border border-rose-500/20' : file.status === 'added' ? 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/20' : 'bg-slate-800/80 text-slate-400'"
                >
                  {{ file.status.toUpperCase() }}
                </span>
              </div>

              <pre class="p-2.5 rounded-lg bg-[#06090f] text-[10px] font-mono text-slate-300 overflow-x-auto max-h-28 leading-relaxed border border-slate-900"><code>{{ file.content }}</code></pre>

              <!-- CoW Action triggers for read-only items -->
              <div v-if="activeLayer.type === 'readonly'" class="flex items-center justify-end gap-2 pt-1">
                <button 
                  @click="triggerCopyOnWrite(file)"
                  class="px-2.5 py-1 rounded-lg text-[10px] font-mono bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 cursor-pointer flex items-center gap-1.5 transition-all"
                  title="Triggers Copy-on-Write duplicate to writable upperdir"
                >
                  <Edit3 class="w-3 h-3" />
                  <span>Modify (CoW)</span>
                </button>
                <button 
                  @click="triggerWhiteout(file)"
                  class="px-2.5 py-1 rounded-lg text-[10px] font-mono bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/30 cursor-pointer flex items-center gap-1.5 transition-all"
                  title="Creates an Overlay2 whiteout device file to mask this base asset"
                >
                  <Trash2 class="w-3 h-3" />
                  <span>Delete (Whiteout)</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Architecture Quick Info Box -->
        <div class="bg-[#0d1322]/80 border border-slate-800/80 rounded-2xl p-4 space-y-2 text-xs backdrop-blur-xl shadow-xl">
          <div class="flex items-center gap-2 font-bold text-white font-mono">
            <Info class="w-4 h-4 text-cyan-400" />
            <span>AUFS vs. Overlay2 Core Mechanics</span>
          </div>
          <div class="text-slate-400 leading-relaxed text-[11px] space-y-1.5">
            <p>
              • <strong class="text-slate-200">AUFS</strong>: Historical union driver utilizing multi-branch filesystem paths, prone to high inode usage on massive trees.
            </p>
            <p>
              • <strong class="text-slate-200">Overlay2</strong>: Modern Linux VFS standard supporting fast page-cache sharing and up to 128 stacked lower layers cleanly.
            </p>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>