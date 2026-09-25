<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  Shield, 
  ShieldAlert, 
  ShieldCheck, 
  Lock, 
  Unlock, 
  Terminal, 
  Check, 
  X, 
  RefreshCw,
  Activity
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
const kernelAuditLogs = ref<string[]>([
  '[seccomp] initialized default BPF filter with 304 blacklisted syscalls',
  '[capset] dropped 22 out of 30 raw kernel capabilities for container security profile',
  '[audit] container initialized in isolated mount and network namespaces'
]);

function pushAuditLog(msg: string) {
  kernelAuditLogs.value.unshift(`[${new Date().toLocaleTimeString()}] ${msg}`);
  if (kernelAuditLogs.value.length > 8) kernelAuditLogs.value.pop();
}

function runSyscallTest(test: SyscallTest) {
  if (isPrivileged.value) {
    testResults.value[test.id] = {
      allowed: true,
      message: `SUCCESS (0): --privileged active. Kernel bypassed all capability checks and Seccomp filters!`,
      exitCode: 0,
    };
    pushAuditLog(`ALLOW (${test.syscall}): --privileged bypass active.`);
    return;
  }

  if (seccompMode.value === 'strict' && (test.id === 'mount_disk' || test.id === 'flush_iptables' || test.id === 'read_host_mem')) {
    testResults.value[test.id] = {
      allowed: false,
      message: `BLOCKED (159): Seccomp filter returned SECCOMP_RET_KILL_PROCESS on syscall ${test.syscall}`,
      exitCode: 159,
    };
    pushAuditLog(`KILL (${test.syscall}): Strict seccomp filter terminated test syscall.`);
    return;
  }

  const cap = capabilities.value.find(c => c.name === test.requiredCap);
  if (cap && cap.enabled) {
    testResults.value[test.id] = {
      allowed: true,
      message: `SUCCESS (0): Syscall ${test.syscall} allowed by granted capability ${test.requiredCap}.`,
      exitCode: 0,
    };
    pushAuditLog(`ALLOW (${test.syscall}): Validated capability token ${test.requiredCap}.`);
  } else {
    testResults.value[test.id] = {
      allowed: false,
      message: `DENIED (EPERM / 1): Operation not permitted. Requires ${test.requiredCap}, which is dropped by Docker's default security profile.`,
      exitCode: 1,
    };
    pushAuditLog(`DENY (${test.syscall}): Missing required capability token ${test.requiredCap}. EPERM.`);
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
    pushAuditLog('SECURITY WARNING: --privileged flag applied! All capabilities enabled & seccomp disabled.');
  } else {
    resetToDockerDefaults();
    pushAuditLog('Security profile restored to standard Docker hardening defaults.');
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

runAllTests();
</script>

<template>
  <div class="space-y-6 font-sans select-none">
    
    <!-- Modern Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 rounded-2xl bg-[#161B22] border border-[#21262d] shadow-xl">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <ShieldCheck class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-bold text-white font-sans flex items-center gap-2">
              <span>Linux Capabilities & Seccomp Syscall Sandbox</span>
              <span class="text-[10px] px-2 py-0.5 rounded font-mono bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
                Kernel Hardening
              </span>
            </h2>
            <p class="text-xs text-slate-400">
              Simulate how Docker uses Linux Capabilities (<code class="text-slate-300">capset</code>) and Seccomp BPF filters to restrict container root.
            </p>
          </div>
        </div>
      </div>

      <!-- Restored Refresh Icon in Reset Button -->
      <div class="flex items-center gap-2">
        <button 
          @click="resetToDockerDefaults"
          class="px-3.5 py-2 rounded-xl bg-[#0D1117] hover:bg-[#202735] text-slate-300 hover:text-white border border-[#2D3848] text-xs font-mono flex items-center gap-1.5 transition-all cursor-pointer shadow-sm"
        >
          <RefreshCw class="w-3.5 h-3.5 text-cyan-400" />
          Reset Defaults
        </button>
      </div>
    </div>

    <!-- PRIVILEGED MODE BANNER -->
    <div 
      class="p-4 rounded-2xl border transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-xl"
      :class="isPrivileged 
        ? 'bg-rose-950/30 border-rose-500/60 shadow-[0_0_25px_rgba(244,63,94,0.15)] ring-1 ring-rose-500/40' 
        : 'bg-[#141A22] border-[#21262d]'"
    >
      <div class="flex items-start gap-3.5">
        <div 
          class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border shadow-inner"
          :class="isPrivileged ? 'bg-rose-600 text-white border-rose-400 animate-pulse' : 'bg-emerald-950/80 text-emerald-400 border-emerald-800'"
        >
          <component :is="isPrivileged ? ShieldAlert : ShieldCheck" class="w-5 h-5" />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-bold text-white font-sans">
              Execution Mode: 
              <span :class="isPrivileged ? 'text-rose-400 font-mono font-bold' : 'text-emerald-400 font-mono font-bold'">
                {{ isPrivileged ? 'docker run --privileged (UNCONFINED ROOT)' : 'docker run (Default Restricted Root)' }}
              </span>
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-1 font-sans leading-relaxed">
            {{ isPrivileged 
              ? 'DANGER: Container has all 30+ Linux capabilities enabled, AppArmor disabled, and Seccomp filters turned off. Container root can execute arbitrary host syscalls!' 
              : 'Secure: Docker drops sensitive capabilities like CAP_SYS_ADMIN, CAP_NET_ADMIN, and applies 300+ blocked syscalls in the default seccomp profile.' }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2 self-start sm:self-auto shrink-0">
        <button 
          @click="togglePrivileged(!isPrivileged)"
          class="px-4 py-2.5 rounded-xl text-xs font-mono font-bold transition-all cursor-pointer flex items-center gap-2 shadow-md"
          :class="isPrivileged 
            ? 'bg-rose-600 hover:bg-rose-500 text-white ring-2 ring-rose-400/30' 
            : 'bg-[#0D1117] hover:bg-[#202735] text-slate-200 border border-[#2D3848]'"
        >
          <component :is="isPrivileged ? Unlock : Lock" class="w-3.5 h-3.5" />
          <span>{{ isPrivileged ? 'Disable --privileged' : 'Enable --privileged' }}</span>
        </button>
      </div>
    </div>

    <!-- MAIN TWO-COLUMN WORKBENCH -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      
      <!-- LEFT 7 COLS: CAPABILITIES MATRIX -->
      <div class="lg:col-span-7 bg-[#141A22] border border-[#21262d] rounded-2xl p-4 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-[#21262d] pb-3">
          <div class="flex items-center gap-2">
            <Shield class="w-4 h-4 text-cyan-400" />
            <h3 class="text-sm font-bold text-white font-sans">Linux Capabilities Matrix (`capset`)</h3>
          </div>
          <span class="text-xs font-mono px-2.5 py-0.5 rounded bg-[#0D1117] border border-[#21262d] text-cyan-300">
            {{ capabilities.filter(c => c.enabled).length }} / {{ capabilities.length }} Active
          </span>
        </div>

        <p class="text-xs text-slate-400 font-sans leading-relaxed">
          Docker breaks down traditional Linux root privileges into independent capability tokens. Toggle individual capabilities to test kernel access:
        </p>

        <div class="space-y-2.5">
          <div 
            v-for="cap in capabilities" 
            :key="cap.name"
            @click="!isPrivileged && (cap.enabled = !cap.enabled); runAllTests()"
            class="p-3 rounded-xl border transition-all cursor-pointer flex items-center justify-between shadow-sm"
            :class="cap.enabled 
              ? cap.name === 'CAP_SYS_ADMIN' || cap.name === 'CAP_NET_ADMIN' || cap.name === 'CAP_SYS_PTRACE'
                ? 'bg-rose-950/20 border-rose-500/50 text-white ring-1 ring-rose-500/20'
                : 'bg-blue-600/15 border-blue-500/50 text-white ring-1 ring-blue-500/20' 
              : 'bg-[#0D1117] border-[#21262d] text-slate-400 opacity-60 hover:opacity-100'"
          >
            <div class="space-y-1 truncate pr-2">
              <div class="flex items-center gap-2">
                <span class="font-mono font-bold text-xs" :class="cap.enabled ? 'text-cyan-300' : 'text-slate-400'">
                  {{ cap.name }}
                </span>
                <span 
                  class="text-[9px] px-2 py-0.5 rounded font-mono uppercase tracking-wider"
                  :class="cap.defaultInDocker ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30' : 'bg-slate-800 text-slate-400 border border-slate-700'"
                >
                  {{ cap.defaultInDocker ? 'Default' : 'Dropped' }}
                </span>
              </div>
              <div class="text-[11px] text-slate-300 font-sans truncate">
                {{ cap.description }}
              </div>
            </div>

            <div 
              class="w-5 h-5 rounded-lg flex items-center justify-center shrink-0 border transition-all shadow-inner"
              :class="cap.enabled 
                ? 'bg-blue-600 border-blue-500 text-white shadow-blue-500/30' 
                : 'border-slate-700 bg-[#0A0D12] text-transparent'"
            >
              <Check class="w-3 h-3" />
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT 5 COLS: INTERACTIVE SYSCALL EXPLOIT SIMULATOR -->
      <div class="lg:col-span-5 bg-[#141A22] border border-[#21262d] rounded-2xl p-4 shadow-xl flex flex-col justify-between space-y-4">
        <div>
          <div class="flex items-center justify-between border-b border-[#21262d] pb-3 mb-3">
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

          <p class="text-xs text-slate-400 mb-3 font-sans leading-relaxed">
            Test kernel enforcement against simulated container breakout commands:
          </p>

          <div class="space-y-2.5">
            <div 
              v-for="t in tests" 
              :key="t.id"
              class="p-3 rounded-xl bg-[#0D1117] border border-[#21262d] space-y-2 shadow-inner"
            >
              <div class="flex items-center justify-between gap-2">
                <code class="text-xs text-cyan-300 font-mono font-bold truncate"># {{ t.command }}</code>
                <button 
                  @click="runSyscallTest(t)"
                  class="px-2.5 py-1 rounded-lg text-[10px] bg-[#161B22] hover:bg-[#202735] text-slate-300 border border-[#2D3848] font-mono cursor-pointer shrink-0 shadow-sm"
                >
                  Test
                </button>
              </div>

              <div 
                v-if="testResults[t.id]" 
                class="p-2.5 rounded-lg text-[11px] font-mono leading-relaxed border"
                :class="testResults[t.id].allowed 
                  ? 'bg-emerald-950/20 text-emerald-300 border-emerald-500/40' 
                  : 'bg-rose-950/20 text-rose-300 border-rose-500/40'"
              >
                <div class="flex items-center gap-1.5 font-bold mb-1">
                  <component :is="testResults[t.id].allowed ? Check : X" class="w-3.5 h-3.5" />
                  <span>{{ testResults[t.id].allowed ? 'SYSCALL ALLOWED (0)' : 'PERMISSION DENIED (EPERM)' }}</span>
                </div>
                <div class="text-[10px] opacity-90">{{ testResults[t.id].message }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Seccomp BPF Profile Selector -->
        <div class="pt-4 border-t border-[#21262d] space-y-2.5">
          <div class="text-[10px] text-slate-400 uppercase font-mono font-bold">Seccomp BPF System Profile:</div>
          <div class="grid grid-cols-3 gap-2 font-mono text-[11px]">
            <button 
              @click="seccompMode = 'default'; runAllTests()"
              class="p-2 rounded-xl border transition-all cursor-pointer text-center font-bold"
              :class="seccompMode === 'default' ? 'bg-blue-600 text-white border-blue-500 shadow-md' : 'bg-[#0D1117] text-slate-400 border-[#21262d]'"
            >
              Default (300+ Blocked)
            </button>
            <button 
              @click="seccompMode = 'strict'; runAllTests()"
              class="p-2 rounded-xl border transition-all cursor-pointer text-center font-bold"
              :class="seccompMode === 'strict' ? 'bg-amber-600 text-white border-amber-500 shadow-md' : 'bg-[#0D1117] text-slate-400 border-[#21262d]'"
            >
              Strict Whitelist
            </button>
            <button 
              @click="seccompMode = 'unconfined'; runAllTests()"
              class="p-2 rounded-xl border transition-all cursor-pointer text-center font-bold"
              :class="seccompMode === 'unconfined' ? 'bg-rose-600 text-white border-rose-500 shadow-md' : 'bg-[#0D1117] text-slate-400 border-[#21262d]'"
            >
              Unconfined
            </button>
          </div>
        </div>

      </div>
    </div>

    <!-- Live Kernel Audit Stream -->
    <div class="bg-[#141A22] border border-[#21262d] rounded-2xl p-4 shadow-xl space-y-2">
      <div class="flex items-center justify-between text-[11px] font-mono text-slate-400 uppercase font-bold">
        <span class="flex items-center gap-1.5">
          <Activity class="w-3.5 h-3.5 text-cyan-400" />
          <span>Kernel Audit Log Stream (auditd / seccomp):</span>
        </span>
        <span class="text-cyan-400">● Active</span>
      </div>
      <div class="p-3 rounded-xl bg-[#090D14] border border-[#1D2430] font-mono text-[11px] text-slate-300 space-y-1 max-h-28 overflow-y-auto">
        <div v-for="(log, idx) in kernelAuditLogs" :key="idx" class="truncate">
          <span class="text-cyan-400 opacity-70">#</span> {{ log }}
        </div>
      </div>
    </div>

  </div>
</template>