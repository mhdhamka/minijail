<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Play, 
  Pause, 
  RotateCcw, 
  CheckCircle2, 
  Terminal, 
  Cpu, 
  ArrowRight,
  Activity,
  FileCode2,
  Workflow
} from 'lucide-vue-next';

interface LifecycleStep {
  step: number;
  title: string;
  subsystem: string;
  color: string;
  badgeBg: string;
  dockerStage: string;
  syscall: string;
  description: string;
  kernelState: string;
  interactiveType: 'overlay' | 'union' | 'namespace' | 'cgroup' | 'network' | 'pivot' | 'exec';
  details: { label: string; value: string; live?: string }[];
}

const steps: LifecycleStep[] = [
  {
    step: 1,
    title: 'OCI Image Resolution & Layer Unpack',
    subsystem: 'Storage Driver (OverlayFS)',
    color: '#38bdf8',
    badgeBg: 'bg-sky-500/20 text-sky-400 border-sky-500/40',
    dockerStage: 'docker run -d --name web -p 8080:80 alpine:3.19',
    syscall: 'openat() + tar unpack + sha256 checksum verification',
    description: 'Docker Engine checks local content-addressable storage cache. Unpacks immutable lower layer tarballs into /var/lib/minijail/images/alpine:3.19/rootfs.',
    kernelState: 'Image blobs verified; lowerdir layers mapped in filesystem table.',
    interactiveType: 'overlay',
    details: [
      { label: 'Base Layer Inode', value: 'ext4 Inode 918237' },
      { label: 'Layer Read Mode', value: 'O_RDONLY (Immutable)' },
      { label: 'Storage Driver', value: 'overlay2' },
    ],
  },
  {
    step: 2,
    title: 'OverlayFS Union Mount Composition',
    subsystem: 'VFS & Filesystem',
    color: '#f59e0b',
    badgeBg: 'bg-amber-500/20 text-amber-400 border-amber-500/40',
    dockerStage: 'Preparing container rootfs with copy-on-write diff',
    syscall: 'mount("overlay", "/merged", "overlay", MS_NODEV, "lowerdir=...,upperdir=...,workdir=...")',
    description: 'Creates an ephemeral upperdir diff directory and workdir. Issues kernel mount syscall to overlay the read-only image layers under a merged directory.',
    kernelState: 'New super_block registered; VFS resolves file lookups through upper -> lower chain.',
    interactiveType: 'union',
    details: [
      { label: 'Upperdir', value: '/var/lib/minijail/containers/c7a8/diff' },
      { label: 'Workdir', value: '/var/lib/minijail/containers/c7a8/work' },
      { label: 'Target Mount', value: '/var/lib/minijail/containers/c7a8/merged' },
    ],
  },
  {
    step: 3,
    title: 'Namespace Boundary Unsharing',
    subsystem: 'Linux Namespaces (nsproxy)',
    color: '#10b981',
    badgeBg: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40',
    dockerStage: 'Spawning container runner process (containerd-shim)',
    syscall: 'clone(SIGCHLD | CLONE_NEWPID | CLONE_NEWNET | CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWIPC, ...)',
    description: 'Kernel allocates a brand new nsproxy struct with fresh namespace pointers. The child process receives its own virtual PID 1, network stack, and mount table.',
    kernelState: 'Allocated new task_struct; nsproxy->pid_ns pointing to new pid namespace (Inode 4026535480).',
    interactiveType: 'namespace',
    details: [
      { label: 'PID Namespace', value: 'pid:[4026535480]' },
      { label: 'NET Namespace', value: 'net:[4026535483]' },
      { label: 'MNT Namespace', value: 'mnt:[4026535482]' },
      { label: 'UTS Namespace', value: 'uts:[4026535481]' },
    ],
  },
  {
    step: 4,
    title: 'cgroups v2 Hierarchy & Quota Enforcement',
    subsystem: 'Control Groups (cgroups v2)',
    color: '#818cf8',
    badgeBg: 'bg-indigo-500/20 text-indigo-400 border-indigo-500/40',
    dockerStage: 'Applying CPU quota and memory bounds',
    syscall: 'mkdir(/sys/fs/cgroup/minijail/c7a8) && write(cgroup.procs, pid)',
    description: 'Runc writes memory.max and cpu.max parameters. Process host PID is added to cgroup.procs to enforce Completely Fair Scheduler (CFS) quotas and OOM constraints.',
    kernelState: 'cgroup attached to task_struct; CFS bandwidth timer initialized.',
    interactiveType: 'cgroup',
    details: [
      { label: 'cpu.max', value: '50000 100000 (0.50 cores)', live: '50ms active' },
      { label: 'memory.max', value: '67108864 (64 MB ceiling)', live: '14.2 MB used' },
      { label: 'cgroup.procs', value: 'Host PID 42193 attached' },
    ],
  },
  {
    step: 5,
    title: 'Virtual Ethernet (veth) Wiring & Port Forwarding',
    subsystem: 'Network Core & iptables NAT',
    color: '#06b6d4',
    badgeBg: 'bg-cyan-500/20 text-cyan-400 border-cyan-500/40',
    dockerStage: 'Configuring bridge docker0 and veth pair',
    syscall: 'ip link add veth_c7a8 type veth peer name eth0 netns <pid>',
    description: 'Docker creates a virtual Ethernet pair. Moves one end into the container network namespace (renamed to eth0) and attaches peer veth_c7a8 to bridge docker0 (172.18.0.1).',
    kernelState: 'veth interfaces bound; netfilter NAT PREROUTING table updated for 8080:80.',
    interactiveType: 'network',
    details: [
      { label: 'Host Interface', value: 'veth_c7a8 (Bridge docker0)' },
      { label: 'Container Interface', value: 'eth0 (172.18.0.2/16)' },
      { label: 'iptables DNAT', value: '0.0.0.0:8080 -> 172.18.0.2:80' },
    ],
  },
  {
    step: 6,
    title: 'Filesystem Imprisonment (pivot_root)',
    subsystem: 'VFS Mount Jail',
    color: '#eab308',
    badgeBg: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/40',
    dockerStage: 'Locking process inside container root directory',
    syscall: 'pivot_root("/merged", "/merged/.oldroot") + umount2("/.oldroot", MNT_DETACH)',
    description: 'Swaps current process rootfs with the OverlayFS merged mount. Unmounts the host oldroot and frees references so container process cannot traverse back up.',
    kernelState: 'current->fs->root updated to merged root; host filesystem completely unreachable.',
    interactiveType: 'pivot',
    details: [
      { label: 'New Root /', value: 'OverlayFS /merged' },
      { label: 'Host Root Traversal', value: 'Blocked (Cannot break out)' },
      { label: 'procfs & sysfs', value: 'Remounted inside new namespace' },
    ],
  },
  {
    step: 7,
    title: 'Execve Entrypoint: Replacing Runc with Application',
    subsystem: 'Process Execution',
    color: '#ec4899',
    badgeBg: 'bg-pink-500/20 text-pink-400 border-pink-500/40',
    dockerStage: 'Container entrypoint running as PID 1',
    syscall: 'execve("/bin/sh", ["sh", "-c", "while true; do ..."], envp)',
    description: 'The bootstrap launcher runc overwrites its own memory space with the user-specified entrypoint binary. The application takes over as PID 1.',
    kernelState: 'ELF binary loaded; process starts execution loop in user mode as container PID 1.',
    interactiveType: 'exec',
    details: [
      { label: 'Container PID', value: '1 (App init)' },
      { label: 'Host PID', value: '42193 (Kernel task table)' },
      { label: 'Container Status', value: 'Running (Active loop)' },
    ],
  },
];

const currentStepIndex = ref(0);
const isPlaying = ref(false);
let playbackTimer: any = null;

// Interactive state toggles for extreme engagement
const interactiveLog = ref<string[]>([
  '[kernel] Initialized container boot sequence manager v2.4',
  '[containerd-shim] Ready to orchestrate clone & mount syscalls'
]);

// Interactive simulation triggers per step
const simState = ref({
  tarUnpacked: false,
  cowTriggered: false,
  namespaceUnshared: false,
  cpuThrottled: false,
  packetSent: false,
  pivotJailed: false,
  pidOneRunning: false
});

const currentStep = computed(() => steps[currentStepIndex.value]);

function logAction(msg: string) {
  interactiveLog.value.unshift(`[${new Date().toLocaleTimeString()}] ${msg}`);
  if (interactiveLog.value.length > 6) interactiveLog.value.pop();
}

function nextStep() {
  if (currentStepIndex.value < steps.length - 1) {
    currentStepIndex.value++;
    logAction(`Transitioned to Phase 0${currentStepIndex.value + 1}: ${steps[currentStepIndex.value].subsystem}`);
  } else {
    currentStepIndex.value = 0;
    logAction('Reset container lifecycle loop to Phase 01.');
  }
}

function prevStep() {
  if (currentStepIndex.value > 0) {
    currentStepIndex.value--;
    logAction(`Rolled back to Phase 0${currentStepIndex.value + 1}`);
  }
}

function jumpToStep(idx: number) {
  currentStepIndex.value = idx;
  logAction(`Jumped directly to Phase 0${idx + 1}: ${steps[idx].subsystem}`);
}

function togglePlayback() {
  if (isPlaying.value) {
    clearInterval(playbackTimer);
    isPlaying.value = false;
    logAction('Paused automated kernel lifecycle simulation.');
  } else {
    isPlaying.value = true;
    logAction('Started automated millisecond kernel playback.');
    playbackTimer = setInterval(() => {
      if (currentStepIndex.value < steps.length - 1) {
        currentStepIndex.value++;
      } else {
        currentStepIndex.value = 0;
      }
    }, 3200);
  }
}

function resetPlayback() {
  clearInterval(playbackTimer);
  isPlaying.value = false;
  currentStepIndex.value = 0;
  simState.value = {
    tarUnpacked: false,
    cowTriggered: false,
    namespaceUnshared: false,
    cpuThrottled: false,
    packetSent: false,
    pivotJailed: false,
    pidOneRunning: false
  };
  logAction('Reset lifecycle simulation state back to zero.');
}

// Step-specific interactive triggers
function triggerStepAction() {
  const t = currentStep.value.interactiveType;
  if (t === 'overlay') {
    simState.value.tarUnpacked = true;
    logAction('Successfully extracted sha256:7b92 layer tarballs into read-only lowerdir.');
  } else if (t === 'union') {
    simState.value.cowTriggered = true;
    logAction('Copy-on-Write upperdir diff active. Container modifications will now write locally.');
  } else if (t === 'namespace') {
    simState.value.namespaceUnshared = true;
    logAction('Kernel clone() unshared nsproxy. New PID namespace created (pid:4026535480).');
  } else if (t === 'cgroup') {
    simState.value.cpuThrottled = true;
    logAction('cgroups v2 quota enforced: memory capped at 64MB, CPU capped at 50% core.');
  } else if (t === 'network') {
    simState.value.packetSent = true;
    logAction('Synthesized HTTP GET packet forwarded through veth pair to container port 80.');
  } else if (t === 'pivot') {
    simState.value.pivotJailed = true;
    logAction('pivot_root executed successfully. Host rootfs successfully locked out.');
  } else if (t === 'exec') {
    simState.value.pidOneRunning = true;
    logAction('execve() completed. Application process is now running as PID 1 inside container.');
  }
}
</script>

<template>
  <div class="space-y-6 font-sans select-none">
    
    <!-- Top Bar Header & Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 rounded-2xl bg-[#161B22] border border-[#21262d] shadow-lg">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-blue-500/10 border border-blue-500/30 text-blue-400">
            <Workflow class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white font-sans flex items-center gap-2">
              <span>Interactive Container Lifecycle Machine</span>
              <span class="text-[10px] px-2 py-0.5 rounded font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                Kernel Live
              </span>
            </h2>
            <p class="text-xs text-slate-400">
              Step through the 7 phases of a container's birth, inspect live kernel registries, and test interactive syscall triggers.
            </p>
          </div>
        </div>
      </div>

      <!-- Playback Controls -->
      <div class="flex items-center gap-2 flex-wrap">
        <button 
          @click="prevStep"
          :disabled="currentStepIndex === 0"
          class="px-3 py-2 rounded-xl bg-[#0D1117] hover:bg-[#202735] text-slate-300 disabled:opacity-40 border border-[#2D3848] text-xs font-mono transition-all cursor-pointer shadow-sm"
        >
          ← Prev
        </button>
        <button 
          @click="togglePlayback"
          class="px-4 py-2 rounded-xl text-xs font-semibold flex items-center gap-2 transition-all cursor-pointer text-white shadow-md"
          :class="isPlaying ? 'bg-amber-600 hover:bg-amber-500 ring-2 ring-amber-500/30' : 'bg-blue-600 hover:bg-blue-500 ring-2 ring-blue-500/30'"
        >
          <component :is="isPlaying ? Pause : Play" class="w-3.5 h-3.5" />
          <span>{{ isPlaying ? 'Pause Simulation' : 'Auto Play Engine' }}</span>
        </button>
        <button 
          @click="nextStep"
          class="px-3 py-2 rounded-xl bg-[#0D1117] hover:bg-[#202735] text-slate-300 border border-[#2D3848] text-xs font-mono transition-all cursor-pointer shadow-sm"
        >
          Next →
        </button>
        <button 
          @click="resetPlayback"
          class="p-2 rounded-xl bg-[#0D1117] hover:bg-[#202735] text-slate-400 hover:text-white border border-[#2D3848] transition-all cursor-pointer"
          title="Reset to Phase 1"
        >
          <RotateCcw class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- MODERNIZED TIMELINE PROCESS FLOW BAR -->
    <div class="relative bg-[#0F131C] border border-[#212836] rounded-2xl p-4 shadow-xl overflow-x-auto">
      <div class="absolute top-1/2 left-8 right-8 h-0.5 bg-[#1F2937] -translate-y-1/2 hidden lg:block z-0 pointer-events-none"></div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-7 gap-3 relative z-10">
        <div 
          v-for="(st, idx) in steps" 
          :key="st.step"
          @click="jumpToStep(idx)"
          class="group relative flex flex-col p-3.5 rounded-xl border transition-all cursor-pointer text-left backdrop-blur-md"
          :class="[
            idx === currentStepIndex 
              ? 'bg-[#182338] border-blue-500 shadow-xl shadow-blue-500/15 ring-2 ring-blue-500/40 scale-[1.02]' 
              : idx < currentStepIndex 
              ? 'bg-[#121824] border-[#253247] hover:border-slate-400' 
              : 'bg-[#0A0E17]/80 border-[#1A2230] hover:border-slate-600'
          ]"
        >
          <!-- Top Row: Phase Indicator & Status Icon -->
          <div class="flex items-center justify-between mb-2">
            <span 
              class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold tracking-wider"
              :style="{
                backgroundColor: idx === currentStepIndex ? st.color + '25' : '#1A2332',
                color: idx === currentStepIndex ? st.color : '#94A3B8',
                border: `1px solid ${idx === currentStepIndex ? st.color + '50' : '#2A374A'}`
              }"
            >
              PHASE 0{{ st.step }}
            </span>
            <div 
              class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold font-mono transition-transform group-hover:scale-110"
              :class="[
                idx < currentStepIndex ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40' :
                idx === currentStepIndex ? 'bg-blue-500 text-white shadow-sm' : 'bg-[#1A2332] text-slate-500 border border-[#2A374A]'
              ]"
            >
              <CheckCircle2 v-if="idx < currentStepIndex" class="w-3.5 h-3.5 text-emerald-400" />
              <span v-else>{{ st.step }}</span>
            </div>
          </div>

          <!-- Phase Subsystem Title -->
          <div class="text-xs font-bold text-slate-100 font-sans tracking-tight line-clamp-2 mt-0.5 group-hover:text-blue-300 transition-colors">
            {{ st.subsystem }}
          </div>
          
          <!-- Short label preview -->
          <div class="text-[10px] text-slate-400 font-mono mt-2 truncate">
            {{ st.title.split(' ')[0] }} {{ st.title.split(' ')[1] || '' }}
          </div>

          <!-- Active Glow Indicator Line at Bottom -->
          <div 
            v-if="idx === currentStepIndex" 
            class="absolute bottom-0 left-3 right-3 h-1 rounded-full shadow-md" 
            :style="{ backgroundColor: st.color }"
          ></div>
        </div>
      </div>
    </div>

    <!-- MAIN INTERACTIVE WORKBENCH: 2 COLUMNS -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: DEEP KERNEL SCHEMATIC & INTERACTIVE SIMULATOR -->
      <div class="lg:col-span-7 bg-[#141A22] border border-[#21262d] rounded-2xl p-5 space-y-4 shadow-xl flex flex-col justify-between">
        <div class="space-y-4">
          
          <!-- Step Banner Header -->
          <div class="flex items-start justify-between border-b border-[#21262d] pb-4">
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <span 
                  class="w-7 h-7 rounded-xl flex items-center justify-center text-xs font-bold font-mono text-black shadow-sm"
                  :style="{ backgroundColor: currentStep.color }"
                >
                  {{ currentStep.step }}
                </span>
                <h3 class="text-base font-bold text-white font-sans">{{ currentStep.title }}</h3>
              </div>
              <div class="text-xs font-mono font-semibold pl-9" :style="{ color: currentStep.color }">
                Subsystem: {{ currentStep.subsystem }}
              </div>
            </div>
            <span class="text-xs px-3 py-1 rounded-xl bg-[#0D1117] border border-[#2D3848] font-mono text-slate-300 shadow-inner">
              Phase 0{{ currentStep.step }} / 07
            </span>
          </div>

          <!-- Description -->
          <p class="text-slate-300 text-xs font-sans leading-relaxed">
            {{ currentStep.description }}
          </p>

          <!-- Docker CLI Command Context -->
          <div class="p-3.5 rounded-xl bg-[#0D1117] border border-[#21262d] space-y-1 font-mono text-xs">
            <div class="text-[10px] text-slate-400 uppercase font-bold flex items-center gap-1.5">
              <Terminal class="w-3.5 h-3.5 text-cyan-400" />
              <span>High-Level Docker CLI Context:</span>
            </div>
            <div class="text-cyan-300 font-bold tracking-tight">{{ currentStep.dockerStage }}</div>
          </div>

          <!-- Kernel Syscall Code Window -->
          <div class="rounded-xl border border-[#21262d] bg-[#0A0D12] overflow-hidden shadow-inner">
            <div class="px-3.5 py-2 bg-[#121822] border-b border-[#21262d] flex items-center justify-between text-[11px] font-mono">
              <span class="text-slate-400 flex items-center gap-1.5">
                <FileCode2 class="w-3.5 h-3.5 text-emerald-400" />
                <span>Actual Linux Kernel Syscall:</span>
              </span>
              <span class="text-emerald-400 font-bold text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">C Runtime</span>
            </div>
            <div class="p-3.5 text-emerald-300 font-mono text-xs overflow-x-auto">
              <code>{{ currentStep.syscall }}</code>
            </div>
          </div>

          <!-- Interactive Action Trigger Box -->
          <div class="p-4 rounded-xl bg-gradient-to-r from-blue-950/40 to-indigo-950/40 border border-blue-500/30 flex items-center justify-between">
            <div class="space-y-0.5">
              <div class="text-xs font-bold text-white font-sans flex items-center gap-1.5">
                <span>Trigger Phase {{ currentStep.step }} Kernel Operation</span>
              </div>
              <div class="text-[11px] text-slate-400">
                Execute interactive system call simulation in real-time memory.
              </div>
            </div>
            <button 
              @click="triggerStepAction"
              class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold shadow-lg shadow-blue-500/20 transition-all cursor-pointer flex items-center gap-1.5 shrink-0"
            >
              <span>Execute Syscall</span>
              <ArrowRight class="w-3.5 h-3.5" />
            </button>
          </div>

        </div>

        <!-- Live Kernel Event Stream Log -->
        <div class="mt-4 pt-4 border-t border-[#21262d] space-y-2">
          <div class="flex items-center justify-between text-[11px] font-mono text-slate-400 uppercase font-bold">
            <span class="flex items-center gap-1.5">
              <Activity class="w-3.5 h-3.5 text-emerald-400" />
              <span>Live Kernel dmesg / Audit Log:</span>
            </span>
            <span class="text-emerald-400"> Streaming</span>
          </div>
          <div class="p-3 rounded-xl bg-[#090D14] border border-[#1D2430] font-mono text-[11px] text-slate-300 space-y-1 max-h-24 overflow-y-auto">
            <div v-for="(log, idx) in interactiveLog" :key="idx" class="truncate">
              <span class="text-cyan-400 opacity-70">#</span> {{ log }}
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT 5 COLS: KERNEL REGISTRIES & STATE INSPECTION -->
      <div class="lg:col-span-5 bg-[#141A22] border border-[#21262d] rounded-2xl p-5 space-y-4 shadow-xl flex flex-col justify-between">
        <div class="space-y-4">
          <div class="border-b border-[#21262d] pb-3">
            <h4 class="text-sm font-bold text-white font-sans flex items-center gap-2">
              <Cpu class="w-4 h-4 text-cyan-400" />
              <span>Kernel Data Structures Active</span>
            </h4>
            <div class="text-xs text-slate-400 mt-0.5">Parameters instantiated in kernel memory</div>
          </div>

          <!-- Detailed Key-Value Metrics -->
          <div class="space-y-2.5">
            <div 
              v-for="(dt, idx) in currentStep.details" 
              :key="idx"
              class="p-3 rounded-xl bg-[#0D1117] border border-[#21262d] font-mono text-xs flex items-center justify-between shadow-inner"
            >
              <span class="text-slate-400 font-medium">{{ dt.label }}:</span>
              <div class="flex items-center gap-2 text-right">
                <span v-if="dt.live" class="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                  {{ dt.live }}
                </span>
                <span class="text-cyan-300 font-bold truncate max-w-[160px]">{{ dt.value }}</span>
              </div>
            </div>
          </div>

          <!-- Kernel State Summary -->
          <div class="p-3.5 rounded-xl bg-indigo-950/25 border border-indigo-500/30 font-mono text-xs space-y-1.5">
            <div class="text-[10px] text-indigo-400 uppercase font-bold flex items-center gap-1.5">
              <span>Resulting Kernel State:</span>
            </div>
            <div class="text-indigo-200 text-[11px] leading-relaxed">{{ currentStep.kernelState }}</div>
          </div>

          <!-- Educational Takeaway Card -->
          <div class="p-3.5 rounded-xl bg-[#0D1117] border border-[#21262d] text-xs font-sans text-slate-300 space-y-2">
            <div class="font-bold text-white flex items-center gap-2 text-xs">
              <span>Why Phase 0{{ currentStep.step }} Matters</span>
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-if="currentStep.step === 1">
              Docker images are immutable tarball layers. By hashing content, layers can be downloaded once and reused across hundreds of containers.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else-if="currentStep.step === 2">
              Instead of copying the full OS (which would take gigabytes and seconds), OverlayFS mounts in sub-milliseconds without duplicating disk bytes.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else-if="currentStep.step === 3">
              Without namespaces, the process would share the host PID table and could see or kill host processes.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else-if="currentStep.step === 4">
              cgroups prevent a runaway infinite loop or memory leak from freezing the physical server or starving neighbor containers.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else-if="currentStep.step === 5">
              The veth pair acts like an invisible Ethernet patch cable strung between the host bridge and container virtual NIC.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else-if="currentStep.step === 6">
              Unlike chroot (which a root user can break out of via fchdir), pivot_root completely swaps the mount namespace root, trapping the container.
            </div>
            <div class="text-[11px] text-slate-400 leading-relaxed" v-else>
              runc disappears cleanly! The application binary becomes PID 1 without leaving intermediary wrapper processes running.
            </div>
          </div>
        </div>

        <!-- Footer Navigation inside Right Column -->
        <div class="pt-4 border-t border-[#21262d] flex items-center justify-between">
          <button 
            @click="prevStep"
            :disabled="currentStepIndex === 0"
            class="px-3.5 py-2 rounded-xl bg-[#0D1117] hover:bg-[#202735] text-slate-300 disabled:opacity-40 text-xs font-mono cursor-pointer border border-[#2D3848]"
          >
            ← Previous Phase
          </button>
          <button 
            @click="nextStep"
            class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold cursor-pointer shadow-md flex items-center gap-1.5"
          >
            <span>{{ currentStepIndex === steps.length - 1 ? 'Restart Lifecycle' : 'Next Phase' }}</span>
            <ArrowRight class="w-3.5 h-3.5" />
          </button>
        </div>

      </div>
    </div>
  </div>
</template>