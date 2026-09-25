<script setup lang="ts">
import { ref } from 'vue';
import { 
  Boxes, 
  Radio,
  RefreshCw,
  AlertTriangle,
  Server,
  Workflow
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
  <div class="space-y-6 font-sans select-none">
    
    <!-- Modern Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-5 rounded-2xl bg-[#161B22] border border-[#21262d] shadow-xl">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-indigo-500/10 border border-indigo-500/30 text-indigo-400">
            <Boxes class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white flex items-center gap-2">
              <span>Multi-Container Compose Topology & Cascading Healthchecks</span>
              <span class="text-[10px] px-2 py-0.5 rounded font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
                Docker Engine v27
              </span>
            </h2>
            <p class="text-xs text-slate-400">
              Simulate microservice relationships in a custom bridge network <code class="text-indigo-300">app_net</code>. Trigger container crashes to watch healthcheck cascades and automated restart policies.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2.5">
        <button 
          @click="resetAll"
          class="px-3.5 py-2 rounded-xl bg-[#0D1117] hover:bg-[#21262d] text-slate-300 hover:text-white border border-[#2D3848] text-xs font-mono font-bold flex items-center gap-2 transition-all cursor-pointer shadow-sm"
        >
          <RefreshCw class="w-3.5 h-3.5 text-indigo-400" />
          <span>Reset Stack</span>
        </button>
      </div>
    </div>

    <!-- Active Event Banner -->
    <div class="p-3.5 rounded-xl bg-[#141A22] border border-[#21262d] text-xs font-mono text-cyan-300 flex items-center justify-between shadow-inner">
      <div class="flex items-center gap-2.5 truncate">
        <Radio class="w-4 h-4 text-cyan-400 shrink-0 animate-pulse" />
        <span class="truncate">{{ activeActionLog }}</span>
      </div>
      <span class="text-[10px] text-slate-500 uppercase shrink-0 ml-3 bg-[#0D1117] px-2 py-1 rounded border border-[#21262d]">bridge: 172.20.0.0/16</span>
    </div>

    <!-- TOPOLOGY NODE GRAPH WITH CONNECTING PIPELINE -->
    <div class="bg-[#141A22] border border-[#21262d] rounded-2xl p-6 relative shadow-xl">
      
      <!-- Horizontal Connecting Line (Lifecycle / Dependency Flow) -->
      <div class="hidden lg:block absolute top-[68px] left-12 right-12 h-0.5 bg-[#21262d] z-0"></div>

      <!-- Service Nodes Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 relative z-10">
        <div 
          v-for="(svc, idx) in services" 
          :key="svc.name"
          class="border rounded-xl p-4 flex flex-col justify-between transition-all shadow-lg relative overflow-hidden"
          :class="svc.status === 'healthy' 
            ? 'bg-[#0E1217] border-[#21262d] hover:border-indigo-500/40 text-white' 
            : svc.status === 'unhealthy'
            ? 'bg-amber-950/20 border-amber-500/60 shadow-[0_0_20px_rgba(245,158,11,0.15)] text-white'
            : svc.status === 'crashed'
            ? 'bg-rose-950/30 border-rose-500/80 shadow-[0_0_20px_rgba(244,63,94,0.2)] text-white'
            : 'bg-cyan-950/20 border-cyan-500/60 animate-pulse text-white'"
        >
          <!-- Active Status Glow Bar -->
          <div 
            class="absolute top-0 left-0 right-0 h-1"
            :class="svc.status === 'healthy' ? 'bg-indigo-500/60' : svc.status === 'unhealthy' ? 'bg-amber-400' : svc.status === 'crashed' ? 'bg-rose-500' : 'bg-cyan-400'"
          ></div>

          <div>
            <!-- Header: Name & Status Badge -->
            <div class="flex items-center justify-between mb-2.5 pt-1">
              <div class="flex items-center gap-2">
                <span 
                  class="w-2.5 h-2.5 rounded-full"
                  :class="svc.status === 'healthy' ? 'bg-emerald-400' : svc.status === 'unhealthy' ? 'bg-amber-400 animate-pulse' : svc.status === 'crashed' ? 'bg-rose-500' : 'bg-cyan-400 animate-ping'"
                ></span>
                <span class="font-bold text-white font-mono text-xs truncate max-w-[110px]">{{ svc.name }}</span>
              </div>
              <span 
                class="text-[9px] font-mono uppercase px-2 py-0.5 rounded font-bold tracking-wider"
                :class="svc.status === 'healthy' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : svc.status === 'unhealthy' ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' : 'bg-rose-500/20 text-rose-400 border border-rose-500/30'"
              >
                {{ svc.status }}
              </span>
            </div>

            <div class="text-[11px] text-slate-400 font-sans mb-3 h-8 leading-tight">
              {{ svc.role }}
            </div>

            <!-- Metadata Box -->
            <div class="space-y-1.5 text-[11px] font-mono text-slate-300 bg-[#0D1117] p-3 rounded-xl border border-[#21262d] mb-4 shadow-inner">
              <div class="flex justify-between items-center">
                <span class="text-slate-500 text-[10px]">Container IP:</span>
                <span class="text-cyan-300 font-bold">{{ svc.ip }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-slate-500 text-[10px]">Image:</span>
                <span class="text-slate-300 truncate ml-2 max-w-[120px]">{{ svc.image }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-slate-500 text-[10px]">Restarts:</span>
                <span class="text-amber-400 font-bold">{{ svc.restarts }}</span>
              </div>
              <div v-if="svc.dependsOn.length > 0" class="pt-2 border-t border-[#21262d]">
                <span class="text-slate-500 text-[10px] block mb-1">depends_on:</span>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="dep in svc.dependsOn" 
                    :key="dep"
                    class="text-[9px] px-1.5 py-0.5 rounded bg-indigo-950/80 text-indigo-300 border border-indigo-800/60 font-mono"
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
            class="w-full py-2 rounded-xl text-xs font-mono font-bold transition-all cursor-pointer disabled:opacity-40 flex items-center justify-center gap-1.5 shadow-md"
            :class="svc.status === 'healthy' 
              ? 'bg-rose-950/60 hover:bg-rose-900 text-rose-300 border border-rose-800/80' 
              : 'bg-[#161B22] text-slate-400 border border-[#21262d]'"
          >
            <AlertTriangle class="w-3.5 h-3.5" v-if="svc.status === 'healthy'" />
            <span>{{ svc.status === 'crashed' ? 'SIGKILL (137)' : svc.status === 'restarting' ? 'Restarting...' : 'Crash Container' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>