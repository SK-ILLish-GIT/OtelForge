import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { PageHeader } from '../components/PageHeader';

const SAMPLE = `name,host,port,username,password
prod-api-1,10.0.1.10,22,ubuntu,secret`;

export function BulkAddPage() {
  const [csv, setCsv] = useState(SAMPLE);
  const [result, setResult] = useState('');
  const [error, setError] = useState('');

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setResult('');
    try {
      const res = await api.bulkInstances(csv);
      setResult(`Created ${res.created} instance(s). ${res.errors.length ? `Errors: ${res.errors.join('; ')}` : 'No errors.'}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="Bulk import"
        subtitle="CSV columns: name, host, port, username, password"
        back={<Link to="/instances" className="back-link">← Back to instances</Link>}
      />
      {error && <div className="alert alert-error">{error}</div>}
      {result && <div className="alert alert-success">{result}</div>}

      <form onSubmit={onSubmit} className="card stack page-form">
        <label className="field">
          CSV content
          <textarea rows={10} value={csv} onChange={(e) => setCsv(e.target.value)} />
        </label>
        <button type="submit">Import instances</button>
      </form>
    </div>
  );
}
