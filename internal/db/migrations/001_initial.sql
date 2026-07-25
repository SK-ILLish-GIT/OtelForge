-- gen_random_uuid() is built into PostgreSQL 13+

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('user', 'admin')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instances (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  host TEXT NOT NULL,
  port INT NOT NULL DEFAULT 22,
  ssh_user TEXT NOT NULL,
  ssh_password_enc BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  launcher_name TEXT NOT NULL,
  launcher_email TEXT NOT NULL,
  task_type TEXT NOT NULL,
  config_content TEXT,
  config_checksum TEXT,
  status TEXT NOT NULL,
  related_event_id UUID REFERENCES events(id),
  rollback_scope TEXT,
  total_jobs INT NOT NULL DEFAULT 0,
  verified_count INT NOT NULL DEFAULT 0,
  failed_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_launcher_email ON events(launcher_email);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);
CREATE INDEX IF NOT EXISTS idx_events_task_type ON events(task_type);
CREATE INDEX IF NOT EXISTS idx_events_owner_id ON events(owner_id);

CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  instance_id UUID NOT NULL REFERENCES instances(id),
  status TEXT NOT NULL,
  stdout TEXT,
  stderr TEXT,
  exit_code INT,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_event_id ON jobs(event_id);

CREATE TABLE IF NOT EXISTS job_checks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  phase TEXT NOT NULL CHECK (phase IN ('pre', 'post')),
  name TEXT NOT NULL,
  passed BOOLEAN NOT NULL,
  message TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS event_instances (
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  instance_id UUID NOT NULL REFERENCES instances(id),
  PRIMARY KEY (event_id, instance_id)
);
