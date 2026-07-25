import { useEffect, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api, type Instance } from '../api/client';
import { EmptyState } from '../components/EmptyState';
import { Modal } from '../components/Modal';
import { PageHeader } from '../components/PageHeader';
import { useAuth } from '../auth/AuthContext';

export function InstancesPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [list, setList] = useState<Instance[]>([]);
  const [form, setForm] = useState({ name: '', host: '', port: 22, sshUser: '', password: '', privateKey: '' });
  const [editTarget, setEditTarget] = useState<Instance | null>(null);
  const [editForm, setEditForm] = useState({ sshUser: '', password: '', privateKey: '' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState('');

  async function load() {
    setList(await api.listInstances());
  }

  useEffect(() => {
    load()
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false));
  }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    try {
      await api.createInstance({
        ...form,
        password: form.password || undefined,
        privateKey: form.privateKey || undefined,
      });
      setForm({ name: '', host: '', port: 22, sshUser: '', password: '', privateKey: '' });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add instance');
    }
  }

  function openEdit(inst: Instance) {
    setEditTarget(inst);
    setEditForm({ sshUser: inst.sshUser, password: '', privateKey: '' });
    setError('');
  }

  async function saveEdit(e: FormEvent) {
    e.preventDefault();
    if (!editTarget) return;
    if (!editForm.password && !editForm.privateKey) {
      setError('Provide a new password or private key to update credentials');
      return;
    }
    setBusy('edit');
    setError('');
    try {
      await api.updateInstance(editTarget.id, {
        sshUser: editForm.sshUser,
        password: editForm.password || undefined,
        privateKey: editForm.privateKey || undefined,
      });
      setEditTarget(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Update failed');
    } finally {
      setBusy('');
    }
  }

  async function quickTask(inst: Instance, taskType: string, label: string) {
    setBusy(inst.id + taskType);
    setError('');
    try {
      const ev = await api.createEvent({
        name: `${label} · ${inst.name}`,
        launcherName: user?.email,
        launcherEmail: user?.email,
        taskType,
        instanceIds: [inst.id],
      });
      navigate(`/events/${ev.id}?live=1`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Launch failed');
    } finally {
      setBusy('');
    }
  }

  function confirmDelete(inst: Instance) {
    if (!window.confirm(`Delete instance "${inst.name}" (${inst.host})? This cannot be undone.`)) return;
    setBusy('delete-' + inst.id);
    api.deleteInstance(inst.id)
      .then(load)
      .catch((e) => setError(String(e)))
      .finally(() => setBusy(''));
  }

  return (
    <div className="page">
      <PageHeader
        title="Instances"
        subtitle="SSH targets for collector deploys and operational tasks."
        actions={
          <>
            <Link to="/instances/bulk" className="btn btn-secondary">Bulk import</Link>
            <Link to="/launch" className="btn btn-primary">Launch event</Link>
          </>
        }
      />
      {error && <div className="alert alert-error">{error}</div>}

      <div className="split-layout">
        <form onSubmit={onSubmit} className="card stack">
          <p className="card-title">Add instance</p>
          <label className="field">
            Name
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="prod-api-1" required />
          </label>
          <div className="grid">
            <label className="field">
              Host
              <input value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} placeholder="10.0.1.10" required />
            </label>
            <label className="field">
              Port
              <input type="number" value={form.port} onChange={(e) => setForm({ ...form, port: +e.target.value })} />
            </label>
          </div>
          <label className="field">
            SSH user
            <input value={form.sshUser} onChange={(e) => setForm({ ...form, sshUser: e.target.value })} placeholder="ec2-user" required />
          </label>
          <label className="field">
            SSH password
            <span className="hint">Optional if using a private key</span>
            <input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
          </label>
          <label className="field">
            Private key (PEM)
            <span className="hint">Optional if using password auth</span>
            <textarea rows={4} value={form.privateKey} onChange={(e) => setForm({ ...form, privateKey: e.target.value })} />
          </label>
          <button type="submit">Add instance</button>
        </form>

        <div>
          {loading ? (
            <div className="loading-block">Loading instances…</div>
          ) : list.length === 0 ? (
            <EmptyState title="No instances yet" hint="Add a host on the left or bulk-import from CSV." />
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Host</th>
                    <th>User</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {list.map((i) => (
                    <tr key={i.id}>
                      <td><strong>{i.name}</strong></td>
                      <td className="host-cell">{i.host}:{i.port}</td>
                      <td>{i.sshUser}</td>
                      <td className="td-actions instance-actions">
                        <button
                          type="button"
                          className="btn btn-secondary btn-sm"
                          disabled={!!busy}
                          onClick={() => quickTask(i, 'ssh_connectivity_test', 'SSH test')}
                        >
                          Test SSH
                        </button>
                        <Link to={`/launch?instance=${i.id}`} className="btn btn-secondary btn-sm">Launch</Link>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={() => openEdit(i)}>Edit</button>
                        <button
                          type="button"
                          className="btn-danger btn-sm"
                          disabled={!!busy}
                          onClick={() => confirmDelete(i)}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      <Modal
        title={editTarget ? `Edit ${editTarget.name}` : 'Edit instance'}
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        footer={
          <>
            <button type="button" className="btn btn-secondary" onClick={() => setEditTarget(null)}>Cancel</button>
            <button type="submit" form="edit-instance-form" disabled={busy === 'edit'}>
              {busy === 'edit' ? 'Saving…' : 'Save credentials'}
            </button>
          </>
        }
      >
        <form id="edit-instance-form" onSubmit={saveEdit} className="stack">
          <p className="muted">Update SSH credentials for <strong>{editTarget?.host}</strong>. Leave blank fields unchanged — provide at least password or private key.</p>
          <label className="field">
            SSH user
            <input value={editForm.sshUser} onChange={(e) => setEditForm({ ...editForm, sshUser: e.target.value })} required />
          </label>
          <label className="field">
            New password
            <input type="password" value={editForm.password} onChange={(e) => setEditForm({ ...editForm, password: e.target.value })} />
          </label>
          <label className="field">
            New private key (PEM)
            <textarea rows={4} value={editForm.privateKey} onChange={(e) => setEditForm({ ...editForm, privateKey: e.target.value })} />
          </label>
        </form>
      </Modal>
    </div>
  );
}
