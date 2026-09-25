<script setup lang="ts">
import { ref } from 'vue';
import { 
  Network, 
  Send, 
  Terminal,
  Radio,
  Activity
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
const packetLogs = ref<string[]>([]);
const packetProtocol = ref<'tcp' | 'udp'>('tcp');
const targetPort = ref(8080);

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
    packetLogs.value.unshift(`[${new Date().toLocaleTimeString()}] [${hop.subsystem}] ${hop.name}: Action verified.`);
    await new Promise(r => setTimeout(r, 650));
  }

  packetLogs.value.unshift(`[${new Date().toLocaleTimeString()}] [OK] Response returned! HTTP/1.1 200 OK (Content-Length: 42, 1.8ms)`);
  setTimeout(() => {
    isPacketInFlight.value = false;
  }, 800);
}
</script>

<template>
  <div class="space-y-6 font-sans select-none">
    
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 rounded-2xl bg-[#161B22] border border-[#21262d] shadow-xl">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-indigo-500/10 border border-indigo-500/30 text-indigo-400">
            <Network class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white font-sans flex items-center gap-2">
              <span>Virtual Networking & Packet Tracer (`veth` + `iptables` NAT)</span>
              <span class="text-[10px] px-2 py-0.5 rounded font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
                Layer 2/3 Bridge
              </span>
            </h2>
            <p class="text-xs text-slate-400">
              Follow how an HTTP request on host port <code class="text-indigo-300">{{ targetPort }}</code> is intercepted by Netfilter, translated via DNAT, and delivered across the virtual Ethernet wire.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2.5">
        <div class="flex items-center bg-[#0D1117] rounded-xl p-1 border border-[#2D3848] text-xs font-mono">
          <button 
            @click="packetProtocol = 'tcp'"
            class="px-2.5 py-1.5 rounded-lg transition-all cursor-pointer"
            :class="packetProtocol === 'tcp' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
          >
            TCP
          </button>
          <button 
            @click="packetProtocol = 'udp'"
            class="px-2.5 py-1.5 rounded-lg transition-all cursor-pointer"
            :class="packetProtocol === 'udp' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
          >
            UDP
          </button>
        </div>

        <button 
          @click="sendPacketTrace"
          :disabled="isPacketInFlight"
          class="px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-mono font-bold flex items-center gap-2 transition-all cursor-pointer shadow-lg disabled:opacity-50 ring-1 ring-indigo-400/30"
        >
          <Send class="w-3.5 h-3.5" :class="isPacketInFlight ? 'animate-bounce' : ''" />
          <span>{{ isPacketInFlight ? 'Tracing Packet...' : 'Send Packet' }}</span>
        </button>
      </div>
    </div>

    <!-- MAIN PACKET FLOW VISUALIZER WITH HORIZONTAL CONNECTOR LINE -->
    <div class="bg-[#141A22] border border-[#21262d] rounded-2xl p-6 space-y-6 shadow-xl relative">
      
      <!-- Continuous Absolute Background Line connecting all steps -->
      <div class="hidden lg:block absolute top-[44px] left-12 right-12 h-0.5 bg-[#21262d] z-0"></div>

      <!-- Pipeline Container with Flex and Connected Cards -->
      <div class="grid grid-cols-1 lg:grid-cols-5 gap-4 relative z-10">
        <template v-for="(hop, idx) in inboundHops" :key="hop.id">
          
          <!-- Hop Card -->
          <div 
            class="p-4 rounded-xl border transition-all relative overflow-hidden flex flex-col justify-between shadow-lg"
            :class="activeHopIndex === idx 
              ? 'bg-indigo-950/60 border-indigo-500 ring-2 ring-indigo-500/40 shadow-[0_0_20px_rgba(99,102,241,0.2)] text-white' 
              : activeHopIndex > idx 
              ? 'bg-[#101622] border-indigo-900/40 text-slate-300' 
              : 'bg-[#0E1217] border-[#21262d] text-slate-400'"
          >
            <!-- Active Traveling Glow -->
            <div 
              v-if="activeHopIndex === idx" 
              class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-cyan-400 via-indigo-400 to-emerald-400 animate-pulse"
            ></div>

            <div>
              <div class="flex items-center justify-between text-[10px] font-mono mb-2">
                <span class="font-bold px-2 py-0.5 rounded bg-[#0D1117] border border-[#21262d]" :class="activeHopIndex === idx ? 'text-indigo-300 border-indigo-500/40' : 'text-slate-400'">
                  HOP 0{{ idx + 1 }}
                </span>
                <span class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-mono border" :class="activeHopIndex === idx ? 'bg-indigo-600 text-white border-indigo-400' : 'bg-[#161B22] text-slate-400 border-[#2D3848]'">
                  {{ idx + 1 }}
                </span>
              </div>

              <div class="text-xs font-bold font-sans text-white mt-2">
                {{ hop.name }}
              </div>

              <div class="text-[11px] font-mono text-cyan-300 mt-1">
                {{ hop.ip }}
              </div>
            </div>

            <div class="text-[10px] text-slate-400 mt-3 font-sans leading-relaxed pt-2 border-t border-[#21262d]/60">
              {{ hop.action }}
            </div>
          </div>

        </template>
      </div>

      <!-- Real-time Wire Animation & Inspection Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-5 pt-4">
        
        <!-- Left: Kernel iptables Rule Inspector -->
        <div class="lg:col-span-7 bg-[#0D1117] border border-[#21262d] rounded-xl p-4 space-y-3 font-mono text-xs shadow-inner">
          <div class="flex items-center justify-between border-b border-[#21262d] pb-2.5 text-[11px]">
            <span class="text-slate-300 flex items-center gap-1.5 font-bold font-sans">
              <Terminal class="w-3.5 h-3.5 text-indigo-400" />
              <span>Host Linux Kernel Netfilter / iptables Rules</span>
            </span>
            <span class="text-emerald-400 font-mono text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/30">NAT PREROUTING</span>
          </div>

          <div class="space-y-2 text-[11px] text-slate-300">
            <div>
              <span class="text-slate-500 text-[10px] block mb-0.5">// 1. Incoming traffic destined for port {{ targetPort }} matches Docker chain</span>
              <div class="p-2.5 rounded-lg bg-[#141A22] text-cyan-300 border border-[#21262d] font-mono shadow-sm">
                iptables -t nat -A PREROUTING -m addrtype --dst-type LOCAL -j DOCKER
              </div>
            </div>
            <div>
              <span class="text-slate-500 text-[10px] block mb-0.5">// 2. DNAT rewrites destination IP/port to container virtual IP</span>
              <div class="p-2.5 rounded-lg bg-[#141A22] text-amber-300 border border-[#21262d] font-mono shadow-sm">
                iptables -t nat -A DOCKER -p tcp -d 0/0 --dport {{ targetPort }} -j DNAT --to-destination 172.18.0.2:80
              </div>
            </div>
            <div>
              <span class="text-slate-500 text-[10px] block mb-0.5">// 3. Bridge forwarding allowed across veth interface boundaries</span>
              <div class="p-2.5 rounded-lg bg-[#141A22] text-emerald-300 border border-[#21262d] font-mono shadow-sm">
                iptables -A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Live Packet Trace Console -->
        <div class="lg:col-span-5 bg-[#0D1117] border border-[#21262d] rounded-xl p-4 space-y-3 flex flex-col justify-between shadow-inner">
          <div>
            <div class="flex items-center justify-between border-b border-[#21262d] pb-2.5 text-xs font-mono">
              <span class="text-slate-300 flex items-center gap-1.5 font-bold font-sans">
                <Radio class="w-3.5 h-3.5 text-cyan-400" />
                <span>Packet Tracer Stream</span>
              </span>
              <span class="text-[10px] text-cyan-400 bg-cyan-500/10 border border-cyan-500/30 px-2 py-0.5 rounded">libpcap socket</span>
            </div>

            <div class="space-y-1.5 mt-3 max-h-44 overflow-y-auto text-[11px] font-mono">
              <div 
                v-for="(log, idx) in packetLogs" 
                :key="idx"
                class="p-2 rounded-lg bg-[#141A22] border border-[#21262d] text-cyan-200 shadow-sm"
              >
                {{ log }}
              </div>
              <div v-if="packetLogs.length === 0" class="text-slate-500 text-center py-8 text-xs">
                Click "Send Packet" to trigger virtual packet traversal across network namespaces.
              </div>
            </div>
          </div>

          <div class="text-[10px] text-slate-400 font-mono pt-3 border-t border-[#21262d] flex items-center justify-between">
            <span class="flex items-center gap-1"><Activity class="w-3 h-3 text-indigo-400" /> Latency: ~0.4ms</span>
            <span>MTU: 1500 bytes</span>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>