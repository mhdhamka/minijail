import type {
  Container,
  CreateContainerPayload,
  KernelPrimitiveGuide,
  CliCommandResponse,
} from './types/container';

const API_BASE = '/api';

async function handleResponse<T>(res: Response, fallbackMsg: string): Promise<T> {
  const contentType = res.headers.get('content-type') || '';
  if (!res.ok) {
    if (contentType.includes('application/json')) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `${fallbackMsg}: ${res.statusText}`);
    } else {
      throw new Error(`${fallbackMsg}: HTTP ${res.status} (${res.statusText})`);
    }
  }
  if (!contentType.includes('application/json')) {
    throw new Error(`${fallbackMsg}: expected JSON response from server`);
  }
  return res.json();
}

export async function fetchContainers(): Promise<Container[]> {
  const res = await fetch(`${API_BASE}/containers`);
  return handleResponse<Container[]>(res, 'Failed to fetch containers');
}

export async function fetchContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}`);
  return handleResponse<Container>(res, `Failed to fetch container ${id}`);
}

export async function createContainer(payload: CreateContainerPayload): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<Container>(res, 'Failed to create container');
}

export async function startContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/start`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to start container');
}

export async function stopContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/stop`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to stop container');
}

export async function pauseContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/pause`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to pause container');
}

export async function unpauseContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/unpause`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to unpause container');
}

export async function killContainer(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/kill`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to kill container');
}

export async function deleteContainer(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/containers/${id}`, { method: 'DELETE' });
  if (!res.ok) {
    const contentType = res.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Failed to delete container: ${res.statusText}`);
    }
    throw new Error(`Failed to delete container: HTTP ${res.status}`);
  }
}

export async function execCommand(id: string, command: string): Promise<{ exitCode: number; stdout: string; stderr: string }> {
  const res = await fetch(`${API_BASE}/containers/${id}/exec`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ command }),
  });
  return handleResponse<{ exitCode: number; stdout: string; stderr: string }>(res, 'Failed to exec command');
}

export async function stressMemory(id: string, deltaMb: number = 24): Promise<{ container: Container; oomKilled: boolean }> {
  const res = await fetch(`${API_BASE}/containers/${id}/stress-mem`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deltaMb }),
  });
  return handleResponse<{ container: Container; oomKilled: boolean }>(res, 'Failed to stress memory');
}

export async function stressCpu(id: string): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/stress-cpu`, { method: 'POST' });
  return handleResponse<Container>(res, 'Failed to stress CPU');
}

export async function executeCli(command: string): Promise<CliCommandResponse> {
  const res = await fetch(`${API_BASE}/cli`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ command }),
  });
  return handleResponse<CliCommandResponse>(res, 'CLI execution error');
}

export async function fetchKernelPrimitives(): Promise<KernelPrimitiveGuide[]> {
  const res = await fetch(`${API_BASE}/primitives`);
  return handleResponse<KernelPrimitiveGuide[]>(res, 'Failed to fetch primitives');
}

export async function updateCgroupLimits(
  id: string, 
  payload: { cpuQuotaCores?: number; memoryLimitMb?: number }
): Promise<Container> {
  const res = await fetch(`${API_BASE}/containers/${id}/cgroup`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<Container>(res, 'Failed to update cgroups configuration');
}
