<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Shield, 
  ShieldAlert, 
  ShieldCheck, 
  Lock, 
  Unlock, 
  AlertTriangle, 
  Terminal, 
  Check, 
  X, 
  FileCode2, 
  Sparkles,
  Info
} from 'lucide-vue-next';

interface LinuxCapability {
  name: string;
  description: string;
  defaultInDocker: boolean;
  enabled: boolean;
  securityRisk: string;
}

const isPrivileged = ref(false);

const capabilities = ref<LinuxCapability[]>([
  {
    name: 'CAP_CHOWN',
    description: 'Make arbitrary changes to file UIDs and GIDs',
    defaultInDocker: true,
    enabled: true,
    securityRisk: 'Low. Needed for package managers and setup scripts.',
  },
  {
    name: 'CAP_NET_BIND_SERVICE',
    description: 'Bind a socket to Internet domain privileged ports (port numbers less than 1024)',
    defaultInDocker: true,
    enabled: true,
    securityRisk: 'Low. Allows binding port 80/443 inside container NET namespace.',
  },
  {
    name: 'CAP_SETUID / CAP_SETGID',
    description: 'Make arbitrary manipulations of process UIDs and GIDs (e.g. su, sudo, setuid)',
    defaultInDocker: true,
    enabled: true,
    securityRisk: 'Medium. Standard in multi-user containers.',
  },
  {
    name: 'CAP_NET_RAW',
    description: 'Use RAW and PACKET sockets (e.g. ping, traceroute, raw packet crafting)',
    defaultInDocker: true,
    enabled: true,
    securityRisk: 'Medium. Allows sending arbitrary raw ethernet frames on veth.',
  },
  {
    name: 'CAP_SYS_ADMIN',
    description: 'The "god mode" capability. Perform mount, unshare, bpf, pivot_root, quota, swapon/swapoff',
    defaultInDocker: false,
    enabled: false,
    securityRisk: 'CRITICAL. Grants container root ability to mount host disks or break out of cgroups/namespaces.',
  },
  {
    name: 'CAP_NET_ADMIN',
    description: 'Perform network administration: modify iptables, route tables, interface IPs',
    defaultInDocker: false,
    enabled: false,
    securityRisk: 'HIGH. Allows rewriting container and veth bridge routing tables.',
  },
  {
    name: 'CAP_SYS_PTRACE',
    description: 'Trace arbitrary processes using ptrace(2) (gdb, strace, reading process memory)',
    defaultInDocker: false,
    enabled: false,
    securityRisk: 'HIGH. If PID namespace is shared, allows injecting code into host processes.',
  },
  {
    name: 'CAP_MKNOD',
    description: 'Create special files using mknod(2) (e.g. block devices like /dev/sda)',
    defaultInDocker: false,
    enabled: false,
    securityRisk: 'HIGH. Could create raw device nodes to read raw disk blocks.',
  },
]);

// Seccomp profiles
const seccompMode = ref<'default' | 'unconfined' | 'strict'>('default');

// Interactive Attack / Syscall Execution Simulator
interface SyscallTest {
  id: string;
  command: string;
  syscall: string;
  requiredCap: string;
  requiresPrivileged?: boolean;
}

const tests: SyscallTest[] = [
  {
    id: 'mount_disk',
    command: 'mount -t ext4 /dev/sda1 /mnt/host_root',
    syscall: 'mount(2)',
    requiredCap: 'CAP_SYS_ADMIN',
  },
  {
    id: 'flush_iptables',
    command: 'iptables -F && iptables -t nat -F',
    syscall: 'setsockopt(SO_SET_REPLACE)',
    requiredCap: 'CAP_NET_ADMIN',
  },
  {
    id: 'ping_gateway',
    command: 'ping -c 1 172.18.0.1',
    syscall: 'socket(AF_INET, SOCK_RAW, IPPROTO_ICMP)',
    requiredCap: 'CAP_NET_RAW',
  },
  {
    id: 'bind_port_80',
    command: 'nc -l -p 80',
    syscall: 'bind(2) to port < 1024',
    requiredCap: 'CAP_NET_BIND_SERVICE',
  },
  {
    id: 'read_host_mem',
    command: 'ptrace(PTRACE_ATTACH, 1, ...)',
    syscall: 'ptrace(2)',
    requiredCap: 'CAP_SYS_PTRACE',
  },
];

const testResults = ref<Record<string, { allowed: boolean; message: string; exitCode: number }>>({});

function runSyscallTest(test: SyscallTest) {
  // If privileged is on, EVERYTHING passes!
  if (isPrivileged.value) {
    testResults.value[test.id] = {
      allowed: true,
      message: `SUCCESS (0): --privileged active. Kernel bypassed all capability checks and Seccomp filters!`,
      exitCode: 0,
    };
    return;
  }

  // Check seccomp block
  if (seccompMode.value === 'strict' && (test.id === 'mount_disk' || test.id === 'flush_iptables' || test.id === 'read_host_mem')) {
    testResults.value[test.id] = {
      allowed: false,
      message: `BLOCKED (159): Seccomp filter returned SECCOMP_RET_KILL_PROCESS on syscall ${test.syscall}`,
      exitCode: 159,
    };
    return;
  }

  // Check capabilities
  const cap = capabilities.value.find(c => c.name === test.requiredCap);
  if (cap && cap.enabled) {
    testResults.value[test.id] = {
      allowed: true,
      message: `SUCCESS (0): Syscall ${test.syscall} allowed by granted capability ${test.requiredCap}.`,
      exitCode: 0,
    };
  } else {
    testResults.value[test.id] = {
      allowed: false,
      message: `DENIED (EPERM / 1): Operation not permitted. Requires ${test.requiredCap}, which is dropped by Docker's default security profile.`,
      exitCode: 1,
    };
  }
}

function runAllTests() {
  for (const t of tests) {
    runSyscallTest(t);
  }
}

function togglePrivileged(val: boolean) {
  isPrivileged.value = val;
  if (val) {
    for (const c of capabilities.value) {
      c.enabled = true;
    }
  } else {
    resetToDockerDefaults();
  }
  runAllTests();
}

function resetToDockerDefaults() {
  isPrivileged.value = false;
  seccompMode.value = 'default';
  for (const c of capabilities.value) {
    c.enabled = c.defaultInDocker;
  }
  runAllTests();
}

// Run initial tests
runAllTests();
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-[#232A35] pb-4">
      <div>
        <div class="flex items-center gap-2">
          <ShieldCheck class="w-5 h-5 text-emerald-400" />
          <h2 class="text-xl font-bold text-white font-sans">
            Linux Capabilities & Seccomp Syscall Sandbox
          </h2>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Simulate how Docker uses Linux Capabilities (`capset`) and Seccomp BPF filters to restrict what even a <code class="text-rose-400">root</code> user inside a container can execute.
        </p>
      </div>

      <!-- Reset & Status -->
      <div class="flex items-center gap-2">
        <button 
          @click="resetToDockerDefaults"
          class="px-3 py-1.5 rounded-lg bg-[#161B22] hover:bg-[#202735] text-slate-300 hover:text-white border border-[#2D3848] text-xs font-mono transition-colors cursor-pointer"
        >
          Reset Docker Defaults
        </button>
      </div>
    </div>

    <!-- PRIVILEGED MODE DANGER BANNER -->
    <div 
      class="p-4 rounded-xl border transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-4"
      :class="isPrivileged 
        ? 'bg-rose-950/40 border-rose-500 shadow-[0_0_20px_rgba(244,63,94,0.25)] ring-1 ring-rose-500/50' 
        : 'bg-[#141A22] border-[#232A35]'"
    >
      <div class="flex items-start gap-3">
        <div 
          class="w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border"
          :class="isPrivileged ? 'bg-rose-600 text-white border-rose-400 animate-pulse' : 'bg-emerald-950/60 text-emerald-400 border-emerald-800'"
        >
          <component :is="isPrivileged ? ShieldAlert : ShieldCheck" class="w-5 h-5" />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-bold text-white font-sans">
              Container Execution Mode: 
              <span :class="isPrivileged ? 'text-rose-400 font-mono' : 'text-emerald-400 font-mono'">
                {{ isPrivileged ? 'docker run --privileged (UNCONFINED ROOT)' : 'docker run (Default Restricted Root)' }}
              </span>
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-0.5 font-sans">
            {{ isPrivileged 
              ? 'DANGER: Container has all Linux capabilities enabled, AppArmor disabled, and Seccomp filters turned off. Container root can execute arbitrary host commands!' 
              : 'Secure: Docker drops sensitive capabilities like CAP_SYS_ADMIN, CAP_NET_ADMIN, and applies 300+ blocked syscalls in default seccomp profile.' }}
          </p>
        </div>
      </div>

      <!-- Toggle Button -->
      <div class="flex items-center gap-2 self-start sm:self-auto shrink-0">
        <button 
          @click="togglePrivileged(!isPrivileged)"
          class="px-3 py-2 rounded-lg text-xs font-mono font-bold transition-all cursor-pointer flex items-center gap-1.5 shadow-sm"
          :class="isPrivileged 
            ? 'bg-rose-600 hover:bg-rose-500 text-white' 
            : 'bg-[#161B22] hover:bg-[#202735] text-slate-300 border border-[#2D3848]'"
        >
          <component :is="isPrivileged ? Unlock : Lock" class="w-3.5 h-3.5" />
          <span>{{ isPrivileged ? 'Disable --privileged' : 'Enable --privileged' }}</span>
        </button>
      </div>
    </div>

    <!-- MAIN TWO-COLUMN WORKBENCH: CAPABILITIES LIST vs LIVE SYSCALL TESTER -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: LINUX CAPABILITIES TOGGLE MATRIX -->
      <div class="lg:col-span-7 bg-[#141A22] border border-[#232A35] rounded-xl p-4 space-y-4">
        <div class="flex items-center justify-between border-b border-[#232A35] pb-2">
          <div class="flex items-center gap-2">
            <Shield class="w-4 h-4 text-cyan-400" />
            <h3 class="text-sm font-bold text-white font-sans">Linux Capabilities Matrix (`capset`)</h3>
          </div>
          <span class="text-xs font-mono text-slate-400">
            {{ capabilities.filter(c => c.enabled).length }} / {{ capabilities.length }} active
          </span>
        </div>

        <p class="text-xs text-slate-400 font-sans">
          Docker segments Linux root power into fine-grained permissions. Toggle individual capabilities to test if sensitive operations succeed or fail:
        </p>

        <div class="space-y-2">
          <div 
            v-for="cap in capabilities" 
            :key="cap.name"
            @click="!isPrivileged && (cap.enabled = !cap.enabled); runAllTests()"
            class="p-2.5 rounded-lg border transition-all cursor-pointer flex items-center justify-between"
            :class="cap.enabled 
              ? cap.name === 'CAP_SYS_ADMIN' || cap.name === 'CAP_NET_ADMIN' || cap.name === 'CAP_SYS_PTRACE'
                ? 'bg-rose-950/20 border-rose-600/50 text-white'
                : 'bg-[#161E2C] border-[#1D63ED]/60 text-white' 
              : 'bg-[#0E1217] border-[#232A35] text-slate-500 opacity-60 hover:opacity-100'"
          >
            <div class="space-y-0.5 truncate pr-2">
              <div class="flex items-center gap-2">
                <span class="font-mono font-bold text-xs" :class="cap.enabled ? 'text-cyan-300' : 'text-slate-400'">
                  {{ cap.name }}
                </span>
                <span 
                  class="text-[9px] px-1 py-0.2 rounded font-mono"
                  :class="cap.defaultInDocker ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800' : 'bg-slate-800 text-slate-400'"
                >
                  {{ cap.defaultInDocker ? 'Docker Default' : 'Dropped by Default' }}
                </span>
              </div>
              <div class="text-[11px] text-slate-400 font-sans truncate">
                {{ cap.description }}
              </div>
            </div>

            <!-- Status Checkbox Indicator -->
            <div 
              class="w-5 h-5 rounded flex items-center justify-center shrink-0 border"
              :class="cap.enabled 
                ? 'bg-[#1D63ED] border-[#1D63ED] text-white' 
                : 'border-slate-600 bg-slate-800 text-transparent'"
            >
              <Check class="w-3 h-3" />
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT 5 COLS: INTERACTIVE SYSCALL EXPLOIT SIMULATOR -->
      <div class="lg:col-span-5 bg-[#141A22] border border-[#232A35] rounded-xl p-4 flex flex-col justify-between space-y-4">
        <div>
          <div class="flex items-center justify-between border-b border-[#232A35] pb-2 mb-3">
            <div class="flex items-center gap-2">
              <Terminal class="w-4 h-4 text-emerald-400" />
              <h3 class="text-sm font-bold text-white font-sans">Syscall Execution Sandbox</h3>
            </div>
            <button 
              @click="runAllTests"
              class="text-[11px] font-mono text-cyan-400 hover:underline cursor-pointer"
            >
              Re-run all
            </button>
          </div>

          <p class="text-xs text-slate-400 mb-3 font-sans">
            Simulate running sensitive commands inside the container terminal. See how the Linux kernel denies unauthorized syscalls:
          </p>

          <div class="space-y-3">
            <div 
              v-for="t in tests" 
              :key="t.id"
              class="p-2.5 rounded-lg bg-[#0E1217] border border-[#232A35] space-y-2"
            >
              <div class="flex items-center justify-between">
                <code class="text-xs text-cyan-300 font-mono font-bold truncate"># {{ t.command }}</code>
                <button 
                  @click="runSyscallTest(t)"
                  class="px-2 py-0.5 rounded text-[10px] bg-[#161B22] hover:bg-[#202735] text-slate-300 border border-[#2D3848] font-mono cursor-pointer shrink-0 ml-1"
                >
                  Test
                </button>
              </div>

              <!-- Result Badge -->
              <div 
                v-if="testResults[t.id]" 
                class="p-2 rounded text-[11px] font-mono leading-relaxed"
                :class="testResults[t.id].allowed 
                  ? 'bg-emerald-950/30 text-emerald-300 border border-emerald-800/60' 
                  : 'bg-rose-950/30 text-rose-300 border border-rose-800/60'"
              >
                <div class="flex items-center gap-1.5 font-bold mb-0.5">
                  <component :is="testResults[t.id].allowed ? Check : X" class="w-3.5 h-3.5" />
                  <span>{{ testResults[t.id].allowed ? 'SYSCALL ALLOWED (0)' : 'PERMISSION DENIED (EPERM)' }}</span>
                </div>
                <div class="text-[10px] opacity-90">{{ testResults[t.id].message }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Seccomp Profile Selector -->
        <div class="pt-3 border-t border-[#232A35] space-y-2">
          <div class="text-[10px] text-slate-400 uppercase font-mono font-bold">Seccomp BPF Profile:</div>
          <div class="grid grid-cols-3 gap-1.5 font-mono text-xs">
            <button 
              @click="seccompMode = 'default'; runAllTests()"
              class="p-1.5 rounded border transition-colors cursor-pointer text-center"
              :class="seccompMode === 'default' ? 'bg-[#1D63ED] text-white border-[#1D63ED]' : 'bg-[#0E1217] text-slate-400 border-[#232A35]'"
            >
              Default (300+ blocked)
            </button>
            <button 
              @click="seccompMode = 'strict'; runAllTests()"
              class="p-1.5 rounded border transition-colors cursor-pointer text-center"
              :class="seccompMode === 'strict' ? 'bg-amber-600 text-white border-amber-600' : 'bg-[#0E1217] text-slate-400 border-[#232A35]'"
            >
              Strict Whitelist
            </button>
            <button 
              @click="seccompMode = 'unconfined'; runAllTests()"
              class="p-1.5 rounded border transition-colors cursor-pointer text-center"
              :class="seccompMode === 'unconfined' ? 'bg-rose-600 text-white border-rose-600' : 'bg-[#0E1217] text-slate-400 border-[#232A35]'"
            >
              Unconfined
            </button>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>
