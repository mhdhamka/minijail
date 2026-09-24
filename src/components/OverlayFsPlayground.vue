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
  ShieldAlert, 
  Sparkles, 
  Eye, 
  FileCode, 
  ArrowUp, 
  Folder, 
  Info,
  Database
} from 'lucide-vue-next';

interface LayerFile {
  path: string;
  size: string;
  action?: 'base' | 'modified' | 'deleted' | 'added';
  content?: string;
  isWhiteout?: boolean;
}

interface ImageLayer {
  id: string;
  name: string;
  type: 'lower' | 'upper';
  readOnly: boolean;
  sizeMb: number;
  dockerCommand: string;
  files: LayerFile[];
}

const currentPreset = ref<'nginx' | 'node' | 'python'>('nginx');

// Sample file contents for realistic inspection
const baseFiles = ref<LayerFile[]>([
  { path: '/etc/os-release', size: '120 B', action: 'base', content: 'NAME="Alpine Linux"\nVERSION_ID=3.19.0\nID=alpine' },
  { path: '/etc/hosts', size: '240 B', action: 'base', content: '127.0.0.1 localhost\n::1 localhost ip6-localhost' },
  { path: '/etc/nginx/nginx.conf', size: '1.2 KB', action: 'base', content: 'worker_processes auto;\nevents { worker_connections 1024; }\nhttp { server { listen 80; } }' },
  { path: '/usr/share/nginx/html/index.html', size: '615 B', action: 'base', content: '<h1>Welcome to nginx in Docker!</h1>\n<p>Served from lowerdir layer.</p>' },
  { path: '/bin/sh', size: '110 KB', action: 'base' },
  { path: '/bin/ping', size: '42 KB', action: 'base' },
  { path: '/lib/ld-musl-x86_64.so.1', size: '620 KB', action: 'base' },
]);

// Upperdir files (container writable layer mutations)
const upperFiles = ref<LayerFile[]>([
  { 
    path: '/etc/nginx/nginx.conf', 
    size: '1.4 KB', 
    action: 'modified', 
    content: 'worker_processes auto;\n# MODIFIED IN CONTAINER LAYER!\nserver { listen 80; server_name app.local; }' 
  },
  { 
    path: '/usr/share/nginx/html/app.js', 
    size: '2.1 KB', 
    action: 'added', 
    content: 'console.log("Interactive Docker OverlayFS App Loaded!");' 
  },
  { 
    path: '/bin/ping', 
    size: '0 B', 
    action: 'deleted', 
    isWhiteout: true, 
    content: 'OverlayFS Character Device 0:0 (.wh.ping) - Hides lower file!' 
  },
  { 
    path: '/var/log/nginx/access.log', 
    size: '14.2 KB', 
    action: 'added', 
    content: '172.18.0.1 - - [24/Sep/2026:12:00:01] "GET / HTTP/1.1" 200' 
  },
]);

// Interactive action inputs
const newFilePath = ref('/app/config.json');
const newFileContent = ref('{\n  "env": "production",\n  "cow_demo": true\n}');
const selectedFile = ref<LayerFile | null>(upperFiles.value[0]);
const activeLayerFilter = ref<'all' | 'merged' | 'upper' | 'lower'>('merged');

// Computed merged view
const mergedFiles = computed(() => {
  const map = new Map<string, LayerFile>();

  // 1. Add all lower files
  for (const f of baseFiles.value) {
    map.set(f.path, { ...f, action: 'base' });
  }

  // 2. Overlay upper files
  for (const uf of upperFiles.value) {
    if (uf.isWhiteout || uf.action === 'deleted') {
      // Deleted file: mark as whiteout/hidden in merged view
      map.delete(uf.path);
    } else {
      map.set(uf.path, { ...uf });
    }
  }

  return Array.from(map.values());
});

// Storage savings calculator
const sharedImageSizeMb = 52.4;
const containerCount = ref(8);
const avgUpperSizeMb = 0.85;

const standardVmDiskUsageMb = computed(() => {
  return +(containerCount.value * sharedImageSizeMb).toFixed(1);
});

const overlayFsDiskUsageMb = computed(() => {
  return +(sharedImageSizeMb + (containerCount.value * avgUpperSizeMb)).toFixed(1);
});

const storageSavedPercent = computed(() => {
  if (standardVmDiskUsageMb.value === 0) return 0;
  const saved = ((standardVmDiskUsageMb.value - overlayFsDiskUsageMb.value) / standardVmDiskUsageMb.value) * 100;
  return +saved.toFixed(1);
});

// Simulation actions
function handleModifyFile(path: string) {
  const existing = upperFiles.value.find(f => f.path === path);
  if (existing) {
    existing.action = 'modified';
    existing.content = `// Updated at ${new Date().toLocaleTimeString()}\n` + (existing.content || '');
    selectedFile.value = existing;
  } else {
    const base = baseFiles.value.find(f => f.path === path);
    const newUpper: LayerFile = {
      path,
      size: '1.2 KB',
      action: 'modified',
      content: `/* Copy-on-Write: Cloned from read-only lowerdir */\n` + (base?.content || 'Modified container content'),
    };
    upperFiles.value.push(newUpper);
    selectedFile.value = newUpper;
  }
}

function handleDeleteFile(path: string) {
  // If in upperdir, change to whiteout
  const existingIdx = upperFiles.value.findIndex(f => f.path === path);
  if (existingIdx >= 0) {
    upperFiles.value.splice(existingIdx, 1);
  }
  // Create whiteout device entry
  upperFiles.value.push({
    path,
    size: '0 B',
    action: 'deleted',
    isWhiteout: true,
    content: `OverlayFS whiteout file (.wh.${path.split('/').pop()}) created with character device 0:0. Lower layer file is completely untouched!`,
  });
  selectedFile.value = upperFiles.value[upperFiles.value.length - 1];
}

function handleAddFile() {
  if (!newFilePath.value.trim()) return;
  const exists = upperFiles.value.find(f => f.path === newFilePath.value);
  if (exists) {
    exists.content = newFileContent.value;
    selectedFile.value = exists;
    return;
  }
  const created: LayerFile = {
    path: newFilePath.value.trim(),
    size: `${(newFileContent.value.length / 1024).toFixed(1)} KB`,
    action: 'added',
    content: newFileContent.value,
  };
  upperFiles.value.push(created);
  selectedFile.value = created;
}

function resetPlayground() {
  upperFiles.value = [
    { 
      path: '/etc/nginx/nginx.conf', 
      size: '1.4 KB', 
      action: 'modified', 
      content: 'worker_processes auto;\n# MODIFIED IN CONTAINER LAYER!\nserver { listen 80; server_name app.local; }' 
    },
    { 
      path: '/usr/share/nginx/html/app.js', 
      size: '2.1 KB', 
      action: 'added', 
      content: 'console.log("Interactive Docker OverlayFS App Loaded!");' 
    },
    { 
      path: '/bin/ping', 
      size: '0 B', 
      action: 'deleted', 
      isWhiteout: true, 
      content: 'OverlayFS Character Device 0:0 (.wh.ping) - Hides lower file!' 
    },
  ];
  selectedFile.value = upperFiles.value[0];
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Layers class="w-5 h-5 text-amber-400" />
          <h2 class="text-xl font-bold text-white font-sans">OverlayFS Layer Stacker & Copy-on-Write Playground</h2>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Explore how Docker merges immutable image layers (<code class="text-cyan-400">lowerdir</code>) with container diffs (<code class="text-amber-400">upperdir</code>) using Linux kernel union mounts.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button 
          @click="resetPlayground"
          class="px-3 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 hover:text-white border border-[#2D3848] text-xs font-mono flex items-center gap-1.5 transition-colors cursor-pointer"
        >
          <RefreshCw class="w-3.5 h-3.5 text-cyan-400" />
          Reset Demo State
        </button>
      </div>
    </div>

    <!-- Storage Savings Highlight Banner -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4">
        <div class="flex items-center justify-between text-xs text-slate-400 font-mono">
          <span>Active Containers</span>
          <span class="text-white font-bold">{{ containerCount }} instances</span>
        </div>
        <input 
          type="range" 
          v-model.number="containerCount" 
          min="1" 
          max="30" 
          class="w-full mt-2 accent-amber-400 cursor-pointer h-1.5 bg-[#202835] rounded-lg appearance-none"
        />
        <div class="text-[11px] text-slate-500 mt-2 font-sans">
          All {{ containerCount }} containers share 1 single read-only base layer on disk.
        </div>
      </div>

      <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4">
        <div class="text-xs text-slate-400 font-mono">Virtual Machines (Full Clones)</div>
        <div class="text-xl font-bold font-mono text-rose-400 mt-1">
          {{ standardVmDiskUsageMb }} MB
        </div>
        <div class="text-[11px] text-slate-500 mt-1 font-sans">
          Each VM duplicates the entire 52MB OS disk space.
        </div>
      </div>

      <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4">
        <div class="text-xs text-slate-400 font-mono">Docker OverlayFS Storage</div>
        <div class="text-xl font-bold font-mono text-emerald-400 mt-1 flex items-center gap-2">
          <span>{{ overlayFsDiskUsageMb }} MB</span>
          <span class="text-xs px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 font-mono">
            -{{ storageSavedPercent }}% DISK
          </span>
        </div>
        <div class="text-[11px] text-slate-500 mt-1 font-sans">
          52.4MB base + {{ (containerCount * avgUpperSizeMb).toFixed(1) }}MB container diffs total.
        </div>
      </div>
    </div>

    <!-- MAIN LAYER STACK & COW VISUALIZER -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: THE 3D/LAYERED ANATOMY -->
      <div class="lg:col-span-7 space-y-4">
        
        <!-- 1. MERGED VIEW (What Container Process Sees at '/') -->
        <div class="bg-[#141A22] border border-[#2B384A] rounded-xl p-4 shadow-lg">
          <div class="flex items-center justify-between mb-3 border-b border-[#232A35] pb-2">
            <div class="flex items-center gap-2">
              <Eye class="w-4 h-4 text-emerald-400" />
              <h3 class="text-sm font-bold text-white font-sans">1. Merged View (`/merged`)</h3>
            </div>
            <span class="text-[11px] font-mono px-2 py-0.5 rounded bg-emerald-950/60 text-emerald-400 border border-emerald-800/60">
              Container RootFS
            </span>
          </div>

          <p class="text-xs text-slate-400 mb-3">
            The container process operates strictly on this unified mount point. If a file exists in upper, it shadows lower.
          </p>

          <!-- Files Grid inside Merged -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs font-mono">
            <div 
              v-for="file in mergedFiles" 
              :key="file.path"
              @click="selectedFile = file"
              class="p-2.5 rounded-lg border transition-all cursor-pointer flex items-center justify-between"
              :class="selectedFile?.path === file.path 
                ? 'bg-[#1D63ED]/20 border-[#1D63ED] text-white shadow-sm' 
                : file.action === 'modified' 
                ? 'bg-amber-950/20 border-amber-600/40 text-amber-200' 
                : file.action === 'added'
                ? 'bg-emerald-950/20 border-emerald-600/40 text-emerald-200'
                : 'bg-[#0E1217] border-[#232A35] text-slate-300 hover:border-slate-500'"
            >
              <div class="flex items-center gap-2 truncate">
                <FileText class="w-3.5 h-3.5 shrink-0" :class="file.action === 'modified' ? 'text-amber-400' : file.action === 'added' ? 'text-emerald-400' : 'text-slate-400'" />
                <span class="truncate">{{ file.path }}</span>
              </div>
              <span 
                class="text-[10px] px-1.5 py-0.5 rounded uppercase font-bold shrink-0 ml-1"
                :class="file.action === 'modified' ? 'bg-amber-500/20 text-amber-400' : file.action === 'added' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-800 text-slate-400'"
              >
                {{ file.action === 'modified' ? 'Diff' : file.action === 'added' ? 'New' : 'Base' }}
              </span>
            </div>
          </div>
        </div>

        <!-- 2. UPPERDIR LAYER (Read-Write Container Layer) -->
        <div class="bg-[#181510] border border-amber-500/40 rounded-xl p-4 relative">
          <div class="flex items-center justify-between mb-3 border-b border-amber-500/20 pb-2">
            <div class="flex items-center gap-2">
              <Edit3 class="w-4 h-4 text-amber-400" />
              <h3 class="text-sm font-bold text-amber-300 font-sans">2. Upper Layer (`upperdir`) — Read / Write</h3>
            </div>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/40">
              Ephemeral Container Diff
            </span>
          </div>

          <p class="text-xs text-amber-200/80 mb-3 font-sans">
            Holds all runtime modifications. Deleting a lower file creates a special Linux whiteout file (<code class="text-white">.wh.&lt;name&gt;</code>) here.
          </p>

          <div class="space-y-1.5 text-xs font-mono">
            <div 
              v-for="file in upperFiles" 
              :key="file.path"
              @click="selectedFile = file"
              class="p-2 rounded-lg border transition-all cursor-pointer flex items-center justify-between"
              :class="selectedFile?.path === file.path ? 'ring-1 ring-amber-400 bg-amber-500/10' : 'bg-[#0E1217]' + ' border-[#2B2317]'"
            >
              <div class="flex items-center gap-2 truncate">
                <span 
                  class="w-2 h-2 rounded-full shrink-0" 
                  :class="file.isWhiteout ? 'bg-rose-400' : file.action === 'modified' ? 'bg-amber-400' : 'bg-emerald-400'"
                ></span>
                <span :class="file.isWhiteout ? 'line-through text-rose-400' : 'text-amber-200'" class="truncate">
                  {{ file.path }}
                </span>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <span v-if="file.isWhiteout" class="text-[10px] text-rose-400 font-bold bg-rose-950/60 px-1.5 py-0.5 rounded border border-rose-800">
                  WHITEOUT (0:0)
                </span>
                <span v-else class="text-[10px] text-slate-400 font-mono">{{ file.size }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. LOWERDIR LAYER (Read-Only Base Image Layers) -->
        <div class="bg-[#111722] border border-cyan-800/60 rounded-xl p-4">
          <div class="flex items-center justify-between mb-3 border-b border-cyan-800/40 pb-2">
            <div class="flex items-center gap-2">
              <HardDrive class="w-4 h-4 text-cyan-400" />
              <h3 class="text-sm font-bold text-cyan-300 font-sans">3. Lower Layer (`lowerdir`) — Read Only</h3>
            </div>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-cyan-950/60 text-cyan-400 border border-cyan-800/60">
              Immutable Base Image
            </span>
          </div>

          <p class="text-xs text-slate-400 mb-3 font-sans">
            Base layers are shared across all containers on the host. Read-only at kernel level — any write attempt triggers immediate Copy-on-Write (CoW).
          </p>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs font-mono">
            <div 
              v-for="file in baseFiles" 
              :key="file.path"
              class="p-2 rounded bg-[#0A0D12] border border-[#1E2633] flex items-center justify-between"
            >
              <span class="text-slate-300 truncate">{{ file.path }}</span>
              <div class="flex items-center gap-1.5 shrink-0">
                <!-- Action buttons to test CoW -->
                <button 
                  @click="handleModifyFile(file.path)"
                  class="px-1.5 py-0.5 rounded text-[10px] bg-amber-950/60 hover:bg-amber-900 text-amber-300 border border-amber-800/60 cursor-pointer"
                  title="Modify file (Triggers Copy-on-Write)"
                >
                  Edit (CoW)
                </button>
                <button 
                  @click="handleDeleteFile(file.path)"
                  class="px-1.5 py-0.5 rounded text-[10px] bg-rose-950/60 hover:bg-rose-900 text-rose-300 border border-rose-800/60 cursor-pointer"
                  title="Delete file (Creates OverlayFS whiteout)"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>

      </div>

      <!-- RIGHT 5 COLS: FILE INSPECTOR & COW INTERACTIVE WRITER -->
      <div class="lg:col-span-5 space-y-4">
        
        <!-- Live File Content Viewer -->
        <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4 flex flex-col justify-between">
          <div>
            <div class="flex items-center justify-between border-b border-[#232A35] pb-2 mb-3">
              <div class="flex items-center gap-2">
                <FileCode class="w-4 h-4 text-cyan-400" />
                <h3 class="text-sm font-bold text-white font-sans">File Layer Inspector</h3>
              </div>
              <span v-if="selectedFile" class="text-[10px] font-mono text-slate-400">
                {{ selectedFile.size }}
              </span>
            </div>

            <div v-if="selectedFile" class="space-y-3 font-mono text-xs">
              <div class="p-2.5 rounded-lg bg-[#0E1217] border border-[#232A35]">
                <div class="text-[10px] text-slate-500 uppercase">Target Path</div>
                <div class="text-white font-bold mt-0.5 truncate">{{ selectedFile.path }}</div>
              </div>

              <!-- Layer Location Status -->
              <div class="p-2.5 rounded-lg border" :class="selectedFile.action === 'deleted' ? 'bg-rose-950/30 border-rose-800 text-rose-300' : selectedFile.action === 'modified' || selectedFile.action === 'added' ? 'bg-amber-950/30 border-amber-800 text-amber-300' : 'bg-cyan-950/30 border-cyan-800 text-cyan-300'">
                <div class="text-[10px] uppercase font-bold">
                  {{ selectedFile.action === 'deleted' ? 'OverlayFS Whiteout Marker' : selectedFile.action === 'modified' ? 'Copied to Upperdir (CoW)' : selectedFile.action === 'added' ? 'Created in Upperdir' : 'Direct Read from Lowerdir' }}
                </div>
                <div class="text-[11px] mt-1 font-sans text-slate-300">
                  <span v-if="selectedFile.action === 'deleted'">
                    The file was deleted in the container. Linux created a character device (0:0) with the name <code>.wh.{{ selectedFile.path.split('/').pop() }}</code> to hide the lower layer.
                  </span>
                  <span v-else-if="selectedFile.action === 'modified'">
                    The base image was never modified! The kernel duplicated this file into <code>upperdir</code> before applying changes.
                  </span>
                  <span v-else-if="selectedFile.action === 'added'">
                    This file exists solely inside the running container's writable diff layer.
                  </span>
                  <span v-else>
                    Shared from base Alpine/Ubuntu image. Reads bypass container disk entirely!
                  </span>
                </div>
              </div>

              <!-- Content Preview -->
              <div class="p-3 rounded-lg bg-[#0A0D12] border border-[#1E2633] text-[11px] font-mono text-slate-300 max-h-48 overflow-y-auto">
                <pre class="whitespace-pre-wrap">{{ selectedFile.content || '// Binary or standard system file content' }}</pre>
              </div>
            </div>

            <div v-else class="text-xs text-slate-500 text-center py-8 font-mono">
              Click any file in the stack to inspect its OverlayFS resolution.
            </div>
          </div>
        </div>

        <!-- Add New File into Container Upperdir -->
        <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-4 space-y-3">
          <div class="flex items-center gap-2 border-b border-[#232A35] pb-2">
            <Plus class="w-4 h-4 text-emerald-400" />
            <h3 class="text-sm font-bold text-white font-sans">Write to Container Layer</h3>
          </div>

          <div class="space-y-2 text-xs font-mono">
            <div>
              <label class="text-[10px] text-slate-400 uppercase">File Path:</label>
              <input 
                type="text" 
                v-model="newFilePath" 
                placeholder="/app/server.js"
                class="w-full mt-1 bg-[#0E1217] border border-[#2B3545] rounded-lg px-2.5 py-1.5 text-white focus:outline-none focus:border-[#1D63ED]"
              />
            </div>

            <div>
              <label class="text-[10px] text-slate-400 uppercase">File Content:</label>
              <textarea 
                v-model="newFileContent" 
                rows="3"
                class="w-full mt-1 bg-[#0E1217] border border-[#2B3545] rounded-lg p-2 text-white font-mono text-xs focus:outline-none focus:border-[#1D63ED]"
              ></textarea>
            </div>

            <button 
              @click="handleAddFile"
              class="w-full py-2 bg-[#1D63ED] hover:bg-[#1A57D0] text-white rounded-lg text-xs font-semibold flex items-center justify-center gap-1.5 transition-colors cursor-pointer"
            >
              <Check class="w-3.5 h-3.5" />
              Write into upperdir
            </button>
          </div>
        </div>

        <!-- Kernel Mount Command Reference -->
        <div class="p-3 rounded-xl bg-[#0A0D12] border border-[#232A35] font-mono text-[11px] text-slate-400 space-y-1">
          <div class="text-[10px] text-slate-500 uppercase font-bold">Linux Mount Syscall Behind Docker:</div>
          <code class="text-cyan-300 block break-all">
            mount -t overlay overlay -o lowerdir=/image,upperdir=/diff,workdir=/work /merged
          </code>
        </div>

      </div>
    </div>
  </div>
</template>
