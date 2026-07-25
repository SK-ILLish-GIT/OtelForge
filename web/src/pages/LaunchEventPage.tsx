import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api, TASK_TYPES, type Instance } from '../api/client';
import { useAuth } from '../auth/AuthContext';
import { InstancePicker } from '../components/InstancePicker';
import { PageHeader } from '../components/PageHeader';
import { TaskPicker } from '../components/TaskPicker';
import { YamlEditor } from '../components/YamlEditor';

const DEFAULT_CONFIG = `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [debug]
`;

export function LaunchEventPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [instances, setInstances] = useState<Instance[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [name, setName] = useState('');
  const [taskType, setTaskType] = useState(searchParams.get('task') || TASK_TYPES[0].value);
  const [configContent, setConfigContent] = useState(DEFAULT_CONFIG);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const task = TASK_TYPES.find((t) => t.value === taskType);
  const needsConfig = task?.config ?? false;

  useEffect(() => {
    api.listInstances().then((list) => {
      setInstances(list);
      const preselect = searchParams.get('instance');
      if (preselect && list.some((i) => i.id === preselect)) {
        setSelected([preselect]);
      }
    }).catch((e) => setError(String(e)));
  }, [searchParams]);

  function toggleInstance(id: string) {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!selected.length) {
      setError('Select at least one instance');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      const ev = await api.createEvent({
        name: name || `${task?.label ?? taskType} · ${new Date().toLocaleString()}`,
        launcherName: user?.email,
        launcherEmail: user?.email,
        taskType,
        instanceIds: selected,
        configContent: needsConfig ? configContent : undefined,
      });
      navigate(`/events/${ev.id}?live=1`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Launch failed');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page launch-page">
      <PageHeader
        title="Launch event"
        subtitle="Pick a task, choose targets, then watch live output on the next screen."
      />
      {error && <div className="alert alert-error">{error}</div>}

      <form onSubmit={onSubmit} className="launch-grid">
        <div className="launch-main">
          <section className="card launch-section">
            <label className="field">
              Event name
              <span className="hint">Optional — auto-generated if blank</span>
              <input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. prod rollout wave 1" />
            </label>
          </section>

          <section className="card launch-section">
            <h2 className="launch-section-title">Task</h2>
            <TaskPicker value={taskType} onChange={setTaskType} />
          </section>

          {needsConfig && (
            <section className="card launch-section">
              <h2 className="launch-section-title">OTel config</h2>
              <YamlEditor value={configContent} onChange={setConfigContent} />
            </section>
          )}
        </div>

        <aside className="launch-aside">
          <section className="card launch-section launch-targets">
            <h2 className="launch-section-title">Targets</h2>
            <InstancePicker
              instances={instances}
              selected={selected}
              onToggle={toggleInstance}
              onSelectAll={() => setSelected(instances.map((i) => i.id))}
            />
          </section>

          <div className="launch-action-card card">
            <div className="launch-action-summary">
              <p className="launch-action-task">{task?.label ?? taskType}</p>
              <p className="muted">
                {selected.length
                  ? `${selected.length} instance${selected.length === 1 ? '' : 's'} selected`
                  : 'Select at least one target'}
              </p>
            </div>
            <button
              type="submit"
              className="btn btn-primary launch-action-btn"
              disabled={submitting || !instances.length || !selected.length}
            >
              {submitting ? 'Launching…' : 'Launch & watch live'}
            </button>
            <p className="muted launch-live-hint">Opens the live log console immediately.</p>
          </div>
        </aside>
      </form>
    </div>
  );
}
