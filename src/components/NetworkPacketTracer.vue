<script setup lang="ts">
import { ref } from 'vue';
import { 
  Network, 
  ArrowRight, 
  Send, 
  Play, 
  RefreshCw, 
  CheckCircle2, 
  ShieldAlert, 
  Globe, 
  Server, 
  Box, 
  Terminal,
  Radio,
  Sparkles
} from 'lucide-vue-next';

interface Hop {
  id: string;
  name: string;
  subsystem: string;
  ip: string;
  action: string;
  tableRule: string;
}

const isPacketInFlight = ref(false);
const activeHopIndex = ref<number>(-1);
const packetType = ref<'inbound' | 'container_to_container'>('inbound');
const requestPort = ref(8080);
const packetLogs = ref<string[]>([]);

const inboundHops: Hop[] = [
  {
    id: 'client',
    name: 'Browser / Curl Client',
    subsystem: 'Host Userspace',
    ip: '192.168.1.50',
    action: 'Originates HTTP GET request to http://localhost:8080',
    tableRule: 'curl http://127.0.0.1:8080/api/v1/health',
  },
  {
    id: 'prerouting',
    name: 'iptables PREROUTING Chain (DNAT)',
    subsystem: 'Linux Netfilter',
    ip: '192.168.1.100:8080',
    action: 'Matches -j DOCKER rule. Rewrites destination IP:Port from Host:8080 to Container:80',
    tableRule: '-A DOCKER -p tcp -m tcp --dport 8080 -j DNAT --to-destination 172.18.0.2:80',
  },
  {
    id: 'bridge',
    name: 'docker0 Linux Bridge',
    subsystem: 'Software Layer 2 Switch',
    ip: '172.18.0.1/16',
    action: 'Inspects ARP cache, forwards Ethernet frame toward matching veth peer interface',
    tableRule: 'brctl show docker0 / ip link set docker0 up',
  },
  {
    id: 'veth',
    name: 'Virtual Ethernet Pair (veth_c7a8)',
    subsystem: 'Kernel Virtual Pipe',
    ip: 'veth_c7a8 <--> eth0',
    action: 'Transmits packet across namespace boundary directly into container socket buffer',
    tableRule: 'veth_c7a8 peer eth0 in netns pid:[4026535480]',
  },
  {
    id: 'container',
    name: 'alpine-web (eth0)',
    subsystem: 'Isolated NET Namespace',
    ip: '172.18.0.2:80',
    action: 'TCP handshake ACK received! Nginx worker processes request and returns HTTP 200 OK',
    tableRule: 'LISTEN 0.0.0.0:80 (PID 1)',
  },
];

async function sendPacketTrace() {
  if (isPacketInFlight.value) return;
  isPacketInFlight.value = true;
  activeHopIndex.value = 0;
  packetLogs.value = [];

  for (let i = 0; i < inboundHops.length; i++) {
    activeHopIndex.value = i;
    const hop = inboundHops[i];
    packetLogs.value.unshift(`[${new Date().toLocaleTimeString()}] ${hop.name}: ${hop.action}`);
    await new Promise(r => setTimeout(r, 700));
  }

  packetLogs.value.unshift(`[${new Date().toLocaleTimeString()}] Response returned! HTTP/1.1 200 OK (Content-Length: 42, 2.1ms)`);
  setTimeout(() => {
    isPacketInFlight.value = false;
  }, 1000);
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Network class="w-5 h-5 text-indigo-400" />
          <h2 class="text-xl font-bold text-white font-sans">
            Virtual Networking & Packet Tracer (`veth` + `iptables` NAT)
          </h2>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Follow how an HTTP request on host port <code class="text-indigo-300">8080</code> is intercepted by Netfilter, translated via DNAT, and delivered across the virtual Ethernet wire into the container's private IP stack.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button 
          @click="sendPacketTrace"
          :disabled="isPacketInFlight"
          class="px-3.5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold flex items-center gap-2 transition-all cursor-pointer shadow-md disabled:opacity-50"
        >
          <Send class="w-3.5 h-3.5" :class="isPacketInFlight ? 'animate-bounce' : ''" />
          <span>{{ isPacketInFlight ? 'Tracing Packet...' : 'Send Test HTTP Request' }}</span>
        </button>
      </div>
    </div>

    <!-- MAIN PACKET FLOW VISUALIZER -->
    <div class="bg-[#141A22] border border-[#232A35] rounded-xl p-5 space-y-6">
      
      <!-- Horizontal Pipeline Stepper -->
      <div class="grid grid-cols-1 md:grid-cols-5 gap-3 relative">
        <div 
          v-for="(hop, idx) in inboundHops" 
          :key="hop.id"
          class="p-3.5 rounded-xl border transition-all relative overflow-hidden"
          :class="activeHopIndex === idx 
            ? 'bg-indigo-950/60 border-indigo-500 ring-2 ring-indigo-500/50 shadow-[0_0_15px_rgba(99,102,241,0.3)]' 
            : activeHopIndex > idx 
            ? 'bg-[#101622] border-indigo-900/60 text-slate-300' 
            : 'bg-[#0E1217] border-[#232A35] text-slate-500'"
        >
          <!-- Active Traveling Glow -->
          <div 
            v-if="activeHopIndex === idx" 
            class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-cyan-400 via-indigo-400 to-emerald-400 animate-pulse"
          ></div>

          <div class="flex items-center justify-between text-[10px] font-mono mb-1">
            <span class="font-bold" :class="activeHopIndex === idx ? 'text-indigo-400' : 'text-slate-400'">
              Hop {{ idx + 1 }}
            </span>
            <span v-if="activeHopIndex === idx" class="w-2 h-2 rounded-full bg-indigo-400 animate-ping"></span>
          </div>

          <div class="text-xs font-bold font-sans text-white truncate">
            {{ hop.name }}
          </div>

          <div class="text-[11px] font-mono text-cyan-300 mt-1 truncate">
            {{ hop.ip }}
          </div>

          <div class="text-[10px] text-slate-400 mt-2 font-sans line-clamp-2">
            {{ hop.action }}
          </div>
        </div>
      </div>

      <!-- Real-time Wire Animation & Inspection -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 pt-2">
        <!-- Left: Kernel iptables Rule Inspector -->
        <div class="lg:col-span-7 bg-[#0A0D12] border border-[#232A35] rounded-xl p-4 space-y-3 font-mono text-xs">
          <div class="flex items-center justify-between border-b border-[#1E2633] pb-2 text-[11px]">
            <span class="text-slate-400 flex items-center gap-1.5 font-bold">
              <Terminal class="w-3.5 h-3.5 text-indigo-400" />
              <span>Host Linux Kernel Netfilter / iptables Rules</span>
            </span>
            <span class="text-emerald-400 font-bold">NAT PREROUTING</span>
          </div>

          <div class="space-y-1.5 text-[11px] text-slate-300">
            <div class="text-slate-500">// 1. Incoming traffic destined for port 8080 matches Docker chain</div>
            <div class="p-2 rounded bg-[#111722] text-cyan-300 border border-[#1E2633]">
              iptables -t nat -A PREROUTING -m addrtype --dst-type LOCAL -j DOCKER
            </div>
            <div class="text-slate-500">// 2. DNAT rewrites destination IP/port to container virtual IP</div>
            <div class="p-2 rounded bg-[#111722] text-amber-300 border border-[#1E2633]">
              iptables -t nat -A DOCKER -p tcp -d 0/0 --dport 8080 -j DNAT --to-destination 172.18.0.2:80
            </div>
            <div class="text-slate-500">// 3. Bridge forwarding allowed across interfaces</div>
            <div class="p-2 rounded bg-[#111722] text-emerald-300 border border-[#1E2633]">
              iptables -A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
            </div>
          </div>
        </div>

        <!-- Right: Live Packet Trace Console -->
        <div class="lg:col-span-5 bg-[#0A0D12] border border-[#232A35] rounded-xl p-4 space-y-2 flex flex-col justify-between">
          <div>
            <div class="flex items-center justify-between border-b border-[#1E2633] pb-2 text-xs font-mono">
              <span class="text-slate-400 flex items-center gap-1.5 font-bold">
                <Radio class="w-3.5 h-3.5 text-cyan-400" />
                <span>Packet Tracer Stream</span>
              </span>
              <span class="text-[10px] text-slate-500">libpcap socket</span>
            </div>

            <div class="space-y-1 mt-2 max-h-48 overflow-y-auto text-[11px] font-mono">
              <div 
                v-for="(log, idx) in packetLogs" 
                :key="idx"
                class="p-1.5 rounded bg-[#11161F] border border-[#1C2533] text-cyan-200"
              >
                {{ log }}
              </div>
              <div v-if="packetLogs.length === 0" class="text-slate-500 text-center py-6">
                Click "Send Test HTTP Request" to trigger virtual packet traversal.
              </div>
            </div>
          </div>

          <div class="text-[10px] text-slate-500 font-mono pt-2 border-t border-[#1E2633] flex items-center justify-between">
            <span>Latency: ~0.4ms</span>
            <span>MTU: 1500 bytes</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
