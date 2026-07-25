import { useState } from 'react';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { api } from '../api/client';
import { EventLivePanel } from '../components/EventLivePanel';
import { PageHeader } from '../components/PageHeader';
import { useAutoSelectJob, useEventPolling } from '../hooks/useEventPolling';
import { formatDateTime, formatRelativeTime, isEventActive, taskLabel } from '../lib/format';

export function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const liveMode = searchParams.get('live') !== '0';
  const { event, error, setError, loading, lastUpdated, refresh } = useEventPolling(id);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const [busy, setBusy] = useState('');

  useAutoSelectJob(event?.jobs, selectedJobId, setSelectedJobId);

  async function run(action: string, fn: () => ReturnType<typeof api.rerunFailed>) {
    if (!event) return;
    setBusy(action);
    setError('');
    try {
      const ev = await fn();
      navigate(`/events/${ev.id}?live=1`);
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy('');
    }
  }

  if (loading && !event) {
    return <div className="page event-detail-page"><div className="loading-block">Loading event…</div></div>;
  }

  const eventActive = event ? isEventActive(event) : false;

  return (
    <div className="page event-detail-page">
      <PageHeader
        title={event?.name ?? 'Event'}
        subtitle={event ? `${taskLabel(event.taskType)} · ${formatDateTime(event.createdAt)}` : undefined}
        back={<Link to="/events" className="back-link">← Back to events</Link>}
        actions={event && (
          <>
            {eventActive && liveMode && (
              <span className="live-banner">
                <span className="live-dot" /> Watching live
              </span>
            )}
            <button type="button" className="btn btn-secondary" disabled={!!busy} onClick={() => refresh()}>
              Refresh
            </button>
            {lastUpdated && eventActive && (
              <span className="muted last-updated">Updated {formatRelativeTime(lastUpdated)}</span>
            )}
            {event.failedCount > 0 && (
              <button
                type="button"
                className="btn btn-secondary"
                disabled={!!busy}
                onClick={() => run('rerun', () => api.rerunFailed(event.id))}
              >
                {busy === 'rerun' ? 'Running…' : 'Re-run failed'}
              </button>
            )}
            <button
              type="button"
              className="btn btn-secondary"
              disabled={!!busy}
              onClick={() => run('clone', () => api.cloneEvent(event.id, { name: `${event.name} (clone)` }))}
            >
              {busy === 'clone' ? 'Cloning…' : 'Clone'}
            </button>
            {selectedJobId && event.jobs?.find((j) => j.id === selectedJobId)?.status === 'FAILED' && (
              <button
                type="button"
                className="btn btn-secondary"
                disabled={!!busy}
                onClick={() => run(`rollback-${selectedJobId}`, () => api.rollbackJob(event.id, selectedJobId))}
              >
                Rollback selected
              </button>
            )}
          </>
        )}
      />

      {error && <div className="alert alert-error">{error}</div>}

      {event && (
        <EventLivePanel
          event={event}
          selectedJobId={selectedJobId}
          onSelectJob={setSelectedJobId}
          live={liveMode}
        />
      )}
    </div>
  );
}
