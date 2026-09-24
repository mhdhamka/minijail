<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Play, 
  Pause, 
  RotateCcw, 
  CheckCircle2, 
  ArrowRight, 
  Terminal, 
  Cpu, 
  Network, 
  FolderTree, 
  ShieldCheck, 
  Sliders, 
  Box, 
  Server,
  Zap,
  Clock,
  Sparkles
} from 'lucide-vue-next';

interface LifecycleStep {
  step: number;
  title: string;
  subsystem: string;
  color: string;
  dockerStage: string;
  syscall: string;
  description: string;
  kernelState: string;
  details: { label: string; value: string }[];
}

const steps: LifecycleStep[] = [
  {
    step: 1,
    title: 'OCI Image Resolution & Layer Unpack',
    subsystem: 'Storage Driver (OverlayFS)',
    color: '#38bdf8',
    dockerStage: 'docker run -d --name web -p 8080:80 alpine:3.19',
    syscall: 'openat() + tar unpack + sha256 checksum verification',
    description: 'Docker Engine checks local content-addressable storage cache. Unpacks immutable lower layer tarballs into /var/lib/minijail/images/alpine:3.19/rootfs.',
    kernelState: 'Image blobs verified; lowerdir layers mapped in filesystem table.',
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
    dockerStage: 'Preparing container rootfs with copy-on-write diff',
    syscall: 'mount("overlay", "/merged", "overlay", MS_NODEV, "lowerdir=...,upperdir=...,workdir=...")',
    description: 'Creates an ephemeral upperdir diff directory and workdir. Issues kernel mount syscall to overlay the read-only image layers under a merged directory.',
    kernelState: 'New super_block registered; VFS resolves file lookups through upper -> lower chain.',
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
    dockerStage: 'Spawning container runner process (containerd-shim)',
    syscall: 'clone(SIGCHLD | CLONE_NEWPID | CLONE_NEWNET | CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWIPC, ...)',
    description: 'Kernel allocates a brand new nsproxy struct with fresh namespace pointers. The child process receives its own virtual PID 1, network stack, and mount table.',
    kernelState: 'Allocated new task_struct; nsproxy->pid_ns pointing to new pid namespace (Inode 4026535480).',
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
    dockerStage: 'Applying CPU quota and memory bounds',
    syscall: 'mkdir(/sys/fs/cgroup/minijail/c7a8) && write(cgroup.procs, pid)',
    description: 'Runc writes memory.max and cpu.max parameters. Process host PID is added to cgroup.procs to enforce Completely Fair Scheduler (CFS) quotas and OOM constraints.',
    kernelState: 'cgroup attached to task_struct; CFS bandwidth timer initialized.',
    details: [
      { label: 'cpu.max', value: '50000 100000 (0.50 cores)' },
      { label: 'memory.max', value: '67108864 (64 MB ceiling)' },
      { label: 'cgroup.procs', value: 'Host PID 42193 attached' },
    ],
  },
  {
    step: 5,
    title: 'Virtual Ethernet (veth) Wiring & Port Forwarding',
    subsystem: 'Network Core & iptables NAT',
    color: '#06b6d4',
    dockerStage: 'Configuring bridge docker0 and veth pair',
    syscall: 'ip link add veth_c7a8 type veth peer name eth0 netns <pid>',
    description: 'Docker creates a virtual Ethernet pair. Moves one end into the container network namespace (renamed to eth0) and attaches peer veth_c7a8 to bridge docker0 (172.18.0.1).',
    kernelState: 'veth interfaces bound; netfilter NAT PREROUTING table updated for 8080:80.',
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
    dockerStage: 'Locking process inside container root directory',
    syscall: 'pivot_root("/merged", "/merged/.oldroot") + umount2("/.oldroot", MNT_DETACH)',
    description: 'Swaps current process rootfs with the OverlayFS merged mount. Unmounts the host oldroot and frees references so container process cannot traverse back up.',
    kernelState: 'current->fs->root updated to merged root; host filesystem completely unreachable.',
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
    dockerStage: 'Container entrypoint running as PID 1',
    syscall: 'execve("/bin/sh", ["sh", "-c", "while true; do ..."], envp)',
    description: 'The bootstrap launcher runc overwrites its own memory space with the user-specified entrypoint binary. The application takes over as PID 1.',
    kernelState: 'ELF binary loaded; process starts execution loop in user mode as container PID 1.',
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

const currentStep = computed(() => steps[currentStepIndex.value]);

function nextStep() {
  if (currentStepIndex.value < steps.length - 1) {
    currentStepIndex.value++;
  } else {
    currentStepIndex.value = 0;
  }
}

function prevStep() {
  if (currentStepIndex.value > 0) {
    currentStepIndex.value--;
  }
}

function jumpToStep(idx: number) {
  currentStepIndex.value = idx;
}

function togglePlayback() {
  if (isPlaying.value) {
    clearInterval(playbackTimer);
    isPlaying.value = false;
  } else {
    isPlaying.value = true;
    playbackTimer = setInterval(() => {
      if (currentStepIndex.value < steps.length - 1) {
        currentStepIndex.value++;
      } else {
        currentStepIndex.value = 0;
      }
    }, 2800);
  }
}

function resetPlayback() {
  clearInterval(playbackTimer);
  isPlaying.value = false;
  currentStepIndex.value = 0;
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <Zap class="w-5 h-5 text-cyan-400" />
          <h2 class="text-xl font-bold text-white font-sans">
            Container Lifecycle Machine: From `docker run` to Executed Process
          </h2>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Scrub through the millisecond-by-millisecond sequence of Linux kernel syscalls that transform a standard executable into an isolated container.
        </p>
      </div>

      <!-- Playback Controls -->
      <div class="flex items-center gap-2">
        <button 
          @click="prevStep"
          :disabled="currentStepIndex === 0"
          class="px-2.5 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 disabled:opacity-40 border border-[#2D3848] text-xs font-mono transition-colors cursor-pointer"
        >
          Previous
        </button>
        <button 
          @click="togglePlayback"
          class="px-3 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-colors cursor-pointer text-white shadow-sm"
          :class="isPlaying ? 'bg-amber-600 hover:bg-amber-500' : 'bg-[#1D63ED] hover:bg-[#1A57D0]'"
        >
          <component :is="isPlaying ? Pause : Play" class="w-3.5 h-3.5" />
          <span>{{ isPlaying ? 'Pause Simulation' : 'Auto Play' }}</span>
        </button>
        <button 
          @click="nextStep"
          class="px-2.5 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 border border-[#2D3848] text-xs font-mono transition-colors cursor-pointer"
        >
          Next Step
        </button>
        <button 
          @click="resetPlayback"
          class="p-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-400 hover:text-white border border-[#2D3848] transition-colors cursor-pointer"
          title="Reset to Step 1"
        >
          <RotateCcw class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Stepper Navigation Bar -->
    <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2">
      <button 
        v-for="(st, idx) in steps" 
        :key="st.step"
        @click="jumpToStep(idx)"
        class="p-2.5 rounded-xl border text-left transition-all cursor-pointer relative overflow-hidden"
        :class="idx === currentStepIndex 
          ? 'bg-[#161F2E] border-[#1D63ED] ring-1 ring-[#1D63ED]/50 shadow-md' 
          : idx < currentStepIndex 
          ? 'bg-[#10151C] border-[#253040] text-slate-300' 
          : 'bg-[#0E1217] border-[#1E2530] text-slate-500'"
      >
        <div class="flex items-center justify-between text-[11px] font-mono">
          <span class="font-bold" :style="{ color: idx === currentStepIndex ? st.color : undefined }">
            Phase {{ st.step }}
          </span>
          <CheckCircle2 v-if="idx < currentStepIndex" class="w-3 h-3 text-emerald-400" />
        </div>
        <div class="text-[11px] font-sans font-semibold mt-1 truncate" :class="idx === currentStepIndex ? 'text-white' : 'text-slate-400'">
          {{ st.subsystem }}
        </div>
        <!-- Progress bar highlight under active item -->
        <div 
          v-if="idx === currentStepIndex" 
          class="absolute bottom-0 left-0 right-0 h-0.5" 
          :style="{ backgroundColor: st.color }"
        ></div>
      </button>
    </div>

    <!-- MAIN ACTIVE STEP WORKBENCH -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: DETAILED KERNEL EXECUTION SCHEMATIC -->
      <div class="lg:col-span-7 bg-[#141A22] border border-[#232A35] rounded-xl p-5 space-y-4">
        
        <!-- Step Banner -->
        <div class="flex items-start justify-between border-b border-[#232A35] pb-3">
          <div>
            <div class="flex items-center gap-2">
              <span 
                class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold font-mono text-black"
                :style="{ backgroundColor: currentStep.color }"
              >
                {{ currentStep.step }}
              </span>
              <h3 class="text-base font-bold text-white font-sans">{{ currentStep.title }}</h3>
            </div>
            <div class="text-xs font-mono mt-1" :style="{ color: currentStep.color }">
              Subsystem: {{ currentStep.subsystem }}
            </div>
          </div>
          <span class="text-xs px-2.5 py-1 rounded-lg bg-[#0E1217] border border-[#2B3545] font-mono text-slate-400">
            Step {{ currentStep.step }} of {{ steps.length }}
          </span>
        </div>

        <!-- Description text -->
        <p class="text-slate-300 text-xs font-sans leading-relaxed">
          {{ currentStep.description }}
        </p>

        <!-- Docker Command Context -->
        <div class="p-3 rounded-lg bg-[#0E1217] border border-[#1E2633] space-y-1 font-mono text-xs">
          <div class="text-[10px] text-slate-500 uppercase font-bold flex items-center gap-1.5">
            <Terminal class="w-3 h-3 text-[#0db7ed]" />
            <span>High-Level Docker CLI Context:</span>
          </div>
          <div class="text-cyan-300 font-bold">{{ currentStep.dockerStage }}</div>
        </div>

        <!-- Kernel Syscall Code Window -->
        <div class="rounded-lg border border-[#232A35] bg-[#0A0D12] overflow-hidden">
          <div class="px-3 py-1.5 bg-[#10141B] border-b border-[#232A35] flex items-center justify-between text-[11px] font-mono">
            <span class="text-slate-400">Actual Linux Kernel Syscall:</span>
            <span class="text-emerald-400 font-bold">C / Go Runtime Call</span>
          </div>
          <div class="p-3 text-emerald-300 font-mono text-xs overflow-x-auto">
            <code>{{ currentStep.syscall }}</code>
          </div>
        </div>

        <!-- Kernel Task State Result -->
        <div class="p-3 rounded-lg bg-indigo-950/20 border border-indigo-800/40 text-xs font-mono">
          <div class="text-[10px] text-indigo-400 uppercase font-bold mb-1">Resulting Linux Kernel State:</div>
          <div class="text-indigo-200">{{ currentStep.kernelState }}</div>
        </div>
      </div>

      <!-- RIGHT 5 COLS: KERNEL DATA STRUCTURES & INSPECTION -->
      <div class="lg:col-span-5 bg-[#141A22] border border-[#232A35] rounded-xl p-5 space-y-4 flex flex-col justify-between">
        <div class="space-y-4">
          <div class="border-b border-[#232A35] pb-2">
            <h4 class="text-sm font-bold text-white font-sans flex items-center gap-2">
              <Cpu class="w-4 h-4 text-cyan-400" />
              <span>Kernel Data Structures Active</span>
            </h4>
            <div class="text-xs text-slate-400 mt-0.5">Parameters instantiated in kernel memory</div>
          </div>

          <div class="space-y-2.5">
            <div 
              v-for="(dt, idx) in currentStep.details" 
              :key="idx"
              class="p-2.5 rounded-lg bg-[#0E1217] border border-[#232A35] font-mono text-xs flex items-center justify-between"
            >
              <span class="text-slate-400">{{ dt.label }}:</span>
              <span class="text-cyan-300 font-bold truncate ml-2">{{ dt.value }}</span>
            </div>
          </div>

          <!-- Educational Takeaway for this step -->
          <div class="p-3 rounded-lg bg-[#0E1217] border border-[#232A35] text-xs font-sans text-slate-300 space-y-1.5">
            <div class="font-bold text-white flex items-center gap-1.5 text-xs">
              <Sparkles class="w-3.5 h-3.5 text-amber-400" />
              <span>Why This Step Matters</span>
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

        <!-- Navigation Buttons -->
        <div class="pt-3 border-t border-[#232A35] flex items-center justify-between">
          <button 
            @click="prevStep"
            :disabled="currentStepIndex === 0"
            class="px-3 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 disabled:opacity-40 text-xs font-mono cursor-pointer"
          >
            ← Back
          </button>
          <button 
            @click="nextStep"
            class="px-3 py-1.5 rounded-lg bg-[#1D63ED] hover:bg-[#1A57D0] text-white text-xs font-semibold cursor-pointer shadow-sm"
          >
            {{ currentStepIndex === steps.length - 1 ? 'Restart from Phase 1' : 'Next: ' + steps[(currentStepIndex + 1) % steps.length].subsystem + ' →' }}
          </button>
        </div>

      </div>
    </div>
  </div>
</template>
