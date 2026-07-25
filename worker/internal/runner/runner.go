package runner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/otelforge/otelforge/internal/crypto"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/models"
	sshclient "github.com/otelforge/otelforge/internal/ssh"
	"github.com/otelforge/otelforge/internal/tasks"
)

type Runner struct {
	store         *db.Store
	encryptionKey string
}

func New(store *db.Store, encryptionKey string) *Runner {
	return &Runner{store: store, encryptionKey: encryptionKey}
}

func (r *Runner) Handle(ctx context.Context, jobID string) error {
	job, err := r.store.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	ev, err := r.store.GetEvent(ctx, job.EventID)
	if err != nil {
		return err
	}
	inst, err := r.store.GetInstance(ctx, job.InstanceID)
	if err != nil {
		return err
	}

	if err := r.store.UpdateJobRunning(ctx, jobID); err != nil {
		return err
	}
	_ = r.store.UpdateEventStatus(ctx, ev.ID, models.EventRunning)

	sshCfg := sshclient.Config{
		Host: inst.Host, Port: inst.Port, User: inst.SSHUser,
	}
	if inst.SSHHostKeyFingerprint != nil {
		sshCfg.ExpectedHostKeyFingerprint = *inst.SSHHostKeyFingerprint
	} else {
		instanceID := inst.ID
		sshCfg.PinHostKey = func(fingerprint string) error {
			return r.store.UpdateInstanceHostKeyFingerprint(ctx, instanceID, fingerprint)
		}
	}
	if len(inst.SSHPasswordEnc) > 0 {
		password, err := crypto.Decrypt(inst.SSHPasswordEnc, r.encryptionKey)
		if err != nil {
			return r.fail(ctx, jobID, ev.ID, "decrypt password", err)
		}
		sshCfg.Password = string(password)
	}
	if len(inst.SSHPrivateKeyEnc) > 0 {
		keyPEM, err := crypto.Decrypt(inst.SSHPrivateKeyEnc, r.encryptionKey)
		if err != nil {
			return r.fail(ctx, jobID, ev.ID, "decrypt private key", err)
		}
		sshCfg.PrivateKeyPEM = string(keyPEM)
	}
	if sshCfg.Password == "" && sshCfg.PrivateKeyPEM == "" {
		return r.fail(ctx, jobID, ev.ID, "credentials", fmt.Errorf("instance has no ssh credentials"))
	}

	client := sshclient.New(sshCfg)
	_ = r.store.UpdateJobActivity(ctx, jobID,
		fmt.Sprintf("Job started. Connecting to %s@%s:%d…\n", inst.SSHUser, inst.Host, inst.Port), "")

	if err := client.TestConnectivity(); err != nil {
		_ = r.store.AddJobCheck(ctx, jobID, models.CheckPre, "ssh_connectivity", false, err.Error())
		return r.failJob(ctx, jobID, ev.ID, "", err.Error(), 1)
	}
	_ = r.store.AddJobCheck(ctx, jobID, models.CheckPre, "ssh_connectivity", true, "reachable")
	_ = r.store.UpdateJobActivity(ctx, jobID,
		fmt.Sprintf("SSH connected to %s.\nRunning %s…\n", inst.Name, ev.TaskType), "")

	script, err := tasks.ScriptFor(ev.TaskType)
	if err != nil {
		return r.fail(ctx, jobID, ev.ID, "script", err)
	}

	var configContent *string
	if ev.TaskType.RequiresConfig() && ev.ConfigContent != nil {
		configContent = ev.ConfigContent
	}

	stdout, stderr, exitCode, err := client.RunScript(script, configContent)
	if err != nil {
		return r.failJob(ctx, jobID, ev.ID, stdout, stderr+err.Error(), 1)
	}

	passed := exitCode == 0
	msg := "exit " + strconv.Itoa(exitCode)
	_ = r.store.AddJobCheck(ctx, jobID, models.CheckPost, "task_execution", passed, msg)

	status := models.JobVerified
	if !passed {
		status = models.JobFailed
	}
	if err := r.store.UpdateJobResult(ctx, jobID, status, stdout, stderr, exitCode); err != nil {
		return err
	}
	return r.store.RollupEvent(ctx, ev.ID)
}

func (r *Runner) fail(ctx context.Context, jobID, eventID, step string, err error) error {
	msg := step + ": " + err.Error()
	return r.failJob(ctx, jobID, eventID, "", msg, 1)
}

func (r *Runner) failJob(ctx context.Context, jobID, eventID, stdout, stderr string, code int) error {
	if err := r.store.UpdateJobResult(ctx, jobID, models.JobFailed, stdout, stderr, code); err != nil {
		return err
	}
	if err := r.store.RollupEvent(ctx, eventID); err != nil {
		return err
	}
	return nil
}
