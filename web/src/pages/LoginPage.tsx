import { useState, type FormEvent } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import '../App.css';

export function LoginPage() {
  const { user, login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  if (user) return <Navigate to="/instances" replace />;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    try {
      await login(email, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    }
  }

  return (
    <div className="login-page">
      <div className="card login-card">
        <div className="login-brand">
          <span className="brand-mark">OF</span>
          <h1>OtelForge</h1>
        </div>
        <p className="login-tagline">Deploy and manage OpenTelemetry collectors across your fleet.</p>
        <form onSubmit={onSubmit} className="stack">
          <label className="field">
            Email
            <input type="email" autoComplete="username" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </label>
          <label className="field">
            Password
            <input type="password" autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </label>
          {error && <div className="alert alert-error">{error}</div>}
          <button type="submit">Sign in</button>
        </form>
      </div>
    </div>
  );
}
