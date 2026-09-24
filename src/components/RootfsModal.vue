<script setup lang="ts">
import { ref } from 'vue';
import type { Container } from '../types/container';
import { execCommand } from '../api';
import { 
  X, 
  FolderTree, 
  Layers, 
  ShieldCheck, 
  FileText, 
  Plus, 
  ArrowDown, 
  FilePlus, 
  ExternalLink 
} from 'lucide-vue-next';

interface Props {
  container: Container;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
}>();

const newFileName = ref('');
const isCreating = ref(false);

async function handleCreateFile() {
  if (!newFileName.value.trim() || isCreating.value) return;
  const path = newFileName.value.startsWith('/') ? newFileName.value : `/${newFileName.value}`;
  isCreating.value = true;
  try {
    await execCommand(props.container.id, `touch ${path}`);
    newFileName.value = '';
    emit('refresh');
  } catch (err) {
    console.error(err);
  } finally {
    isCreating.value = false;
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-sm">
    <div class="bg-slate-900 border border-slate-700/80 rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="p-5 border-b border-slate-800 flex items-center justify-between bg-slate-950/50">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-emerald-950/80 border border-emerald-800/80 text-emerald-400">
            <FolderTree class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white font-mono">
              RootFS Layering & pivot_root Inspector
            </h2>
            <p class="text-xs text-slate-400 font-mono">
              OverlayFS Copy-on-Write and Mount Namespace Isolation
            </p>
          </div>
        </div>
        <button 
          @click="emit('close')"
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Body -->
      <div class="p-5 overflow-y-auto space-y-5 text-xs font-mono">
        <!-- Visual OverlayFS Stack -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
          <div class="text-xs font-bold text-slate-200 flex items-center gap-2">
            <Layers class="w-4 h-4 text-cyan-400" />
            <span>OverlayFS 4-Layer Architecture</span>
          </div>

          <!-- Layer 1: Merged View (pivot_root destination) -->
          <div class="p-3 rounded-lg bg-emerald-950/40 border border-emerald-800/60 space-y-1">
            <div class="flex items-center justify-between">
              <span class="text-emerald-300 font-bold text-xs flex items-center gap-1.5">
                <ShieldCheck class="w-4 h-4" />
                <span>Merged View (Container Root /)</span>
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-emerald-900/60 text-emerald-200">
                pivot_root target
              </span>
            </div>
            <div class="text-[11px] text-slate-400 truncate">
              Path: {{ container.rootfs.mergedDir }}
            </div>
            <div class="text-[10px] text-emerald-400/80">
              The container process only sees this union. It cannot see or traverse to host directories.
            </div>
          </div>

          <!-- Arrow -->
          <div class="flex justify-center text-slate-600">
            <ArrowDown class="w-4 h-4" />
          </div>

          <!-- Layer 2: UpperDir (Read-Write container diffs) -->
          <div class="p-3 rounded-lg bg-cyan-950/40 border border-cyan-800/60 space-y-1">
            <div class="flex items-center justify-between">
              <span class="text-cyan-300 font-bold text-xs flex items-center gap-1.5">
                <FilePlus class="w-4 h-4" />
                <span>UpperDir (Read-Write COW Layer)</span>
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-cyan-900/60 text-cyan-200">
                {{ container.rootfs.modifiedFiles.length }} diffs
              </span>
            </div>
            <div class="text-[11px] text-slate-400 truncate">
              Path: {{ container.rootfs.upperDir }}
            </div>
            <div class="text-[10px] text-cyan-400/80">
              All file modifications, writes, and whiteout deletions persist here without touching the base image.
            </div>
          </div>

          <!-- Arrow -->
          <div class="flex justify-center text-slate-600">
            <ArrowDown class="w-4 h-4" />
          </div>

          <!-- Layer 3: LowerDir (Read-Only base image) -->
          <div class="p-3 rounded-lg bg-slate-900 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between">
              <span class="text-slate-300 font-bold text-xs flex items-center gap-1.5">
                <FolderTree class="w-4 h-4 text-slate-400" />
                <span>LowerDir (Read-Only Base Image)</span>
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-300">
                Immutable
              </span>
            </div>
            <div class="text-[11px] text-slate-400 truncate">
              Path: {{ container.rootfs.lowerDir }}
            </div>
            <div class="text-[10px] text-slate-500">
              Shared read-only image layers (Alpine/Busybox root filesystem tarball).
            </div>
          </div>
        </div>

        <!-- Interactive COW Diff Tester -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold text-slate-200">
              Copy-on-Write (UpperDir) Modifications
            </span>
            <span class="text-[10px] text-slate-500">Tracked in realtime</span>
          </div>

          <!-- Add test file form -->
          <form @submit.prevent="handleCreateFile" class="flex items-center gap-2">
            <input 
              v-model="newFileName"
              type="text" 
              placeholder="/app/config.json or /tmp/test.log"
              class="flex-1 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-slate-100 text-xs focus:outline-none focus:border-cyan-500"
              :disabled="container.status !== 'running'"
            />
            <button 
              type="submit"
              :disabled="!newFileName.trim() || isCreating || container.status !== 'running'"
              class="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 disabled:opacity-30 text-white font-medium text-xs flex items-center gap-1 transition-colors cursor-pointer"
            >
              <Plus class="w-3.5 h-3.5" />
              <span>Touch File</span>
            </button>
          </form>

          <!-- Diffs Table -->
          <div class="rounded-lg border border-slate-800 overflow-hidden">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 border-b border-slate-800">
                <tr>
                  <th class="px-3 py-2">File Path</th>
                  <th class="px-3 py-2">COW Action</th>
                  <th class="px-3 py-2">Size</th>
                  <th class="px-3 py-2">Time</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                <tr v-for="(file, idx) in container.rootfs.modifiedFiles" :key="idx" class="hover:bg-slate-900/40">
                  <td class="px-3 py-1.5 text-cyan-300 font-semibold">{{ file.path }}</td>
                  <td class="px-3 py-1.5">
                    <span class="px-1.5 py-0.5 rounded text-[10px] bg-cyan-950 text-cyan-400 border border-cyan-800">
                      {{ file.action }}
                    </span>
                  </td>
                  <td class="px-3 py-1.5 text-slate-400">{{ file.sizeBytes }} B</td>
                  <td class="px-3 py-1.5 text-slate-500">{{ file.timestamp }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Technical Deep Dive: pivot_root vs chroot -->
        <div class="p-3.5 rounded-xl bg-slate-950/40 border border-slate-800 space-y-1.5 text-[11px] text-slate-400">
          <div class="text-slate-200 font-bold flex items-center gap-1.5">
            <ShieldCheck class="w-3.5 h-3.5 text-cyan-400" />
            <span>Why pivot_root instead of chroot?</span>
          </div>
          <p>
            Standard <code class="text-cyan-300">chroot</code> only changes the current process's view of <code class="text-cyan-300">/</code>, but doesn't change the underlying mount table. A root process can break out using <code class="text-cyan-300">fchdir()</code>.
          </p>
          <p>
            In contrast, <code class="text-emerald-400">pivot_root</code> swaps the entire filesystem mount namespace, moves the old root to <code class="text-slate-300">.oldroot</code>, and detaches/unmounts it with <code class="text-emerald-400">syscall.MNT_DETACH</code>, eliminating any path back to the host filesystem.
          </p>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-800 bg-slate-950/60 flex justify-end">
        <button 
          @click="emit('close')"
          class="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 bg-slate-800 hover:bg-slate-700 transition-colors cursor-pointer"
        >
          Close Inspector
        </button>
      </div>
    </div>
  </div>
</template>
