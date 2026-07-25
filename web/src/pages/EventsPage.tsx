import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, type Event } from '../api/client';
import { EmptyState } from '../components/EmptyState';
import { PageHeader } from '../components/PageHeader';
import { ProgressBar } from '../components/ProgressBar';
import { StatusBadge } from '../components/StatusBadge';
import { formatDateTime, isEventActive, taskLabel } from '../lib/format';

export function EventsPage() {
  const [list, setList] = useState<Event[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setList(await api.listEvents());
  }, []);

  useEffect(() => {
    load()
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false));
  }, [load]);

  useEffect(() => {
    if (!list.some(isEventActive)) return;
    const timer = window.setInterval(() => {
      load().catch(() => {});
    }, 5000);
    return () => window.clearInterval(timer);
  }, [list, load]);

  return (
    <div className="page">
      <PageHeader
        title="Events"
        subtitle="History of deploy and operational runs across your instances."
        actions={<Link to="/launch" className="btn btn-primary">Launch event</Link>}
      />
      {error && <div className="alert alert-error">{error}</div>}

      {loading ? (
        <div className="loading-block">Loading events…</div>
      ) : !list.length ? (
        <EmptyState
          title="No events yet"
          hint="Launch a task against one or more instances to get started."
          action={<Link to="/launch" className="btn btn-primary">Launch event</Link>}
        />
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Task</th>
                <th>Status</th>
                <th>Progress</th>
                <th>Launcher</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {list.map((ev) => (
                <tr key={ev.id} className={isEventActive(ev) ? 'row-active' : undefined}>
                  <td><Link to={`/events/${ev.id}`}><strong>{ev.name}</strong></Link></td>
                  <td>{taskLabel(ev.taskType)}</td>
                  <td><StatusBadge status={ev.status} pulse={isEventActive(ev)} /></td>
                  <td className="progress-cell">
                    <ProgressBar verified={ev.verifiedCount} failed={ev.failedCount} total={ev.totalJobs} />
                    <span className="muted progress-label">
                      {ev.verifiedCount}/{ev.totalJobs} ok
                      {ev.failedCount > 0 && <span className="error"> · {ev.failedCount} failed</span>}
                    </span>
                  </td>
                  <td>{ev.launcherEmail}</td>
                  <td>{formatDateTime(ev.createdAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
