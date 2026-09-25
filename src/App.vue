<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import type { Container, CreateContainerPayload } from './types/container';
import { 
  fetchContainers, 
  startContainer, 
  stopContainer, 
  pauseContainer, 
  unpauseContainer, 
  killContainer, 
  deleteContainer,
  createContainer
} from './api';

import Sidebar, { type ActiveTab } from './components/common/Sidebar.vue';
import Header from './components/common/Header.vue';
import ContainersTable from './components/ContainersTable.vue';
import ContainerCard from './components/ContainerCard.vue';
import ContainerDrawer from './components/ContainerDrawer.vue';
import ContainerSpawnerModal from './components/ContainerSpawnerModal.vue';
import DockerCliModal from './components/DockerCliModal.vue';
import KernelPrimitivesModal from './components/KernelPrimitivesModal.vue';
import Images from './components/Images.vue';
import Volumes from './components/Volumes.vue';
import KernelInternalsView from './components/KernelInternalsView.vue';
import ContainerWorkbench from './components/ContainerWorkbench.vue';
import CgroupsMonitorView from './components/CgroupsMonitorView.vue';

import { 
  Search, 
  Plus, 
  Play, 
  Square, 
  Trash2, 
  LayoutList, 
  LayoutGrid, 
  Box, 
  Sliders,
  Activity
} from 'lucide-vue-next';

const containers = ref<Container[]>([]);
const connected = ref(false);
const searchQuery = ref('');
const statusFilter = ref<'all' | 'running' | 'paused' | 'stopped' | 'oom_killed'>('all');
const activeTab = ref<ActiveTab>('workbench');
const viewMode = ref<'table' | 'cards'>('table');

// Selected containers for bulk actions
const selectedIds = ref<string[]>([]);

// Modals and Drawers
const showSpawner = ref(false);
const spawnerImage = ref('alpine:3.19');
const showCli = ref(false);
const showPrimitivesModal = ref(false);

const activeDrawerContainer = ref<Container | null>(null);
const activeDrawerTab = ref<'logs' | 'exec' | 'inspect' | 'files' | 'stats'>('logs');

let pollInterval: any = null;

onMounted(async () => {
  await loadContainers();
  pollInterval = setInterval(loadContainers, 3000);
});

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval);
});

async function loadContainers() {
  try {
    const data = await fetchContainers();
    containers.value = data;
    connected.value = true;

    // Refresh active drawer container if open
    if (activeDrawerContainer.value) {
      const match = data.find(c => c.id === activeDrawerContainer.value?.id);
      if (match) activeDrawerContainer.value = match;
    }
  } catch (err) {
    console.warn('Container engine connecting...', err);
    connected.value = false;
  }
}

const activeCount = computed(() => {
  return containers.value.filter(c => c.status === 'running').length;
});

const totalMemoryMB = computed(() => {
  const bytes = containers.value
    .filter(c => c.status === 'running')
    .reduce((sum, c) => sum + c.cgroups.memoryMaxBytes, 0);
  return Math.round(bytes / (1024 * 1024));
});

const totalCpuCores = computed(() => {
  return containers.value
    .filter(c => c.status === 'running')
    .reduce((sum, c) => sum + (c.cgroups.cpuQuotaUs / c.cgroups.cpuPeriodUs), 0);
});

const filteredContainers = computed(() => {
  return containers.value.filter(c => {
    // Status filter
    if (statusFilter.value !== 'all' && c.status !== statusFilter.value) {
      return false;
    }
    // Search query
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase();
      return (
        c.name.toLowerCase().includes(query) ||
        c.id.toLowerCase().includes(query) ||
        c.image.toLowerCase().includes(query) ||
        c.command.toLowerCase().includes(query) ||
        c.portBindings.some(p => p.toLowerCase().includes(query))
      );
    }
    return true;
  });
});

// Selection logic
function toggleSelect(id: string) {
  const idx = selectedIds.value.indexOf(id);
  if (idx > -1) {
    selectedIds.value.splice(idx, 1);
  } else {
    selectedIds.value.push(id);
  }
}

function toggleSelectAll() {
  if (selectedIds.value.length === filteredContainers.value.length) {
    selectedIds.value = [];
  } else {
    selectedIds.value = filteredContainers.value.map(c => c.id);
  }
}

// Open container drawer
function openContainerDetails(container: Container, tab: 'logs' | 'exec' | 'inspect' | 'files' | 'stats' = 'logs') {
  activeDrawerContainer.value = container;
  activeDrawerTab.value = tab;
}

// Lifecycle actions
async function handleStart(id: string) {
  try {
    await startContainer(id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

async function handleStop(id: string) {
  try {
    await stopContainer(id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

async function handlePause(id: string) {
  try {
    await pauseContainer(id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

async function handleUnpause(id: string) {
  try {
    await unpauseContainer(id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

async function handleKill(id: string) {
  try {
    await killContainer(id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

async function handleDelete(id: string) {
  try {
    await deleteContainer(id);
    selectedIds.value = selectedIds.value.filter(sId => sId !== id);
    await loadContainers();
  } catch (e) {
    console.error(e);
  }
}

// Bulk Actions
async function handleBulkStart() {
  for (const id of selectedIds.value) {
    await startContainer(id).catch(console.error);
  }
  await loadContainers();
}

async function handleBulkStop() {
  for (const id of selectedIds.value) {
    await stopContainer(id).catch(console.error);
  }
  await loadContainers();
}

async function handleBulkDelete() {
  for (const id of selectedIds.value) {
    await deleteContainer(id).catch(console.error);
  }
  selectedIds.value = [];
  await loadContainers();
}

// Spawner handler
function openSpawnerWithImage(img: string) {
  spawnerImage.value = img;
  showSpawner.value = true;
}

async function handleDeploy(payload: CreateContainerPayload) {
  try {
    await createContainer(payload);
    showSpawner.value = false;
    await loadContainers();
    activeTab.value = 'containers';
  } catch (e) {
    console.error(e);
  }
}

import { useTheme } from './composables/useTheme';
useTheme();
</script>

<template>
  <div class="min-h-screen bg-[#0b0e14] text-slate-100 flex selection:bg-blue-500/30 app-root-bg transition-colors duration-200">
    <!-- Left Sidebar -->
    <Sidebar 
      :active-tab="activeTab"
      :running-containers-count="activeCount"
      :total-containers-count="containers.length"
      :connected="connected"
      :total-memory-m-b="totalMemoryMB"
      :total-cpu-cores="totalCpuCores"
      @change-tab="activeTab = $event"
      @open-cli="showCli = true"
      @open-spawner="openSpawnerWithImage('alpine:3.19')"
    />

    <!-- Main Workspace -->
    <div class="flex-1 flex flex-col min-w-0 overflow-y-auto">
      <Header 
        :connected="connected"
        :active-count="activeCount"
        :total-count="containers.length"
        v-model:search-query="searchQuery"
        @open-spawner="openSpawnerWithImage('alpine:3.19')"
        @open-cli="showCli = true"
        @open-primitives="showPrimitivesModal = true"
        @refresh="loadContainers"
      />

      <main class="flex-1 p-5 sm:p-7 space-y-6 max-w-7xl w-full mx-auto">
        <!-- 0. WORKBENCH -->
        <ContainerWorkbench 
          v-if="activeTab === 'workbench'"
          :initial-containers="containers"
          @refresh="loadContainers"
          @inspect-container="openContainerDetails($event, 'inspect')"
          @open-metrics="activeTab = 'metrics'"
          @open-kernel="activeTab = 'kernel'"
        />

        <!-- 1. CONTAINERS VIEW -->
        <div v-else-if="activeTab === 'containers'" class="space-y-5">
          <!-- Action & Filter Bar -->
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-slate-900/50 backdrop-blur-md p-4 rounded-2xl border border-slate-800/80 shadow-lg">
            <div>
              <div class="flex items-center gap-2.5">
                <h2 class="text-xl font-bold text-white tracking-tight">Containers</h2>
                <span class="text-xs px-2.5 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700 font-mono font-medium">
                  {{ filteredContainers.length }}
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-1">
                Isolated Linux process sandboxes bounded by cgroups v2 and OverlayFS.
              </p>
            </div>

            <!-- View Mode Toggle & Bulk Actions -->
            <div class="flex items-center gap-3 flex-wrap">
              <!-- Bulk Actions Bar -->
              <div v-if="selectedIds.length > 0" class="flex items-center gap-2 bg-slate-950 p-1.5 rounded-xl border border-slate-800 animate-fadeIn">
                <span class="text-xs px-2 text-slate-400 font-mono">{{ selectedIds.length }} selected:</span>
                <button 
                  @click="handleBulkStart"
                  class="p-1.5 rounded-lg hover:bg-slate-800 text-emerald-400 transition-colors cursor-pointer"
                  title="Start selected"
                >
                  <Play class="w-3.5 h-3.5" />
                </button>
                <button 
                  @click="handleBulkStop"
                  class="p-1.5 rounded-lg hover:bg-slate-800 text-slate-300 hover:text-white transition-colors cursor-pointer"
                  title="Stop selected"
                >
                  <Square class="w-3.5 h-3.5" />
                </button>
                <button 
                  @click="handleBulkDelete"
                  class="p-1.5 rounded-lg hover:bg-slate-800 text-rose-400 hover:text-rose-300 transition-colors cursor-pointer"
                  title="Delete selected"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>

              <!-- View Switcher -->
              <div class="flex items-center rounded-xl bg-slate-950 border border-slate-800 p-1 text-slate-400 shadow-inner">
                <button
                  @click="activeTab = 'workbench'"
                  class="p-2 rounded-lg transition-all cursor-pointer hover:text-white"
                  title="Split Workbench"
                >
                  <Sliders class="w-4 h-4" />
                </button>
                <button
                  @click="viewMode = 'table'"
                  class="p-2 rounded-lg transition-all cursor-pointer"
                  :class="viewMode === 'table' ? 'bg-blue-600 text-white shadow-md shadow-blue-900/30' : 'hover:text-slate-200'"
                  title="Table view"
                >
                  <LayoutList class="w-4 h-4" />
                </button>
                <button
                  @click="viewMode = 'cards'"
                  class="p-2 rounded-lg transition-all cursor-pointer"
                  :class="viewMode === 'cards' ? 'bg-blue-600 text-white shadow-md shadow-blue-900/30' : 'hover:text-slate-200'"
                  title="Card grid view"
                >
                  <LayoutGrid class="w-4 h-4" />
                </button>
              </div>

              <!-- Run Image Button -->
              <button
                @click="openSpawnerWithImage('alpine:3.19')"
                class="px-4 py-2 rounded-xl text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 transition-all flex items-center gap-1.5 cursor-pointer shadow-lg shadow-blue-600/20"
              >
                <Plus class="w-4 h-4" />
                <span>Run Image</span>
              </button>
            </div>
          </div>

          <!-- Status Filter Tabs -->
          <div class="flex items-center gap-1.5 overflow-x-auto text-xs font-mono border-b border-slate-800/80 pb-3">
            <button 
              @click="statusFilter = 'all'"
              class="px-3.5 py-2 rounded-xl transition-all cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'all' 
                ? 'bg-slate-800 text-cyan-400 border border-slate-700 shadow-sm font-semibold' 
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/50'"
            >
              All ({{ containers.length }})
            </button>
            <button 
              @click="statusFilter = 'running'"
              class="px-3.5 py-2 rounded-xl transition-all cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'running' 
                ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/80 shadow-sm font-semibold' 
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/50'"
            >
              Running ({{ containers.filter(c => c.status === 'running').length }})
            </button>
            <button 
              @click="statusFilter = 'paused'"
              class="px-3.5 py-2 rounded-xl transition-all cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'paused' 
                ? 'bg-amber-950/80 text-amber-400 border border-amber-800/80 shadow-sm font-semibold' 
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/50'"
            >
              Paused ({{ containers.filter(c => c.status === 'paused').length }})
            </button>
            <button 
              @click="statusFilter = 'stopped'"
              class="px-3.5 py-2 rounded-xl transition-all cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'stopped' 
                ? 'bg-slate-800 text-slate-300 border border-slate-700 shadow-sm font-semibold' 
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/50'"
            >
              Stopped ({{ containers.filter(c => c.status === 'stopped').length }})
            </button>
            <button 
              @click="statusFilter = 'oom_killed'"
              class="px-3.5 py-2 rounded-xl transition-all cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'oom_killed' 
                ? 'bg-rose-950 text-rose-300 border border-rose-800 shadow-sm font-semibold' 
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/50'"
            >
              OOM Killed ({{ containers.filter(c => c.status === 'oom_killed').length }})
            </button>
          </div>

          <!-- Table View -->
          <ContainersTable 
            v-if="viewMode === 'table'"
            :containers="filteredContainers"
            :selected-ids="selectedIds"
            @toggle-select="toggleSelect"
            @select-all="toggleSelectAll"
            @open-details="openContainerDetails"
            @start="handleStart"
            @stop="handleStop"
            @pause="handlePause"
            @unpause="handleUnpause"
            @kill="handleKill"
            @delete="handleDelete"
          />

          <!-- Grid / Cards View -->
          <div v-else-if="filteredContainers.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            <ContainerCard 
              v-for="container in filteredContainers"
              :key="container.id"
              :container="container"
              @start="handleStart"
              @stop="handleStop"
              @pause="handlePause"
              @unpause="handleUnpause"
              @kill="handleKill"
              @delete="handleDelete"
              @open-details="openContainerDetails"
            />
          </div>

          <!-- Empty State -->
          <div v-else class="p-14 text-center rounded-2xl bg-slate-900/60 border border-slate-800 space-y-4 shadow-xl">
            <div class="w-14 h-14 mx-auto rounded-2xl bg-slate-950 flex items-center justify-center text-slate-500 border border-slate-800">
              <Box class="w-7 h-7" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-200 font-mono">No Containers Found</h3>
              <p class="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
                {{ searchQuery ? 'No containers match your search query.' : 'No containers match the selected status filter.' }}
              </p>
            </div>
            <button 
              @click="openSpawnerWithImage('alpine:3.19')"
              class="px-4.5 py-2.5 rounded-xl text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 transition-all inline-flex items-center gap-1.5 cursor-pointer shadow-lg shadow-blue-600/20"
            >
              <Plus class="w-4 h-4" />
              <span>Deploy New Container</span>
            </button>
          </div>
        </div>

        <!-- 2. IMAGES VIEW -->
        <Images 
          v-else-if="activeTab === 'images'"
          @run-image="openSpawnerWithImage"
        />

        <!-- 3. VOLUMES VIEW -->
        <Volumes 
          v-else-if="activeTab === 'volumes'"
        />

        <!-- 4. KERNEL INTERNALS VIEW -->
        <KernelInternalsView 
          v-else-if="activeTab === 'kernel'"
          :containers="containers"
        />

        <!-- 5. CGROUPS MONITOR VIEW -->
        <CgroupsMonitorView 
          v-else-if="activeTab === 'metrics'"
          :containers="containers"
          @refresh="loadContainers"
        />
      </main>
    </div>

    <!-- Modals & Drawers -->
    <ContainerDrawer 
      v-if="activeDrawerContainer"
      :container="activeDrawerContainer"
      :initial-tab="activeDrawerTab"
      @close="activeDrawerContainer = null"
      @refresh="loadContainers"
    />

    <ContainerSpawnerModal 
      v-if="showSpawner"
      :initial-image="spawnerImage"
      @close="showSpawner = false"
      @deploy="handleDeploy"
    />

    <DockerCliModal 
      v-if="showCli"
      @close="showCli = false"
      @refresh="loadContainers"
    />

    <KernelPrimitivesModal 
      v-if="showPrimitivesModal"
      @close="showPrimitivesModal = false"
    />
  </div>
</template>