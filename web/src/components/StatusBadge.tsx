import { statusClass } from '../lib/format';

export function StatusBadge({ status, pulse }: { status: string; pulse?: boolean }) {
  const cls = statusClass(status) + (pulse ? ' badge-pulse' : '');
  return <span className={cls}>{status}</span>;
}
