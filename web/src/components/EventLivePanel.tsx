import { useEffect, useState } from 'react';
import type { Event, Job } from '../api/client';
import { instanceLabel, isActiveStatus } from '../lib/format';
import { ProgressBar } from './ProgressBar';
import { LiveLogConsole } from './LiveLogConsole';
import { StatusBadge } from './StatusBadge';

const STALE_QUEUE_MS = 8000;

type Props = {
  event: Event;
  selectedJobId: string | null;
  onSelectJob: (id: string) => void;
  live?: boolean;
};

export function EventLivePanel({ event, selectedJobId, onSelectJob, live }: Props) {
  const jobs = event.jobs ?? [];
  const selected = jobs.find((j) => j.id === selectedJobId) ?? jobs[0] ?? null;
  const eventActive = isActiveStatus(event.status);
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!eventActive || !jobs.some((j) => j.status === 'QUEUED')) return;
    const timer = window.setInterval(() => setNow(Date.now()), 2000);
    return () => window.clearInterval(timer);
  }, [eventActive, jobs]);

  const ageMs = now - new Date(event.createdAt).getTime();
  const workerLikelyDown = eventActive && jobs.some((j) => j.status === 'QUEUED') && ageMs > STALE_QUEUE_MS;

  return (
    <div className="event-live-panel">
      {workerLikelyDown && (
        <div className="alert alert-error worker-alert">
          Jobs are stuck in <strong>QUEUED</strong> — the worker is probably not running.
          Start it with: <code>docker compose up -d worker</code>
        </div>
      )}
      <div className="event-live-summary card">
        <div className="event-live-summary-top">
          <StatusBadge status={event.status} pulse={eventActive} />
          <ProgressBar verified={event.verifiedCount} failed={event.failedCount} total={event.totalJobs} />
          <span className="muted event-live-counts">
            {event.verifiedCount}/{event.totalJobs} verified
            {event.failedCount > 0 && <span className="error"> · {event.failedCount} failed</span>}
          </span>
        </div>
      </div>

      <div className="event-live-split">
        <aside className="event-job-list">
          <p className="event-job-list-title">Targets ({jobs.length})</p>
          {jobs.map((job) => (
            <JobCard
              key={job.id}
              job={job}
              selected={job.id === selected?.id}
              onClick={() => onSelectJob(job.id)}
            />
          ))}
        </aside>

        <section className="event-log-panel">
          {selected ? (
            <>
              <div className="event-log-panel-header">
                <div>
                  <strong>{selected.instanceName ?? 'Instance'}</strong>
                  <span className="muted"> · {instanceLabel(selected)}</span>
                </div>
                <StatusBadge status={selected.status} pulse={isActiveStatus(selected.status)} />
              </div>
              <LiveLogConsole job={selected} live={live && eventActive} />
            </>
          ) : (
            <div className="loading-block">No jobs in this event.</div>
          )}
        </section>
      </div>
    </div>
  );
}

function JobCard({ job, selected, onClick }: { job: Job; selected: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      className={`event-job-card${selected ? ' selected' : ''}${job.status === 'FAILED' ? ' failed' : ''}`}
      onClick={onClick}
    >
      <div className="event-job-card-top">
        <strong>{job.instanceName ?? job.instanceId.slice(0, 8)}</strong>
        <StatusBadge status={job.status} pulse={isActiveStatus(job.status)} />
      </div>
      <div className="muted event-job-card-host">
        {job.instanceHost
          ? `${job.instanceHost}${job.instancePort && job.instancePort !== 22 ? `:${job.instancePort}` : ''}`
          : '—'}
      </div>
      {(job.stdout || job.stderr) && (
        <div className="event-job-card-preview muted">
          {(job.stdout || job.stderr || '').split('\n').filter(Boolean).slice(-1)[0]?.slice(0, 60)}
        </div>
      )}
    </button>
  );
}
