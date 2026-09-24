<script setup lang="ts">
import { ref } from 'vue';
import { 
  Boxes, 
  Network, 
  HardDrive, 
  Play, 
  Square, 
  RotateCcw, 
  CheckCircle2, 
  AlertTriangle, 
  ArrowRight, 
  Activity, 
  FolderTree,
  Radio,
  Server
} from 'lucide-vue-next';

interface ComposeService {
  name: string;
  image: string;
  role: string;
  color: string;
  status: 'healthy' | 'unhealthy' | 'crashed' | 'restarting';
  ip: string;
  ports: string[];
  dependsOn: string[];
  restartPolicy: string;
  restarts: number;
}

const services = ref<ComposeService[]>([
  {
    name: 'frontend-nginx',
    image: 'nginx:alpine',
    role: 'Reverse Proxy & Client UI',
    color: '#06b6d4',
    status: 'healthy',
    ip: '172.20.0.2',
    ports: ['80:80', '443:443'],
    dependsOn: ['api-backend'],
    restartPolicy: 'unless-stopped',
    restarts: 0,
  },
  {
    name: 'api-backend',
    image: 'node:20-alpine',
    role: 'REST API & Business Logic',
    color: '#3b82f6',
    status: 'healthy',
    ip: '172.20.0.3',
    ports: ['3000:3000'],
    dependsOn: ['redis-cache', 'postgres-db'],
    restartPolicy: 'always',
    restarts: 0,
  },
  {
    name: 'redis-cache',
    image: 'redis:7-alpine',
    role: 'Session & Job Queue In-Memory Store',
    color: '#ef4444',
    status: 'healthy',
    ip: '172.20.0.4',
    ports: ['6379:6379'],
    dependsOn: [],
    restartPolicy: 'unless-stopped',
    restarts: 0,
  },
  {
    name: 'postgres-db',
    image: 'postgres:16-alpine',
    role: 'Persistent Relational Database',
    color: '#8b5cf6',
    status: 'healthy',
    ip: '172.20.0.5',
    ports: ['5432:5432'],
    dependsOn: [],
    restartPolicy: 'always',
    restarts: 0,
  },
]);

const activeActionLog = ref<string>('Docker Compose stack "microservices" running on user-defined bridge network `app_net` (172.20.0.0/16)');

function killService(name: string) {
  const svc = services.value.find(s => s.name === name);
  if (!svc) return;

  svc.status = 'crashed';
  activeActionLog.value = `[Crash Event] Sent SIGKILL to ${name}. Container process exited with code 137.`;

  // Cascading dependency impact
  for (const other of services.value) {
    if (other.dependsOn.includes(name)) {
      other.status = 'unhealthy';
    }
  }

  // Automatic restart policy simulation
  setTimeout(() => {
    svc.status = 'restarting';
    activeActionLog.value = `[Daemon Restart Policy] Policy "${svc.restartPolicy}" restarting container ${name}...`;
    setTimeout(() => {
      svc.status = 'healthy';
      svc.restarts++;
      activeActionLog.value = `[Healthcheck Passed] Container ${name} healthy! Restored connectivity.`;
      // Recover dependents
      for (const other of services.value) {
        if (other.status === 'unhealthy') {
          other.status = 'healthy';
        }
      }
    }, 1500);
  }, 1800);
}

function resetAll() {
  for (const s of services.value) {
    s.status = 'healthy';
    s.restarts = 0;
  }
  activeActionLog.value = 'Docker Compose stack reset to initial healthy state.';
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Boxes class="w-5 h-5 text-indigo-400" />
          <h2 class="text-xl font-bold text-white font-sans">
            Multi-Container Compose Topology & Cascading Healthchecks
          </h2>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Simulate microservice relationships in a custom Docker bridge network. Kill a dependency (like Redis or Postgres) to observe healthcheck cascades and automated restart policies (`restart: always`).
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button 
          @click="resetAll"
          class="px-3 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 hover:text-white border border-[#2D3848] text-xs font-mono transition-colors cursor-pointer"
        >
          Reset All Services
        </button>
      </div>
    </div>

    <!-- Active Event Banner -->
    <div class="p-3 rounded-lg bg-[#141A22] border border-[#232A35] text-xs font-mono text-cyan-300 flex items-center justify-between">
      <div class="flex items-center gap-2 truncate">
        <Radio class="w-4 h-4 text-cyan-400 shrink-0 animate-pulse" />
        <span class="truncate">{{ activeActionLog }}</span>
      </div>
      <span class="text-[10px] text-slate-500 uppercase shrink-0 ml-2">bridge: app_net (172.20.0.0/16)</span>
    </div>

    <!-- TOPOLOGY NODE GRAPH -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div 
        v-for="svc in services" 
        :key="svc.name"
        class="bg-[#141A22] border rounded-xl p-4 flex flex-col justify-between transition-all"
        :class="svc.status === 'healthy' 
          ? 'border-[#232A35] hover:border-indigo-500/50' 
          : svc.status === 'unhealthy'
          ? 'border-amber-500/60 bg-amber-950/20 shadow-[0_0_15px_rgba(245,158,11,0.15)]'
          : svc.status === 'crashed'
          ? 'border-rose-500 bg-rose-950/30 shadow-[0_0_15px_rgba(244,63,94,0.2)]'
          : 'border-cyan-500 bg-cyan-950/20 animate-pulse'"
      >
        <div>
          <!-- Header: Name & Status -->
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2">
              <span 
                class="w-2.5 h-2.5 rounded-full"
                :class="svc.status === 'healthy' ? 'bg-emerald-400' : svc.status === 'unhealthy' ? 'bg-amber-400 animate-pulse' : svc.status === 'crashed' ? 'bg-rose-500' : 'bg-cyan-400 animate-ping'"
              ></span>
              <span class="font-bold text-white font-mono text-xs">{{ svc.name }}</span>
            </div>
            <span 
              class="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded font-bold"
              :class="svc.status === 'healthy' ? 'bg-emerald-500/10 text-emerald-400' : svc.status === 'unhealthy' ? 'bg-amber-500/10 text-amber-400' : 'bg-rose-500/20 text-rose-400'"
            >
              {{ svc.status }}
            </span>
          </div>

          <div class="text-[11px] text-slate-400 font-sans mb-3">
            {{ svc.role }}
          </div>

          <!-- Metadata -->
          <div class="space-y-1 text-[11px] font-mono text-slate-300 bg-[#0E1217] p-2.5 rounded-lg border border-[#1E2530] mb-3">
            <div class="flex justify-between">
              <span class="text-slate-500">IP:</span>
              <span class="text-cyan-300 font-bold">{{ svc.ip }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500">Image:</span>
              <span class="text-slate-300 truncate ml-2">{{ svc.image }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500">Restarts:</span>
              <span class="text-amber-400 font-bold">{{ svc.restarts }}</span>
            </div>
            <div v-if="svc.dependsOn.length > 0" class="pt-1 border-t border-[#1C2330]">
              <span class="text-slate-500 text-[10px] block mb-0.5">depends_on:</span>
              <div class="flex flex-wrap gap-1">
                <span 
                  v-for="dep in svc.dependsOn" 
                  :key="dep"
                  class="text-[9px] px-1 py-0.2 rounded bg-indigo-950 text-indigo-300 border border-indigo-800"
                >
                  {{ dep }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Failure Simulator Trigger Button -->
        <button 
          @click="killService(svc.name)"
          :disabled="svc.status === 'crashed' || svc.status === 'restarting'"
          class="w-full py-1.5 rounded-lg text-xs font-mono font-semibold transition-colors cursor-pointer disabled:opacity-40"
          :class="svc.status === 'healthy' 
            ? 'bg-rose-950/60 hover:bg-rose-900 text-rose-300 border border-rose-800/80' 
            : 'bg-slate-800 text-slate-400'"
        >
          {{ svc.status === 'crashed' ? 'SIGKILL sent (137)' : svc.status === 'restarting' ? 'Auto-Restarting...' : 'Crash Service (SIGKILL)' }}
        </button>
      </div>
    </div>
  </div>
</template>
