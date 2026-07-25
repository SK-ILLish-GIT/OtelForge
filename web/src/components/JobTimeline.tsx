import type { Job } from '../api/client';
import { isActiveStatus } from '../lib/format';

type Step = {
  id: string;
  label: string;
  state: 'done' | 'active' | 'pending' | 'failed';
};

export function buildJobSteps(job: Job): Step[] {
  const pre = job.checks?.find((c) => c.phase === 'pre' && c.name === 'ssh_connectivity');
  const post = job.checks?.find((c) => c.phase === 'post' && c.name === 'task_execution');

  const queued: Step = {
    id: 'queued',
    label: 'Queued',
    state: job.status === 'QUEUED' ? 'active' : 'done',
  };

  let connectState: Step['state'] = 'pending';
  if (pre?.passed) connectState = 'done';
  else if (pre && !pre.passed) connectState = 'failed';
  else if (job.status === 'RUNNING') connectState = 'active';
  else if (job.status !== 'QUEUED') connectState = 'done';

  const connect: Step = {
    id: 'connect',
    label: 'SSH connect',
    state: connectState,
  };

  let runState: Step['state'] = 'pending';
  if (post?.passed) runState = 'done';
  else if (post && !post.passed) runState = 'failed';
  else if (job.status === 'RUNNING' && pre?.passed) runState = 'active';
  else if (job.status === 'VERIFIED' || job.status === 'FAILED') runState = post ? (post.passed ? 'done' : 'failed') : 'done';

  const run: Step = {
    id: 'run',
    label: 'Run task',
    state: runState,
  };

  let doneState: Step['state'] = 'pending';
  if (job.status === 'VERIFIED') doneState = 'done';
  else if (job.status === 'FAILED') doneState = 'failed';
  else if (!isActiveStatus(job.status)) doneState = 'done';

  const done: Step = {
    id: 'done',
    label: job.status === 'FAILED' ? 'Failed' : 'Complete',
    state: doneState,
  };

  return [queued, connect, run, done];
}

export function JobTimeline({ job }: { job: Job }) {
  const steps = buildJobSteps(job);
  return (
    <ol className="job-timeline">
      {steps.map((step, i) => (
        <li key={step.id} className={`job-step job-step-${step.state}${i < steps.length - 1 ? ' has-next' : ''}`}>
          <span className="job-step-dot" aria-hidden />
          <span className="job-step-label">{step.label}</span>
        </li>
      ))}
    </ol>
  );
}
