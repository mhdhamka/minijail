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
import ImagesView from './components/ImagesView.vue';
import VolumesView from './components/VolumesView.vue';
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
  Sliders
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
  <div class="min-h-screen bg-[#0F1318] text-slate-100 flex selection:bg-[#1D63ED]/30 app-root-bg transition-colors duration-200">
    <!-- Docker Desktop Left Sidebar -->
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

    <!-- Right Main Work Area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-y-auto">
      <!-- Top Header Navigation -->
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

      <!-- Main Content Views -->
      <main class="flex-1 p-4 sm:p-6 space-y-5 max-w-7xl w-full mx-auto">
        <!-- 0. WORKBENCH (Simulated Terminal Left + Container Operations Right) -->
        <ContainerWorkbench 
          v-if="activeTab === 'workbench'"
          :initial-containers="containers"
          @refresh="loadContainers"
          @inspect-container="openContainerDetails($event, 'inspect')"
          @open-metrics="activeTab = 'metrics'"
          @open-kernel="activeTab = 'kernel'"
        />

        <!-- 1. CONTAINERS VIEW -->
        <div v-else-if="activeTab === 'containers'" class="space-y-4">
          <!-- Action & Filter Bar (Authentic Docker Desktop styling) -->
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <!-- Left: Title & Count -->
            <div>
              <h2 class="text-xl font-bold text-white flex items-center gap-2 font-sans">
                <span>Containers</span>
                <span class="text-xs px-2 py-0.5 rounded-full bg-[#1E2633] text-slate-400 border border-[#2B3545] font-mono">
                  {{ filteredContainers.length }}
                </span>
              </h2>
              <p class="text-xs text-slate-400 mt-0.5">
                Isolated Linux process sandboxes bounded by cgroups v2 and OverlayFS.
              </p>
            </div>

            <!-- Right: View Mode Toggle & Bulk Actions -->
            <div class="flex items-center gap-2">
              <!-- Bulk Actions if selected -->
              <div v-if="selectedIds.length > 0" class="flex items-center gap-1.5 bg-[#161B22] p-1 rounded-lg border border-[#232A35]">
                <span class="text-xs px-2 text-slate-400 font-mono">{{ selectedIds.length }} selected:</span>
                <button 
                  @click="handleBulkStart"
                  class="p-1.5 rounded hover:bg-[#202836] text-emerald-400 transition-colors cursor-pointer"
                  title="Start selected"
                >
                  <Play class="w-3.5 h-3.5" />
                </button>
                <button 
                  @click="handleBulkStop"
                  class="p-1.5 rounded hover:bg-[#202736] text-slate-300 hover:text-white transition-colors cursor-pointer"
                  title="Stop selected"
                >
                  <Square class="w-3.5 h-3.5" />
                </button>
                <button 
                  @click="handleBulkDelete"
                  class="p-1.5 rounded hover:bg-[#202836] text-rose-400 hover:text-rose-300 transition-colors cursor-pointer"
                  title="Delete selected"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>

              <!-- View Switcher (Table vs Cards vs Workbench) -->
              <div class="flex items-center rounded-lg bg-[#161B22] border border-[#232A35] p-1 text-slate-400">
                <button
                  @click="activeTab = 'workbench'"
                  class="p-1.5 rounded transition-colors cursor-pointer text-slate-400 hover:text-white"
                  title="Split Workbench (Simulated Terminal + Operations)"
                >
                  <Sliders class="w-4 h-4" />
                </button>
                <button
                  @click="viewMode = 'table'"
                  class="p-1.5 rounded transition-colors cursor-pointer"
                  :class="viewMode === 'table' ? 'bg-[#1D63ED] text-white shadow-sm' : 'hover:text-slate-200'"
                  title="Table view (Docker Desktop default)"
                >
                  <LayoutList class="w-4 h-4" />
                </button>
                <button
                  @click="viewMode = 'cards'"
                  class="p-1.5 rounded transition-colors cursor-pointer"
                  :class="viewMode === 'cards' ? 'bg-[#1D63ED] text-white shadow-sm' : 'hover:text-slate-200'"
                  title="Card grid view"
                >
                  <LayoutGrid class="w-4 h-4" />
                </button>
              </div>

              <!-- + Run Image -->
              <button
                @click="openSpawnerWithImage('alpine:3.19')"
                class="px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-[#1D63ED] hover:bg-[#1A57D0] transition-colors flex items-center gap-1.5 cursor-pointer shadow-sm"
              >
                <Plus class="w-4 h-4" />
                <span>Run Image</span>
              </button>
            </div>
          </div>

          <!-- Filter Tabs -->
          <div class="flex items-center gap-1 overflow-x-auto text-xs font-mono border-b border-[#232A35] pb-2">
            <button 
              @click="statusFilter = 'all'"
              class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'all' 
                ? 'bg-[#1E2633] text-[#0db7ed] border border-[#2D3848] font-semibold' 
                : 'text-slate-400 hover:text-slate-200'"
            >
              All ({{ containers.length }})
            </button>
            <button 
              @click="statusFilter = 'running'"
              class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'running' 
                ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800 font-semibold' 
                : 'text-slate-400 hover:text-slate-200'"
            >
              Running ({{ containers.filter(c => c.status === 'running').length }})
            </button>
            <button 
              @click="statusFilter = 'paused'"
              class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'paused' 
                ? 'bg-amber-950/80 text-amber-400 border border-amber-800 font-semibold' 
                : 'text-slate-400 hover:text-slate-200'"
            >
              Paused ({{ containers.filter(c => c.status === 'paused').length }})
            </button>
            <button 
              @click="statusFilter = 'stopped'"
              class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'stopped' 
                ? 'bg-[#1E2633] text-slate-300 border border-[#2D3848] font-semibold' 
                : 'text-slate-400 hover:text-slate-200'"
            >
              Stopped ({{ containers.filter(c => c.status === 'stopped').length }})
            </button>
            <button 
              @click="statusFilter = 'oom_killed'"
              class="px-3 py-1.5 rounded-lg transition-colors cursor-pointer shrink-0 font-medium"
              :class="statusFilter === 'oom_killed' 
                ? 'bg-rose-950 text-rose-300 border border-rose-800 font-semibold' 
                : 'text-slate-400 hover:text-slate-200'"
            >
              OOM Killed ({{ containers.filter(c => c.status === 'oom_killed').length }})
            </button>
          </div>

          <!-- Table View (Classic Docker Desktop layout) -->
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
          <div v-else-if="filteredContainers.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
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
          <div v-else class="p-12 text-center rounded-xl bg-[#161B22] border border-[#232A35] space-y-4">
            <div class="w-12 h-12 mx-auto rounded-xl bg-[#1A222D] flex items-center justify-center text-slate-500">
              <Box class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-200 font-mono">No Containers Found</h3>
              <p class="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
                {{ searchQuery ? 'No containers match your search query.' : 'No containers match the selected status filter.' }}
              </p>
            </div>
            <button 
              @click="openSpawnerWithImage('alpine:3.19')"
              class="px-4 py-2 rounded-lg text-xs font-semibold text-white bg-[#1D63ED] hover:bg-[#1A57D0] transition-colors inline-flex items-center gap-1.5 cursor-pointer"
            >
              <Plus class="w-4 h-4" />
              <span>Deploy New Container</span>
            </button>
          </div>
        </div>

        <!-- 2. IMAGES VIEW -->
        <ImagesView 
          v-else-if="activeTab === 'images'"
          @run-image="openSpawnerWithImage"
        />

        <!-- 3. VOLUMES VIEW -->
        <VolumesView 
          v-else-if="activeTab === 'volumes'"
        />

        <!-- 4. UNDER THE HOOD (KERNEL INTERNALS) VIEW -->
        <KernelInternalsView 
          v-else-if="activeTab === 'kernel'"
          :containers="containers"
        />

        <!-- 5. REAL-TIME RESOURCE MONITOR (CGROUPS THROTTLING & MEMORY CHARTS) -->
        <CgroupsMonitorView 
          v-else-if="activeTab === 'metrics'"
          :containers="containers"
          @refresh="loadContainers"
        />
      </main>
    </div>

    <!-- Modals & Drawers -->

    <!-- Container Details Drawer (Logs, Exec, Inspect, Files, Stats) -->
    <ContainerDrawer 
      v-if="activeDrawerContainer"
      :container="activeDrawerContainer"
      :initial-tab="activeDrawerTab"
      @close="activeDrawerContainer = null"
      @refresh="loadContainers"
    />

    <!-- Container Spawner Modal -->
    <ContainerSpawnerModal 
      v-if="showSpawner"
      :initial-image="spawnerImage"
      @close="showSpawner = false"
      @deploy="handleDeploy"
    />

    <!-- Docker CLI Terminal Modal -->
    <DockerCliModal 
      v-if="showCli"
      @close="showCli = false"
      @refresh="loadContainers"
    />

    <!-- Kernel Primitives Modal Quick View -->
    <KernelPrimitivesModal 
      v-if="showPrimitivesModal"
      @close="showPrimitivesModal = false"
    />
  </div>
</template>
