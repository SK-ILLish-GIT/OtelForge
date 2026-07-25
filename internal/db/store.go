package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/otelforge/otelforge/internal/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, created_at`,
		email, passwordHash, role,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) CreateInstance(ctx context.Context, ownerID, name, host string, port int, sshUser string, passwordEnc, keyEnc []byte) (*models.Instance, error) {
	var inst models.Instance
	err := s.pool.QueryRow(ctx, `
		INSERT INTO instances (owner_id, name, host, port, ssh_user, ssh_password_enc, ssh_private_key_enc)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, owner_id, name, host, port, ssh_user, ssh_password_enc, ssh_private_key_enc, ssh_host_key_fingerprint, created_at`,
		ownerID, name, host, port, sshUser, passwordEnc, keyEnc,
	).Scan(&inst.ID, &inst.OwnerID, &inst.Name, &inst.Host, &inst.Port, &inst.SSHUser, &inst.SSHPasswordEnc, &inst.SSHPrivateKeyEnc, &inst.SSHHostKeyFingerprint, &inst.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Store) GetInstance(ctx context.Context, id string) (*models.Instance, error) {
	var inst models.Instance
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, host, port, ssh_user, ssh_password_enc, ssh_private_key_enc, ssh_host_key_fingerprint, created_at
		FROM instances WHERE id = $1`, id,
	).Scan(&inst.ID, &inst.OwnerID, &inst.Name, &inst.Host, &inst.Port, &inst.SSHUser, &inst.SSHPasswordEnc, &inst.SSHPrivateKeyEnc, &inst.SSHHostKeyFingerprint, &inst.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Store) ListInstances(ctx context.Context, ownerID string, admin bool) ([]models.Instance, error) {
	query := `
		SELECT i.id, i.owner_id, u.email, i.name, i.host, i.port, i.ssh_user, i.created_at
		FROM instances i JOIN users u ON u.id = i.owner_id`
	args := []any{}
	if !admin {
		query += ` WHERE i.owner_id = $1`
		args = append(args, ownerID)
	}
	query += ` ORDER BY i.created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Instance
	for rows.Next() {
		var inst models.Instance
		if err := rows.Scan(&inst.ID, &inst.OwnerID, &inst.OwnerEmail, &inst.Name, &inst.Host, &inst.Port, &inst.SSHUser, &inst.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, inst)
	}
	return out, rows.Err()
}

func (s *Store) UpdateInstance(ctx context.Context, id, ownerID, name, host string, port int, sshUser string, enc []byte) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE instances SET name=$1, host=$2, port=$3, ssh_user=$4, ssh_password_enc=$5, ssh_host_key_fingerprint=NULL
		WHERE id=$6 AND owner_id=$7`, name, host, port, sshUser, enc, id, ownerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}
	return nil
}

func (s *Store) UpdateInstanceSSH(ctx context.Context, id, ownerID, sshUser string, passwordEnc, keyEnc []byte) (*models.Instance, error) {
	var inst models.Instance
	err := s.pool.QueryRow(ctx, `
		UPDATE instances SET ssh_user=$1, ssh_password_enc=$2, ssh_private_key_enc=$3
		WHERE id=$4 AND owner_id=$5
		RETURNING id, owner_id, name, host, port, ssh_user, ssh_password_enc, ssh_private_key_enc, ssh_host_key_fingerprint, created_at`,
		sshUser, passwordEnc, keyEnc, id, ownerID,
	).Scan(&inst.ID, &inst.OwnerID, &inst.Name, &inst.Host, &inst.Port, &inst.SSHUser, &inst.SSHPasswordEnc, &inst.SSHPrivateKeyEnc, &inst.SSHHostKeyFingerprint, &inst.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("instance not found")
	}
	return &inst, nil
}

func (s *Store) UpdateInstanceHostKeyFingerprint(ctx context.Context, instanceID, fingerprint string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE instances SET ssh_host_key_fingerprint = $1
		WHERE id = $2 AND (ssh_host_key_fingerprint IS NULL OR ssh_host_key_fingerprint = $1)`,
		fingerprint, instanceID)
	return err
}

func (s *Store) DeleteInstance(ctx context.Context, id, ownerID string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM instances WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}
	return nil
}

type CreateEventParams struct {
	OwnerID        string
	Name           string
	LauncherName   string
	LauncherEmail  string
	TaskType       models.TaskType
	ConfigContent  *string
	ConfigChecksum *string
	RelatedEventID *string
	RollbackScope  *string
	InstanceIDs    []string
}

func (s *Store) CreateEventWithJobs(ctx context.Context, p CreateEventParams) (*models.Event, []models.Job, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var ev models.Event
	err = tx.QueryRow(ctx, `
		INSERT INTO events (owner_id, name, launcher_name, launcher_email, task_type, config_content, config_checksum, status, related_event_id, rollback_scope, total_jobs)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, owner_id, name, launcher_name, launcher_email, task_type, config_content, config_checksum, status, related_event_id, rollback_scope, total_jobs, verified_count, failed_count, created_at`,
		p.OwnerID, p.Name, p.LauncherName, p.LauncherEmail, p.TaskType, p.ConfigContent, p.ConfigChecksum,
		models.EventQueued, p.RelatedEventID, p.RollbackScope, len(p.InstanceIDs),
	).Scan(&ev.ID, &ev.OwnerID, &ev.Name, &ev.LauncherName, &ev.LauncherEmail, &ev.TaskType, &ev.ConfigContent, &ev.ConfigChecksum,
		&ev.Status, &ev.RelatedEventID, &ev.RollbackScope, &ev.TotalJobs, &ev.VerifiedCount, &ev.FailedCount, &ev.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	ev.InstanceIDs = p.InstanceIDs
	var jobs []models.Job
	for _, iid := range p.InstanceIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO event_instances (event_id, instance_id) VALUES ($1,$2)`, ev.ID, iid); err != nil {
			return nil, nil, err
		}
		var job models.Job
		err = tx.QueryRow(ctx, `
			INSERT INTO jobs (event_id, instance_id, status) VALUES ($1,$2,$3)
			RETURNING id, event_id, instance_id, status`,
			ev.ID, iid, models.JobQueued,
		).Scan(&job.ID, &job.EventID, &job.InstanceID, &job.Status)
		if err != nil {
			return nil, nil, err
		}
		jobs = append(jobs, job)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &ev, jobs, nil
}

func (s *Store) ListEvents(ctx context.Context, f models.EventFilter) ([]models.Event, error) {
	var b strings.Builder
	b.WriteString(`SELECT id, owner_id, name, launcher_name, launcher_email, task_type, config_content, config_checksum, status, related_event_id, rollback_scope, total_jobs, verified_count, failed_count, created_at FROM events WHERE 1=1`)
	args := []any{}
	n := 1
	if f.OwnerID != "" {
		fmt.Fprintf(&b, " AND owner_id = $%d", n)
		args = append(args, f.OwnerID)
		n++
	}
	if f.LauncherEmail != "" {
		fmt.Fprintf(&b, " AND launcher_email ILIKE $%d", n)
		args = append(args, "%"+f.LauncherEmail+"%")
		n++
	}
	if f.Status != "" {
		fmt.Fprintf(&b, " AND status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.TaskType != "" {
		fmt.Fprintf(&b, " AND task_type = $%d", n)
		args = append(args, f.TaskType)
		n++
	}
	if f.From != nil {
		fmt.Fprintf(&b, " AND created_at >= $%d", n)
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		fmt.Fprintf(&b, " AND created_at <= $%d", n)
		args = append(args, *f.To)
		n++
	}
	b.WriteString(" ORDER BY created_at DESC")

	rows, err := s.pool.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Event
	for rows.Next() {
		var ev models.Event
		if err := rows.Scan(&ev.ID, &ev.OwnerID, &ev.Name, &ev.LauncherName, &ev.LauncherEmail, &ev.TaskType, &ev.ConfigContent, &ev.ConfigChecksum,
			&ev.Status, &ev.RelatedEventID, &ev.RollbackScope, &ev.TotalJobs, &ev.VerifiedCount, &ev.FailedCount, &ev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (s *Store) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	var ev models.Event
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, launcher_name, launcher_email, task_type, config_content, config_checksum, status, related_event_id, rollback_scope, total_jobs, verified_count, failed_count, created_at
		FROM events WHERE id = $1`, id,
	).Scan(&ev.ID, &ev.OwnerID, &ev.Name, &ev.LauncherName, &ev.LauncherEmail, &ev.TaskType, &ev.ConfigContent, &ev.ConfigChecksum,
		&ev.Status, &ev.RelatedEventID, &ev.RollbackScope, &ev.TotalJobs, &ev.VerifiedCount, &ev.FailedCount, &ev.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `SELECT instance_id FROM event_instances WHERE event_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var iid string
		if err := rows.Scan(&iid); err != nil {
			return nil, err
		}
		ev.InstanceIDs = append(ev.InstanceIDs, iid)
	}
	return &ev, rows.Err()
}

func (s *Store) ListJobsForEvent(ctx context.Context, eventID string) ([]models.Job, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT j.id, j.event_id, j.instance_id, j.status, j.stdout, j.stderr, j.exit_code, j.started_at, j.finished_at,
		       i.name, i.host, i.port
		FROM jobs j
		JOIN instances i ON i.id = j.instance_id
		WHERE j.event_id = $1 ORDER BY j.id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.EventID, &j.InstanceID, &j.Status, &j.Stdout, &j.Stderr, &j.ExitCode, &j.StartedAt, &j.FinishedAt,
			&j.InstanceName, &j.InstanceHost, &j.InstancePort); err != nil {
			return nil, err
		}
		checks, err := s.ListChecksForJob(ctx, j.ID)
		if err != nil {
			return nil, err
		}
		j.Checks = checks
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (s *Store) ListChecksForJob(ctx context.Context, jobID string) ([]models.JobCheck, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, job_id, phase, name, passed, message FROM job_checks WHERE job_id = $1`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.JobCheck
	for rows.Next() {
		var c models.JobCheck
		if err := rows.Scan(&c.ID, &c.JobID, &c.Phase, &c.Name, &c.Passed, &c.Message); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetJob(ctx context.Context, id string) (*models.Job, error) {
	var j models.Job
	err := s.pool.QueryRow(ctx, `
		SELECT id, event_id, instance_id, status, stdout, stderr, exit_code, started_at, finished_at
		FROM jobs WHERE id = $1`, id,
	).Scan(&j.ID, &j.EventID, &j.InstanceID, &j.Status, &j.Stdout, &j.Stderr, &j.ExitCode, &j.StartedAt, &j.FinishedAt)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *Store) UpdateJobRunning(ctx context.Context, jobID string) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET status=$1, started_at=$2, stdout=$3, stderr=$4 WHERE id=$5`,
		models.JobRunning, now, "Job started. Preparing SSH connection…\n", "", jobID)
	return err
}

func (s *Store) UpdateJobActivity(ctx context.Context, jobID, stdout, stderr string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET stdout=$1, stderr=$2 WHERE id=$3 AND status=$4`,
		stdout, stderr, jobID, models.JobRunning)
	return err
}

func (s *Store) UpdateJobResult(ctx context.Context, jobID string, status models.JobStatus, stdout, stderr string, exitCode int) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET status=$1, stdout=$2, stderr=$3, exit_code=$4, finished_at=$5 WHERE id=$6`,
		status, stdout, stderr, exitCode, now, jobID)
	return err
}

func (s *Store) AddJobCheck(ctx context.Context, jobID string, phase models.CheckPhase, name string, passed bool, message string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO job_checks (job_id, phase, name, passed, message) VALUES ($1,$2,$3,$4,$5)`,
		jobID, phase, name, passed, message)
	return err
}

func (s *Store) UpdateEventStatus(ctx context.Context, eventID string, status models.EventStatus) error {
	_, err := s.pool.Exec(ctx, `UPDATE events SET status=$1 WHERE id=$2`, status, eventID)
	return err
}

func (s *Store) RollupEvent(ctx context.Context, eventID string) error {
	var verified, failed, total int
	err := s.pool.QueryRow(ctx, `
		SELECT
		 COUNT(*) FILTER (WHERE status = 'VERIFIED'),
		 COUNT(*) FILTER (WHERE status = 'FAILED'),
		 COUNT(*)
		FROM jobs WHERE event_id = $1`, eventID).Scan(&verified, &failed, &total)
	if err != nil {
		return err
	}

	var pending int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM jobs WHERE event_id = $1 AND status IN ('QUEUED','RUNNING')`, eventID).Scan(&pending)
	if err != nil {
		return err
	}

	status := models.EventRunning
	if pending == 0 {
		switch {
		case verified == total:
			status = models.EventCompleted
		case failed == total:
			status = models.EventFailed
		default:
			status = models.EventPartial
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE events SET status=$1, verified_count=$2, failed_count=$3, total_jobs=$4 WHERE id=$5`,
		status, verified, failed, total, eventID)
	return err
}

func (s *Store) ResetFailedJobs(ctx context.Context, eventID string) ([]models.Job, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE jobs SET status=$1, stdout=NULL, stderr=NULL, exit_code=NULL, started_at=NULL, finished_at=NULL
		WHERE event_id=$2 AND status=$3
		RETURNING id, event_id, instance_id, status`,
		models.JobQueued, eventID, models.JobFailed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.EventID, &j.InstanceID, &j.Status); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	if err := s.UpdateEventStatus(ctx, eventID, models.EventRunning); err != nil {
		return nil, err
	}
	return jobs, rows.Err()
}

func (s *Store) UserOwnsInstance(ctx context.Context, userID, instanceID string) (bool, error) {
	var owner string
	err := s.pool.QueryRow(ctx, `SELECT owner_id FROM instances WHERE id=$1`, instanceID).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == userID, nil
}

func (s *Store) UserOwnsEvent(ctx context.Context, userID, eventID string) (bool, error) {
	var owner string
	err := s.pool.QueryRow(ctx, `SELECT owner_id FROM events WHERE id=$1`, eventID).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == userID, nil
}
