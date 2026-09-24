export type ContainerStatus = 'created' | 'running' | 'paused' | 'stopped' | 'oom_killed';

export interface LogEntry {
  timestamp: string;
  stream: 'stdout' | 'stderr' | 'system';
  message: string;
}

export interface ProcessInfo {
  hostPid: number;
  containerPid: number;
  user: string;
  cpu: number;
  memoryMb: number;
  command: string;
}

export interface NamespaceInfo {
  pidInode: string;
  utsInode: string;
  mntInode: string;
  netInode: string;
  ipcInode: string;
  userInode: string;
  cgroupInode: string;
  hostname: string;
  virtualIp: string;
  vethPair: string;
}

export interface CgroupLimits {
  cgroupPath: string;
  memoryMaxBytes: number;
  memoryUsageBytes: number;
  memoryMaxUsage: number;
  cpuQuotaUs: number;
  cpuPeriodUs: number;
  cpuShares: number;
  cpuPercent: number;
  throttlePeriods: number;
  throttledTimeUs: number;
  oomKillEvents: number;
  oomScoreAdj: number;
}

export interface OverlayFileDiff {
  path: string;
  action: string;
  sizeBytes: number;
  timestamp: string;
}

export interface RootfsOverlay {
  baseImage: string;
  lowerDir: string;
  upperDir: string;
  workDir: string;
  mergedDir: string;
  modifiedFiles: OverlayFileDiff[];
}

export interface MetricPoint {
  timestamp: string;
  cpuPercent: number;
  cpuQuotaPercent: number;
  memoryUsageBytes: number;
  memoryMaxBytes: number;
  isThrottled: boolean;
  throttlePeriods: number;
  throttledTimeUs: number;
  oomKillEvents: number;
}

export interface Container {
  id: string;
  name: string;
  image: string;
  command: string;
  status: ContainerStatus;
  createdAt: string;
  startedAt?: string;
  hostPid: number;
  containerPid: number;
  portBindings: string[];
  env: string[];
  namespaces: NamespaceInfo;
  cgroups: CgroupLimits;
  rootfs: RootfsOverlay;
  processes: ProcessInfo[];
  logs: LogEntry[];
  metricsHistory?: MetricPoint[];
}

export interface CreateContainerPayload {
  name: string;
  image: string;
  command: string;
  memoryLimitMb: number;
  cpuQuotaCores: number;
  enablePidNs: boolean;
  enableUtsNs: boolean;
  enableNetNs: boolean;
  enableMntNs: boolean;
  hostname?: string;
  portBindings?: string[];
  env?: string[];
}

export interface KernelPrimitiveGuide {
  name: string;
  linuxFlag: string;
  syscallFile: string;
  goCode: string;
  explanation: string;
}

export interface CliCommandResponse {
  output: string;
  error?: string;
  exitCode: number;
}
