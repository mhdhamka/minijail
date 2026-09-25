<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Container } from '../../types/container';
import { 
  Network, 
  Layers, 
  FolderTree, 
  Cpu, 
  ShieldCheck, 
  AlertCircle, 
  Eye, 
  EyeOff, 
  ArrowRight, 
  Server, 
  Box, 
  Lock,
  Unlock,
  Globe
} from 'lucide-vue-next';

interface Props {
  containers?: Container[];
}

const props = withDefaults(defineProps<Props>(), {
  containers: () => [],
});

// Toggles for namespace boundaries
const showPidBoundary = ref(true);
const showNetBoundary = ref(true);
const showMntBoundary = ref(true);
const showUtsBoundary = ref(true);

// Active inspected node or boundary for the educational side-panel
type InspectionType = 'pid' | 'net' | 'mnt' | 'uts' | 'veth' | 'overlay';
const inspectedTarget = ref<{
  type: InspectionType;
  title: string;
  subtitle: string;
  hostView: string;
  containerView: string;
  kernelInode: string;
  syscall: string;
  isolationRiskWithout: string;
}>({
  type: 'pid',
  title: 'PID Namespace (Process ID Virtualization)',
  subtitle: 'CLONE_NEWPID / unshare(CLONE_NEWPID)',
  hostView: 'Sees the real Host PID (e.g. PID 42193) inside the root process tree along with systemd (PID 1) and dockerd (PID 842).',
  containerView: 'Virtualizes PID 1 inside the boundary. The container entrypoint thinks it is the init process of the whole machine.',
  kernelInode: 'pid:[4026535480]',
  syscall: 'clone(CLONE_NEWPID | SIGCHLD, ...)',
  isolationRiskWithout: 'Without PID namespace, container processes can inspect all host processes and send SIGKILL to host daemons or other containers!',
});

// Selected container for demonstration
const selectedContainerName = ref<string>('alpine-web');

const activeContainer = computed(() => {
  if (props.containers && props.containers.length > 0) {
    const match = props.containers.find(c => c.name === selectedContainerName.value);
    if (match) return match;
    return props.containers[0];
  }
  return {
    id: 'c7a8f912e34b1509d845e67041ab34fe20194851239857ab90e812d46e10f391',
    name: 'alpine-web',
    image: 'alpine:3.19',
    hostPid: 42193,
    containerPid: 1,
    portBindings: ['8080:80/tcp'],
    namespaces: {
      pidInode: 'pid:[4026535480]',
      netInode: 'net:[4026535483]',
      mntInode: 'mnt:[4026535482]',
      utsInode: 'uts:[4026535481]',
      hostname: 'alpine-web',
      virtualIp: '172.18.0.2/16',
      vethPair: 'veth_c7a8 <-> eth0',
    },
    rootfs: {
      lowerDir: '/var/lib/minijail/images/alpine:3.19/rootfs',
      upperDir: '/var/lib/minijail/containers/c7a8f912/diff',
      workDir: '/var/lib/minijail/containers/c7a8f912/work',
      mergedDir: '/var/lib/minijail/containers/c7a8f912/merged',
    },
    processes: [
      { hostPid: 42193, containerPid: 1, user: 'root', cpu: 2.1, memoryMb: 18.4, command: 'sh -c while true...' },
      { hostPid: 42194, containerPid: 8, user: 'root', cpu: 0.1, memoryMb: 4.2, command: 'sleep 4' },
    ],
  };
});

function selectInspect(type: InspectionType) {
  if (type === 'pid') {
    inspectedTarget.value = {
      type: 'pid',
      title: 'PID Namespace (Process Isolation)',
      subtitle: 'CLONE_NEWPID / procfs virtual tree',
      hostView: `Process running as Host PID ${activeContainer.value.hostPid || 42193}. Host admin can observe, strace, and manage it directly from the host.`,
      containerView: `Process views itself strictly as PID 1! Cannot see Host PID 1 (systemd) or other containers. A 'kill -9 1' inside exits the container container-wide.`,
      kernelInode: activeContainer.value.namespaces?.pidInode || 'pid:[4026535480]',
      syscall: 'unshare(CLONE_NEWPID) or clone(CLONE_NEWPID)',
      isolationRiskWithout: 'Without PID namespace: Any user in a container can see passwords/env in /proc/<pid>/cmdline of all host services and kill host processes.',
    };
  } else if (type === 'net') {
    inspectedTarget.value = {
      type: 'net',
      title: 'Network Namespace (Stack & Port Isolation)',
      subtitle: 'CLONE_NEWNET / veth pair & private IP routing',
      hostView: `Host bridge 'docker0' (172.18.0.1) connects via peer device '${activeContainer.value.namespaces?.vethPair?.split(' ')[0] || 'veth_c7a8'}'. Host IP is 192.168.1.100.`,
      containerView: `Private loopback 'lo' and isolated 'eth0' with container IP ${activeContainer.value.namespaces?.virtualIp || '172.18.0.2'}. Container binds port 80 without conflicting with other containers!`,
      kernelInode: activeContainer.value.namespaces?.netInode || 'net:[4026535483]',
      syscall: 'unshare(CLONE_NEWNET) + ip link add veth peer',
      isolationRiskWithout: 'Without NET namespace: Only ONE container could listen on port 80 or 443! Containers would share loopback (127.0.0.1) with host PostgreSQL or Redis.',
    };
  } else if (type === 'mnt') {
    inspectedTarget.value = {
      type: 'mnt',
      title: 'Mount Namespace (Filesystem Boundary)',
      subtitle: 'CLONE_NEWNS / pivot_root over OverlayFS',
      hostView: `Host rootfs remains at '/'. Host sees container rootfs at '${activeContainer.value.rootfs?.mergedDir || '/var/lib/minijail/containers/...'}' with lower and upper directories.`,
      containerView: `Container root '/' is imprisoned via pivot_root. Container sees a clean Alpine/Ubuntu filesystem tree and cannot access host '/etc/shadow' or host devices.`,
      kernelInode: activeContainer.value.namespaces?.mntInode || 'mnt:[4026535482]',
      syscall: 'unshare(CLONE_NEWNS) + syscall.PivotRoot(newRoot, oldRoot)',
      isolationRiskWithout: 'Without MNT namespace: An attacker executing `rm -rf /` or modifying `/etc/sudoers` inside the container would destroy or compromise the entire host operating system!',
    };
  } else if (type === 'uts') {
    inspectedTarget.value = {
      type: 'uts',
      title: 'UTS Namespace (Hostname & Domain)',
      subtitle: 'CLONE_NEWUTS / sethostname syscall',
      hostView: `Host hostname is 'linux-prod-node01'. Unaffected by container configuration.`,
      containerView: `Container hostname is '${activeContainer.value.name}'. Container can run 'hostname backend-01' without changing the host system name.`,
      kernelInode: activeContainer.value.namespaces?.utsInode || 'uts:[4026535481]',
      syscall: 'unshare(CLONE_NEWUTS) + sethostname(...)',
      isolationRiskWithout: 'Without UTS namespace: Services that bind or register based on local hostname would conflict or corrupt host networking logs.',
    };
  } else if (type === 'veth') {
    inspectedTarget.value = {
      type: 'veth',
      title: 'Virtual Ethernet Pair (veth)',
      subtitle: 'Linux Kernel Virtual Cable',
      hostView: `Host has 'veth_c7a8' plugged into bridge 'docker0' (IP 172.18.0.1). iptables performs DNAT: 0.0.0.0:8080 -> 172.18.0.2:80.`,
      containerView: `Container sees 'eth0' configured with 172.18.0.2/16 and default gateway 172.18.0.1. Traffic is forwarded transparently.`,
      kernelInode: 'net_device:veth',
      syscall: 'ip link add veth_host type veth peer name eth0 netns <pid>',
      isolationRiskWithout: 'Without virtual interfaces: Containers would bind directly to physical NIC eth0, exposing raw MAC and promiscuous sniffing.',
    };
  } else if (type === 'overlay') {
    inspectedTarget.value = {
      type: 'overlay',
      title: 'OverlayFS Layering (Copy-on-Write)',
      subtitle: 'lowerdir (read-only image) + upperdir (container diff)',
      hostView: `Storage driver keeps base image in lowerdir and captures changes in upperdir (/var/lib/minijail/containers/<id>/diff).`,
      containerView: `Container seamlessly views both as a unified filesystem in mergedDir. Modifying '/etc/hosts' copies up to upperdir without altering base image!`,
      kernelInode: 'fs:overlay',
      syscall: 'mount -t overlay overlay -o lowerdir=...,upperdir=...,workdir=... /merged',
      isolationRiskWithout: 'Without OverlayFS: Every container launch would require copying several gigabytes of files, wasting disk space and I/O.',
    };
  }
}

function toggleAll(enable: boolean) {
  showPidBoundary.value = enable;
  showNetBoundary.value = enable;
  showMntBoundary.value = enable;
  showUtsBoundary.value = enable;
}
</script>

<template>
  <div class="space-y-4 font-sans antialiased text-slate-100">
    <!-- Top Header and Diagram Controls -->
    <div class="bg-gradient-to-r from-slate-900/90 via-[#141A22] to-slate-900/90 border border-slate-800/80 rounded-2xl p-5 shadow-xl backdrop-blur-md">
      <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
              <Layers class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-base font-bold text-white tracking-tight flex items-center gap-2">
                Interactive Linux Namespaces Isolation
                <span class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-300 border border-cyan-500/30">
                  Kernel 6.6+
                </span>
              </h3>
              <p class="text-xs text-slate-400 mt-0.5">
                Simulate kernel isolation layers (<code class="text-emerald-300">CLONE_NEWPID</code>, <code class="text-indigo-300">CLONE_NEWNET</code>, <code class="text-amber-300">CLONE_NEWNS</code>) in real-time.
              </p>
            </div>
          </div>
        </div>

        <!-- Master Toggles -->
        <div class="flex items-center gap-2.5 self-start lg:self-auto">
          <button 
            @click="toggleAll(true)"
            class="px-3 py-2 rounded-xl text-xs font-mono font-medium bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 transition-all cursor-pointer flex items-center gap-2 shadow-sm"
          >
            <Lock class="w-3.5 h-3.5" />
            <span>Isolate All</span>
          </button>
          <button 
            @click="toggleAll(false)"
            class="px-3 py-2 rounded-xl text-xs font-mono font-medium bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/30 transition-all cursor-pointer flex items-center gap-2 shadow-sm"
          >
            <Unlock class="w-3.5 h-3.5" />
            <span>Remove All (Shared)</span>
          </button>
        </div>
      </div>

      <!-- Namespace Boundary Toggle Switches -->
      <div class="mt-5 pt-4 border-t border-slate-800/80 grid grid-cols-2 sm:grid-cols-4 gap-3">
        <!-- PID Toggle -->
        <button 
          @click="showPidBoundary = !showPidBoundary; selectInspect('pid')"
          class="p-3 rounded-xl border text-left transition-all cursor-pointer select-none relative overflow-hidden group"
          :class="showPidBoundary 
            ? 'bg-emerald-950/30 border-emerald-500/50 text-emerald-200 shadow-lg shadow-emerald-950/20' 
            : 'bg-slate-950/40 border-slate-800 text-slate-500 hover:border-slate-700'"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold font-mono flex items-center gap-2">
              <span class="w-2 h-2 rounded-full transition-colors" :class="showPidBoundary ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.8)] animate-pulse' : 'bg-slate-600'"></span>
              PID Namespace
            </span>
            <component :is="showPidBoundary ? Eye : EyeOff" class="w-4 h-4 opacity-70 group-hover:opacity-100 transition-opacity" />
          </div>
          <div class="text-[10px] mt-1.5 font-mono opacity-90">
            {{ showPidBoundary ? 'Active (PID 1 Virtualized)' : 'Exposed to Host' }}
          </div>
        </button>

        <!-- NET Toggle -->
        <button 
          @click="showNetBoundary = !showNetBoundary; selectInspect('net')"
          class="p-3 rounded-xl border text-left transition-all cursor-pointer select-none relative overflow-hidden group"
          :class="showNetBoundary 
            ? 'bg-indigo-950/30 border-indigo-500/50 text-indigo-200 shadow-lg shadow-indigo-950/20' 
            : 'bg-slate-950/40 border-slate-800 text-slate-500 hover:border-slate-700'"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold font-mono flex items-center gap-2">
              <span class="w-2 h-2 rounded-full transition-colors" :class="showNetBoundary ? 'bg-indigo-400 shadow-[0_0_8px_rgba(129,140,248,0.8)] animate-pulse' : 'bg-slate-600'"></span>
              NET Namespace
            </span>
            <component :is="showNetBoundary ? Eye : EyeOff" class="w-4 h-4 opacity-70 group-hover:opacity-100 transition-opacity" />
          </div>
          <div class="text-[10px] mt-1.5 font-mono opacity-90">
            {{ showNetBoundary ? 'Active (veth + 172.18.0.2)' : 'Host Network Shared' }}
          </div>
        </button>

        <!-- MNT Toggle -->
        <button 
          @click="showMntBoundary = !showMntBoundary; selectInspect('mnt')"
          class="p-3 rounded-xl border text-left transition-all cursor-pointer select-none relative overflow-hidden group"
          :class="showMntBoundary 
            ? 'bg-amber-950/30 border-amber-500/50 text-amber-200 shadow-lg shadow-amber-950/20' 
            : 'bg-slate-950/40 border-slate-800 text-slate-500 hover:border-slate-700'"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold font-mono flex items-center gap-2">
              <span class="w-2 h-2 rounded-full transition-colors" :class="showMntBoundary ? 'bg-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.8)] animate-pulse' : 'bg-slate-600'"></span>
              MNT Namespace
            </span>
            <component :is="showMntBoundary ? Eye : EyeOff" class="w-4 h-4 opacity-70 group-hover:opacity-100 transition-opacity" />
          </div>
          <div class="text-[10px] mt-1.5 font-mono opacity-90">
            {{ showMntBoundary ? 'Active (pivot_root jail)' : 'Host Root Exposed' }}
          </div>
        </button>

        <!-- UTS Toggle -->
        <button 
          @click="showUtsBoundary = !showUtsBoundary; selectInspect('uts')"
          class="p-3 rounded-xl border text-left transition-all cursor-pointer select-none relative overflow-hidden group"
          :class="showUtsBoundary 
            ? 'bg-cyan-950/30 border-cyan-500/50 text-cyan-200 shadow-lg shadow-cyan-950/20' 
            : 'bg-slate-950/40 border-slate-800 text-slate-500 hover:border-slate-700'"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold font-mono flex items-center gap-2">
              <span class="w-2 h-2 rounded-full transition-colors" :class="showUtsBoundary ? 'bg-cyan-400 shadow-[0_0_8px_rgba(34,211,238,0.8)] animate-pulse' : 'bg-slate-600'"></span>
              UTS Namespace
            </span>
            <component :is="showUtsBoundary ? Eye : EyeOff" class="w-4 h-4 opacity-70 group-hover:opacity-100 transition-opacity" />
          </div>
          <div class="text-[10px] mt-1.5 font-mono opacity-90 truncate">
            {{ showUtsBoundary ? activeContainer.name : 'Host Hostname' }}
          </div>
        </button>
      </div>
    </div>

    <!-- MAIN INTERACTIVE ARCHITECTURE CANVAS -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-4">
      
      <!-- LEFT 8 COLS: SCHEMATIC DIAGRAM -->
      <div class="lg:col-span-8 bg-slate-950/60 border border-slate-800/80 rounded-2xl p-5 relative overflow-hidden shadow-xl backdrop-blur-md">
        
        <!-- Host OS Canvas Border Title -->
        <div class="flex items-center justify-between pb-3.5 mb-4 border-b border-slate-800/80">
          <div class="flex items-center gap-2.5">
            <Server class="w-4 h-4 text-slate-400" />
            <span class="text-xs font-bold font-mono text-slate-200 uppercase tracking-wider">
              Host Linux Kernel Space
            </span>
          </div>

          <div class="flex items-center gap-3 text-[11px] font-mono text-slate-400">
            <span class="flex items-center gap-1.5"><Globe class="w-3.5 h-3.5 text-slate-500" /> 192.168.1.100</span>
            <span class="text-slate-700">|</span>
            <span class="text-slate-300 font-semibold">node01</span>
          </div>
        </div>

        <!-- Diagram Grid: Host Entities vs Container Jails -->
        <div class="space-y-4">
          
          <!-- 1. PID LAYER -->
          <div 
            @click="selectInspect('pid')"
            class="p-4 rounded-xl border transition-all cursor-pointer group hover:border-slate-600"
            :class="showPidBoundary 
              ? 'bg-slate-900/80 border-emerald-500/40 shadow-lg shadow-emerald-950/10' 
              : 'bg-rose-950/10 border-dashed border-rose-500/50'"
          >
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <div class="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400">
                  <Cpu class="w-4 h-4" />
                </div>
                <div>
                  <span class="text-xs font-bold font-mono text-emerald-300 block">
                    Process Isolation (PID Namespace)
                  </span>
                  <span v-if="!showPidBoundary" class="text-[10px] text-rose-400 font-mono">
                    ⚠️ Unsafe: Container can view host process tree
                  </span>
                </div>
              </div>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-slate-400">
                {{ showPidBoundary ? activeContainer.namespaces?.pidInode : 'root_pid_ns' }}
              </span>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <!-- Host View Process Box -->
              <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 font-mono text-xs space-y-2">
                <div class="text-[10px] text-slate-400 uppercase font-bold flex items-center justify-between">
                  <span>Host OS Process Tree</span>
                  <span class="text-slate-500">Global PIDs</span>
                </div>
                <div class="space-y-1.5 text-[11px]">
                  <div class="flex items-center justify-between text-slate-400">
                    <span>PID 1</span>
                    <span class="text-slate-500">/sbin/init (systemd)</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-400">
                    <span>PID 842</span>
                    <span class="text-slate-500">dockerd daemon</span>
                  </div>
                  <div class="flex items-center justify-between p-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 font-medium">
                    <span>PID {{ activeContainer.hostPid || 42193 }}</span>
                    <span>{{ activeContainer.name }} (entrypoint)</span>
                  </div>
                </div>
              </div>

              <!-- Container View Process Box -->
              <div 
                class="p-3 rounded-xl border font-mono text-xs space-y-2 transition-all"
                :class="showPidBoundary 
                  ? 'bg-emerald-950/20 border-emerald-500/40 text-emerald-200' 
                  : 'bg-rose-950/20 border-rose-500/30 text-rose-200'"
              >
                <div class="text-[10px] uppercase font-bold flex items-center justify-between" :class="showPidBoundary ? 'text-emerald-400' : 'text-rose-400'">
                  <span>Container Perspective</span>
                  <span>{{ showPidBoundary ? 'Virtual PID Tree' : 'Host Exposed' }}</span>
                </div>

                <div v-if="showPidBoundary" class="space-y-1.5 text-[11px]">
                  <div class="flex items-center justify-between p-1.5 rounded-lg bg-emerald-500/20 text-emerald-300 font-bold border border-emerald-500/30">
                    <span class="flex items-center gap-1.5">
                      <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                      PID 1
                    </span>
                    <span>sh -c 'while true...'</span>
                  </div>
                  <div class="text-[10px] text-emerald-400/80 pt-1">
                    ✓ Hidden from host systemd & processes
                  </div>
                </div>

                <div v-else class="space-y-1.5 text-[11px] text-rose-300">
                  <div class="p-2 rounded-lg bg-rose-500/10 border border-rose-500/20 text-[10px] leading-relaxed">
                    No PID barrier: Container processes map directly to host space, granting full visibility over system operations.
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 2. NETWORK LAYER -->
          <div 
            @click="selectInspect('net')"
            class="p-4 rounded-xl border transition-all cursor-pointer group hover:border-slate-600"
            :class="showNetBoundary 
              ? 'bg-slate-900/80 border-indigo-500/40 shadow-lg shadow-indigo-950/10' 
              : 'bg-rose-950/10 border-dashed border-rose-500/50'"
          >
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <div class="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400">
                  <Network class="w-4 h-4" />
                </div>
                <div>
                  <span class="text-xs font-bold font-mono text-indigo-300 block">
                    Network Isolation (NET Namespace)
                  </span>
                  <span v-if="!showNetBoundary" class="text-[10px] text-rose-400 font-mono">
                    ⚠️ Unsafe: Direct host NIC port binding risk
                  </span>
                </div>
              </div>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-slate-400">
                {{ showNetBoundary ? activeContainer.namespaces?.netInode : 'root_net_ns' }}
              </span>
            </div>

            <!-- Visual Cable Graphic -->
            <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-3 text-xs font-mono">
              <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
                
                <!-- Host Bridge Interface -->
                <div class="p-2.5 rounded-xl bg-slate-900 border border-slate-800 w-full sm:w-auto text-center sm:text-left">
                  <div class="text-[10px] text-slate-400">Host Bridge: <span class="text-indigo-300 font-bold">docker0</span></div>
                  <div class="text-[11px] text-slate-200 mt-0.5">172.18.0.1/16</div>
                </div>

                <!-- Cable Connection (veth pair) -->
                <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-indigo-950/40 border border-indigo-800/60 text-[11px] text-indigo-300">
                  <span class="font-bold">{{ showNetBoundary ? 'veth_c7a8' : 'direct eth0' }}</span>
                  <ArrowRight class="w-3.5 h-3.5 text-indigo-400" />
                  <span class="font-bold">{{ showNetBoundary ? 'eth0' : 'host nic' }}</span>
                </div>

                <!-- Container Interface -->
                <div 
                  class="p-2.5 rounded-xl border w-full sm:w-auto text-center sm:text-left transition-all"
                  :class="showNetBoundary 
                    ? 'bg-indigo-950/30 border-indigo-500/50 text-indigo-200' 
                    : 'bg-rose-950/30 border-rose-500/50 text-rose-300'"
                >
                  <div class="text-[10px] uppercase font-bold" :class="showNetBoundary ? 'text-indigo-400' : 'text-rose-400'">
                    {{ showNetBoundary ? 'Container IP' : 'Host Interface' }}
                  </div>
                  <div class="text-[11px] font-bold mt-0.5">
                    {{ showNetBoundary ? activeContainer.namespaces?.virtualIp : '192.168.1.100' }}
                  </div>
                </div>
              </div>

              <div class="text-[10px] text-slate-400 flex items-center justify-between pt-2 border-t border-slate-900">
                <span>iptables DNAT: <code class="text-indigo-300">tcp:8080 -> 172.18.0.2:80</code></span>
                <span class="text-slate-500">Loopback: 127.0.0.1 (private)</span>
              </div>
            </div>
          </div>

          <!-- 3. MOUNT & ROOTFS LAYER -->
          <div 
            @click="selectInspect('mnt')"
            class="p-4 rounded-xl border transition-all cursor-pointer group hover:border-slate-600"
            :class="showMntBoundary 
              ? 'bg-slate-900/80 border-amber-500/40 shadow-lg shadow-amber-950/10' 
              : 'bg-rose-950/10 border-dashed border-rose-500/50'"
          >
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <div class="p-1.5 rounded-lg bg-amber-500/10 text-amber-400">
                  <FolderTree class="w-4 h-4" />
                </div>
                <div>
                  <span class="text-xs font-bold font-mono text-amber-300 block">
                    Filesystem Jail (Mount Namespace + pivot_root)
                  </span>
                  <span v-if="!showMntBoundary" class="text-[10px] text-rose-400 font-mono">
                    ⚠️ Unsafe: Full host root filesystem exposure
                  </span>
                </div>
              </div>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-slate-400">
                {{ showMntBoundary ? activeContainer.namespaces?.mntInode : 'root_mnt_ns' }}
              </span>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs font-mono">
              <!-- OverlayFS Layer Stack -->
              <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-2">
                <div class="text-[10px] text-slate-400 uppercase font-bold flex items-center justify-between">
                  <span>OverlayFS Stack</span>
                  <span class="text-amber-400">Copy-on-Write</span>
                </div>
                <div class="space-y-1.5 text-[11px]">
                  <div class="p-1.5 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-300 flex items-center justify-between">
                    <span>Upper Layer (Read-Write)</span>
                    <span class="text-[10px]">diff/</span>
                  </div>
                  <div class="p-1.5 rounded-lg bg-slate-900 text-slate-300 flex items-center justify-between">
                    <span>Lower Layer (Read-Only)</span>
                    <span class="text-[10px] text-slate-400">alpine:3.19</span>
                  </div>
                </div>
              </div>

              <!-- Jailed Root vs Host Root -->
              <div 
                class="p-3 rounded-xl border space-y-2 transition-all"
                :class="showMntBoundary 
                  ? 'bg-amber-950/20 border-amber-500/40 text-amber-200' 
                  : 'bg-rose-950/20 border-rose-500/30 text-rose-200'"
              >
                <div class="text-[10px] uppercase font-bold" :class="showMntBoundary ? 'text-amber-400' : 'text-rose-400'">
                  {{ showMntBoundary ? 'pivot_root Isolation' : 'No Filesystem Jail' }}
                </div>
                <div v-if="showMntBoundary" class="space-y-1 text-[11px]">
                  <div>Container Root: <code class="text-amber-300">/</code></div>
                  <div class="text-slate-400">Host <code class="text-slate-300">/etc/shadow</code>: <span class="text-emerald-400 font-bold">Protected</span></div>
                </div>
                <div v-else class="text-[11px] text-rose-300 leading-relaxed">
                  Host root mounted directly. Commands executed inside have unhindered access to host storage configuration.
                </div>
              </div>
            </div>
          </div>

        </div>
      </div>

      <!-- RIGHT 4 COLS: INTERACTIVE KERNEL INSPECTION PANEL -->
      <div class="lg:col-span-4 bg-slate-950/80 border border-slate-800/80 rounded-2xl p-5 flex flex-col justify-between space-y-4 shadow-xl backdrop-blur-md">
        <div class="space-y-4">
          
          <!-- Inspected Item Title -->
          <div class="border-b border-slate-800/80 pb-3.5">
            <div class="flex items-center gap-2">
              <div class="p-1.5 rounded-lg bg-cyan-500/10 text-cyan-400">
                <ShieldCheck class="w-4 h-4" />
              </div>
              <h4 class="text-sm font-bold text-white tracking-tight">{{ inspectedTarget.title }}</h4>
            </div>
            <div class="text-xs text-cyan-400 font-mono mt-1">{{ inspectedTarget.subtitle }}</div>
          </div>

          <!-- Kernel Inode & Syscall Badges -->
          <div class="grid grid-cols-2 gap-2 text-xs font-mono">
            <div class="p-2.5 rounded-xl bg-slate-900/90 border border-slate-800">
              <div class="text-[10px] text-slate-500 uppercase font-semibold">Kernel Inode</div>
              <div class="text-emerald-400 font-bold mt-1 truncate">{{ inspectedTarget.kernelInode }}</div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-900/90 border border-slate-800">
              <div class="text-[10px] text-slate-500 uppercase font-semibold">Syscall</div>
              <div class="text-cyan-400 font-bold mt-1 truncate">{{ inspectedTarget.syscall.split(' ')[0] }}</div>
            </div>
          </div>

          <!-- Dual Perspectives -->
          <div class="space-y-3 text-xs">
            <!-- Host View Box -->
            <div class="p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <div class="text-[10px] font-mono uppercase font-bold text-slate-400 mb-1 flex items-center gap-2">
                <Server class="w-3.5 h-3.5 text-slate-400" />
                <span>Host OS Perspective</span>
              </div>
              <p class="text-slate-300 leading-relaxed text-[11px]">
                {{ inspectedTarget.hostView }}
              </p>
            </div>

            <!-- Container View Box -->
            <div class="p-3 rounded-xl bg-cyan-950/10 border border-cyan-900/30">
              <div class="text-[10px] font-mono uppercase font-bold text-cyan-400 mb-1 flex items-center gap-2">
                <Box class="w-3.5 h-3.5 text-cyan-400" />
                <span>Container Jailed Perspective</span>
              </div>
              <p class="text-slate-300 leading-relaxed text-[11px]">
                {{ inspectedTarget.containerView }}
              </p>
            </div>

            <!-- Security Vulnerability Box -->
            <div class="p-3 rounded-xl bg-rose-950/20 border border-rose-900/40 text-rose-300 text-[11px] leading-relaxed">
              <div class="text-[10px] font-mono uppercase font-bold text-rose-400 mb-1 flex items-center gap-1.5">
                <AlertCircle class="w-3.5 h-3.5" />
                <span>Risk Without Boundary:</span>
              </div>
              {{ inspectedTarget.isolationRiskWithout }}
            </div>
          </div>

          <!-- Syscall Code Snippet -->
          <div class="p-3 rounded-xl bg-slate-900 border border-slate-800 font-mono text-[11px] text-cyan-300 overflow-x-auto shadow-inner">
            <span class="text-slate-500">// Engine System Call Configuration</span><br/>
            <code>cmd.SysProcAttr = &syscall.SysProcAttr{<br/>
            &nbsp;&nbsp;Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,<br/>
            }</code>
          </div>
        </div>

        <!-- Quick Select Buttons -->
        <div class="pt-4 border-t border-slate-800/80">
          <div class="text-[10px] text-slate-400 font-mono uppercase mb-2 flex items-center gap-1.5">
            <span>Quick Inspect Layer:</span>
          </div>
          <div class="flex flex-wrap gap-1.5">
            <button 
              @click="selectInspect('pid')" 
              class="px-2.5 py-1.5 rounded-lg text-[10px] font-mono font-medium transition-all cursor-pointer"
              :class="inspectedTarget.type === 'pid' ? 'bg-emerald-600 text-white shadow-md shadow-emerald-900/20' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
            >
              PID
            </button>
            <button 
              @click="selectInspect('net')" 
              class="px-2.5 py-1.5 rounded-lg text-[10px] font-mono font-medium transition-all cursor-pointer"
              :class="inspectedTarget.type === 'net' ? 'bg-indigo-600 text-white shadow-md shadow-indigo-900/20' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
            >
              NET
            </button>
            <button 
              @click="selectInspect('mnt')" 
              class="px-2.5 py-1.5 rounded-lg text-[10px] font-mono font-medium transition-all cursor-pointer"
              :class="inspectedTarget.type === 'mnt' ? 'bg-amber-600 text-white shadow-md shadow-amber-900/20' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
            >
              MNT
            </button>
            <button 
              @click="selectInspect('veth')" 
              class="px-2.5 py-1.5 rounded-lg text-[10px] font-mono font-medium transition-all cursor-pointer"
              :class="inspectedTarget.type === 'veth' ? 'bg-indigo-600 text-white shadow-md shadow-indigo-900/20' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
            >
              veth Cable
            </button>
            <button 
              @click="selectInspect('overlay')" 
              class="px-2.5 py-1.5 rounded-lg text-[10px] font-mono font-medium transition-all cursor-pointer"
              :class="inspectedTarget.type === 'overlay' ? 'bg-amber-600 text-white shadow-md shadow-amber-900/20' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
            >
              OverlayFS
            </button>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>