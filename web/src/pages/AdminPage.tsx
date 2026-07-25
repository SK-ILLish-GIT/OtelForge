import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { api, TASK_TYPES, type Event, type Instance } from '../api/client';
import { EmptyState } from '../components/EmptyState';
import { PageHeader } from '../components/PageHeader';
import { StatusBadge } from '../components/StatusBadge';
import { formatDateTime, taskLabel } from '../lib/format';

export function AdminPage() {
  const [filters, setFilters] = useState({ launcherEmail: '', status: '', taskType: '' });
  const [events, setEvents] = useState<Event[]>([]);
  const [instances, setInstances] = useState<Instance[]>([]);
  const [error, setError] = useState('');
  const [loaded, setLoaded] = useState(false);
  const [searching, setSearching] = useState(false);

  function buildQuery() {
    const params = new URLSearchParams();
    if (filters.launcherEmail) params.set('launcherEmail', filters.launcherEmail);
    if (filters.status) params.set('status', filters.status);
    if (filters.taskType) params.set('taskType', filters.taskType);
    const q = params.toString();
    return q ? `?${q}` : '';
  }

  async function onSearch(e: FormEvent) {
    e.preventDefault();
    setError('');
    setSearching(true);
    try {
      const [evList, instList] = await Promise.all([
        api.adminEvents(buildQuery()),
        api.adminInstances(),
      ]);
      setEvents(evList);
      setInstances(instList);
      setLoaded(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
    } finally {
      setSearching(false);
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="Admin"
        subtitle="Cross-user view of events and registered instances."
      />
      {error && <div className="alert alert-error">{error}</div>}

      <form onSubmit={onSearch} className="card grid filters">
        <label className="field" style={{ marginBottom: 0 }}>
          Launcher email
          <input
            placeholder="user@company.com"
            value={filters.launcherEmail}
            onChange={(e) => setFilters({ ...filters, launcherEmail: e.target.value })}
          />
        </label>
        <label className="field" style={{ marginBottom: 0 }}>
          Status
          <select value={filters.status} onChange={(e) => setFilters({ ...filters, status: e.target.value })}>
            <option value="">Any status</option>
            <option value="QUEUED">Queued</option>
            <option value="RUNNING">Running</option>
            <option value="COMPLETED">Completed</option>
            <option value="PARTIAL">Partial</option>
            <option value="FAILED">Failed</option>
          </select>
        </label>
        <label className="field" style={{ marginBottom: 0 }}>
          Task
          <select value={filters.taskType} onChange={(e) => setFilters({ ...filters, taskType: e.target.value })}>
            <option value="">Any task</option>
            {TASK_TYPES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
        </label>
        <button type="submit" disabled={searching}>{searching ? 'Searching…' : 'Search'}</button>
      </form>

      {loaded && (
        <>
          <h2>Events ({events.length})</h2>
          {events.length === 0 ? (
            <EmptyState title="No matching events" />
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Task</th>
                    <th>Status</th>
                    <th>Launcher</th>
                    <th>Created</th>
                  </tr>
                </thead>
                <tbody>
                  {events.map((ev) => (
                    <tr key={ev.id}>
                      <td><Link to={`/events/${ev.id}`}><strong>{ev.name}</strong></Link></td>
                      <td>{taskLabel(ev.taskType)}</td>
                      <td><StatusBadge status={ev.status} /></td>
                      <td>{ev.launcherEmail}</td>
                      <td>{formatDateTime(ev.createdAt)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <h2>All instances ({instances.length})</h2>
          {instances.length === 0 ? (
            <EmptyState title="No instances registered" />
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Host</th>
                    <th>Owner</th>
                  </tr>
                </thead>
                <tbody>
                  {instances.map((inst) => (
                    <tr key={inst.id}>
                      <td><strong>{inst.name}</strong></td>
                      <td className="host-cell">{inst.host}:{inst.port}</td>
                      <td>{inst.ownerEmail || inst.ownerId}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
