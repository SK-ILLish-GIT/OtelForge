import { TASK_TYPES } from '../api/client';

export function taskLabel(taskType: string): string {
  return TASK_TYPES.find((t) => t.value === taskType)?.label ?? taskType;
}

export function statusClass(status: string): string {
  const s = status.toLowerCase();
  if (['completed', 'verified', 'success', 'partial'].includes(s)) return 'badge badge-success';
  if (['running', 'queued', 'pending'].includes(s)) return 'badge badge-running';
  if (['failed'].includes(s)) return 'badge badge-failed';
  return 'badge';
}

export function formatDateTime(value: string): string {
  return new Date(value).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}

export function isActiveStatus(status: string): boolean {
  return ['RUNNING', 'QUEUED', 'running', 'queued', 'pending'].includes(status);
}

export function isEventActive(event: { status: string }): boolean {
  return isActiveStatus(event.status);
}

export function formatRelativeTime(from: Date): string {
  const seconds = Math.max(0, Math.floor((Date.now() - from.getTime()) / 1000));
  if (seconds < 5) return 'just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  return formatDateTime(from.toISOString());
}

export function instanceLabel(job: { instanceName?: string; instanceHost?: string; instancePort?: number; instanceId: string }): string {
  if (job.instanceName) {
    const host = job.instanceHost ? `${job.instanceHost}${job.instancePort && job.instancePort !== 22 ? `:${job.instancePort}` : ''}` : '';
    return host ? `${job.instanceName} · ${host}` : job.instanceName;
  }
  return job.instanceId.slice(0, 8);
}
