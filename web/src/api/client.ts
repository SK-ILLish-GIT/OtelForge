const API_URL = import.meta.env.VITE_API_URL || '';

export type User = { id: string; email: string; role: string };

function authHeader(): HeadersInit {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...authHeader(),
      ...(options.headers || {}),
    },
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; user: User }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  listInstances: () => request<Instance[]>('/api/v1/instances'),
  createInstance: (body: CreateInstance) =>
    request<Instance>('/api/v1/instances', { method: 'POST', body: JSON.stringify(body) }),
  bulkInstances: (csv: string) =>
    request<{ created: number; instances: Instance[]; errors: string[] }>('/api/v1/instances/bulk', {
      method: 'POST',
      body: JSON.stringify({ csv }),
    }),
  deleteInstance: (id: string) => request<void>(`/api/v1/instances/${id}`, { method: 'DELETE' }),
  updateInstance: (id: string, body: UpdateInstance) =>
    request<Instance>(`/api/v1/instances/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  listEvents: (query = '') => request<Event[]>(`/api/v1/events${query}`),
  getEvent: (id: string) => request<Event>(`/api/v1/events/${id}`),
  createEvent: (body: CreateEvent) =>
    request<Event>('/api/v1/events', { method: 'POST', body: JSON.stringify(body) }),
  rerunFailed: (id: string) => request<Event>(`/api/v1/events/${id}/rerun-failed`, { method: 'POST' }),
  cloneEvent: (id: string, body: Partial<CreateEvent>) =>
    request<Event>(`/api/v1/events/${id}/clone`, { method: 'POST', body: JSON.stringify(body) }),
  rollbackJob: (eventId: string, jobId: string) =>
    request<Event>(`/api/v1/events/${eventId}/jobs/${jobId}/rollback`, { method: 'POST' }),
  adminEvents: (query = '') => request<Event[]>(`/api/v1/admin/events${query}`),
  adminInstances: () => request<Instance[]>('/api/v1/admin/instances'),
};

export type Instance = {
  id: string;
  ownerId?: string;
  ownerEmail?: string;
  name: string;
  host: string;
  port: number;
  sshUser: string;
  createdAt: string;
};

export type CreateInstance = {
  name: string;
  host: string;
  port: number;
  sshUser: string;
  password?: string;
  privateKey?: string;
};

export type UpdateInstance = {
  sshUser: string;
  password?: string;
  privateKey?: string;
};

export type Job = {
  id: string;
  instanceId: string;
  instanceName?: string;
  instanceHost?: string;
  instancePort?: number;
  status: string;
  stdout?: string;
  stderr?: string;
  exitCode?: number;
  checks?: { phase: string; name: string; passed: boolean; message: string }[];
};

export type Event = {
  id: string;
  name: string;
  launcherName: string;
  launcherEmail: string;
  taskType: string;
  status: string;
  totalJobs: number;
  verifiedCount: number;
  failedCount: number;
  createdAt: string;
  instanceIds?: string[];
  jobs?: Job[];
};

export type CreateEvent = {
  name: string;
  launcherName?: string;
  launcherEmail?: string;
  taskType: string;
  configContent?: string;
  instanceIds: string[];
};

export const TASK_TYPES = [
  { value: 'deploy_config', label: 'Deploy OTel Config', config: true },
  { value: 'validate_config', label: 'Validate Config Only', config: true },
  { value: 'restart_collector', label: 'Restart Collector', config: false },
  { value: 'check_status', label: 'Check Status', config: false },
  { value: 'fetch_logs', label: 'Fetch Logs', config: false },
  { value: 'rollback_config', label: 'Rollback Config', config: false },
  { value: 'stop_collector', label: 'Stop Collector', config: false },
  { value: 'ssh_connectivity_test', label: 'SSH Connectivity Test', config: false },
  { value: 'install_otel_agent', label: 'Install OTel Agent', config: false },
];
