import { Link } from 'react-router-dom';
import type { Instance } from '../api/client';

type Props = {
  instances: Instance[];
  selected: string[];
  onToggle: (id: string) => void;
  onSelectAll: () => void;
};

export function InstancePicker({ instances, selected, onToggle, onSelectAll }: Props) {
  if (!instances.length) {
    return (
      <div className="instance-picker-empty">
        <p className="muted">No instances registered yet.</p>
        <Link to="/instances" className="btn btn-secondary btn-sm">Add instance</Link>
      </div>
    );
  }

  return (
    <div className="instance-picker">
      <div className="instance-picker-header">
        <span className="instance-picker-count">{selected.length} of {instances.length} selected</span>
        {instances.length > 1 && (
          <button type="button" className="link-btn" onClick={onSelectAll}>Select all</button>
        )}
      </div>
      <div className="instance-picker-list">
        {instances.map((inst) => {
          const isSelected = selected.includes(inst.id);
          return (
            <button
              key={inst.id}
              type="button"
              className={`instance-picker-card${isSelected ? ' selected' : ''}`}
              onClick={() => onToggle(inst.id)}
            >
              <span className={`instance-picker-check${isSelected ? ' checked' : ''}`} aria-hidden />
              <span className="instance-picker-info">
                <strong>{inst.name}</strong>
                <span className="muted">{inst.host}:{inst.port} · {inst.sshUser}</span>
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
