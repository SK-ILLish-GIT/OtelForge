import { useEffect, useRef } from 'react';
import type { Job } from '../api/client';
import { isActiveStatus } from '../lib/format';
import { JobTimeline } from './JobTimeline';

type Props = {
  job: Job;
  live?: boolean;
};

export function LiveLogConsole({ job, live }: Props) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const content = buildConsoleText(job);
  const waiting = live && isActiveStatus(job.status) && !job.stdout && !job.stderr;

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [content, job.checks?.length]);

  return (
    <div className="live-console">
      <div className="live-console-header">
        <JobTimeline job={job} />
        {live && isActiveStatus(job.status) && (
          <span className="live-indicator">
            <span className="live-dot" /> Live
          </span>
        )}
      </div>
      <div className="live-console-body" ref={scrollRef}>
        {waiting && (
          <div className="live-console-waiting">
            <span className="terminal-cursor">Waiting for worker</span>
          </div>
        )}
        {content ? (
          <pre className="live-console-output">{content}</pre>
        ) : !waiting ? (
          <p className="muted live-console-empty">No output captured for this job.</p>
        ) : null}
      </div>
    </div>
  );
}

function buildConsoleText(job: Job): string {
  const lines: string[] = [];

  for (const check of job.checks ?? []) {
    const mark = check.passed ? 'OK' : 'FAIL';
    lines.push(`[${check.phase}] ${check.name}: ${mark} — ${check.message}`);
  }

  if (job.stdout) lines.push(job.stdout.replace(/\n$/, ''));
  if (job.stderr) lines.push(job.stderr.replace(/\n$/, ''));

  return lines.join('\n');
}
