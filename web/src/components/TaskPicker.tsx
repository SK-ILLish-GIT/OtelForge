import { TASK_TYPES } from '../api/client';

const TASK_HINTS: Record<string, string> = {
  install_otel_agent: 'Install collector binary + systemd unit.',
  deploy_config: 'Backup, deploy YAML, validate, restart.',
  validate_config: 'Validate YAML remotely without applying.',
  ssh_connectivity_test: 'SSH reachability only.',
  restart_collector: 'Restart otelcol service.',
  check_status: 'Report collector install/status.',
  fetch_logs: 'Fetch recent journal logs.',
  rollback_config: 'Restore previous config backup.',
  stop_collector: 'Stop otelcol service.',
};

const TASK_GROUPS: { title: string; tasks: string[] }[] = [
  { title: 'Deploy', tasks: ['deploy_config', 'validate_config', 'rollback_config'] },
  { title: 'Operations', tasks: ['restart_collector', 'stop_collector', 'check_status', 'fetch_logs'] },
  { title: 'Setup & test', tasks: ['install_otel_agent', 'ssh_connectivity_test'] },
];

type Props = {
  value: string;
  onChange: (value: string) => void;
};

export function TaskPicker({ value, onChange }: Props) {
  const hint = TASK_HINTS[value];

  return (
    <div className="task-picker">
      {TASK_GROUPS.map((group) => (
        <div key={group.title} className="task-group">
          <p className="task-group-title">{group.title}</p>
          <div className="task-options">
            {group.tasks.map((taskValue) => {
              const meta = TASK_TYPES.find((t) => t.value === taskValue);
              if (!meta) return null;
              const active = value === taskValue;
              return (
                <button
                  key={taskValue}
                  type="button"
                  className={`task-option${active ? ' active' : ''}`}
                  onClick={() => onChange(taskValue)}
                >
                  <span className="task-option-label">{meta.label}</span>
                  {meta.config && <span className="task-option-tag">YAML</span>}
                </button>
              );
            })}
          </div>
        </div>
      ))}
      {hint && <p className="task-hint">{hint}</p>}
    </div>
  );
}
