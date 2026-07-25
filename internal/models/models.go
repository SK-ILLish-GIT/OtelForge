package models

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type TaskType string

const (
	TaskDeployConfig   TaskType = "deploy_config"
	TaskValidateConfig TaskType = "validate_config"
	TaskRestart        TaskType = "restart_collector"
	TaskCheckStatus    TaskType = "check_status"
	TaskFetchLogs      TaskType = "fetch_logs"
	TaskRollback       TaskType = "rollback_config"
	TaskStop           TaskType = "stop_collector"
	TaskSSHTest        TaskType = "ssh_connectivity_test"
	TaskInstallAgent   TaskType = "install_otel_agent"
)

func (t TaskType) RequiresConfig() bool {
	return t == TaskDeployConfig || t == TaskValidateConfig
}

func AllTaskTypes() []TaskType {
	return []TaskType{
		TaskDeployConfig, TaskValidateConfig, TaskRestart, TaskCheckStatus,
		TaskFetchLogs, TaskRollback, TaskStop, TaskSSHTest, TaskInstallAgent,
	}
}

type EventStatus string

const (
	EventQueued    EventStatus = "QUEUED"
	EventRunning   EventStatus = "RUNNING"
	EventCompleted EventStatus = "COMPLETED"
	EventPartial   EventStatus = "PARTIAL"
	EventFailed    EventStatus = "FAILED"
)

type JobStatus string

const (
	JobQueued   JobStatus = "QUEUED"
	JobRunning  JobStatus = "RUNNING"
	JobVerified JobStatus = "VERIFIED"
	JobFailed   JobStatus = "FAILED"
)

type CheckPhase string

const (
	CheckPre  CheckPhase = "pre"
	CheckPost CheckPhase = "post"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Instance struct {
	ID               string    `json:"id"`
	OwnerID          string    `json:"ownerId"`
	OwnerEmail       string    `json:"ownerEmail,omitempty"`
	Name             string    `json:"name"`
	Host             string    `json:"host"`
	Port             int       `json:"port"`
	SSHUser          string    `json:"sshUser"`
	SSHPasswordEnc          []byte    `json:"-"`
	SSHPrivateKeyEnc        []byte    `json:"-"`
	SSHHostKeyFingerprint   *string   `json:"-"`
	CreatedAt               time.Time `json:"createdAt"`
}

type Event struct {
	ID             string      `json:"id"`
	OwnerID        string      `json:"ownerId"`
	Name           string      `json:"name"`
	LauncherName   string      `json:"launcherName"`
	LauncherEmail  string      `json:"launcherEmail"`
	TaskType       TaskType    `json:"taskType"`
	ConfigContent  *string     `json:"configContent,omitempty"`
	ConfigChecksum *string     `json:"configChecksum,omitempty"`
	Status         EventStatus `json:"status"`
	RelatedEventID *string     `json:"relatedEventId,omitempty"`
	RollbackScope  *string     `json:"rollbackScope,omitempty"`
	TotalJobs      int         `json:"totalJobs"`
	VerifiedCount  int         `json:"verifiedCount"`
	FailedCount    int         `json:"failedCount"`
	CreatedAt      time.Time   `json:"createdAt"`
	InstanceIDs    []string    `json:"instanceIds,omitempty"`
	Jobs           []Job       `json:"jobs,omitempty"`
}

type Job struct {
	ID           string     `json:"id"`
	EventID      string     `json:"eventId"`
	InstanceID   string     `json:"instanceId"`
	InstanceName string     `json:"instanceName,omitempty"`
	InstanceHost string     `json:"instanceHost,omitempty"`
	InstancePort int        `json:"instancePort,omitempty"`
	Status       JobStatus  `json:"status"`
	Stdout     *string    `json:"stdout,omitempty"`
	Stderr     *string    `json:"stderr,omitempty"`
	ExitCode   *int       `json:"exitCode,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Checks     []JobCheck `json:"checks,omitempty"`
}

type JobCheck struct {
	ID      string     `json:"id"`
	JobID   string     `json:"jobId"`
	Phase   CheckPhase `json:"phase"`
	Name    string     `json:"name"`
	Passed  bool       `json:"passed"`
	Message string     `json:"message"`
}

type EventFilter struct {
	OwnerID       string
	LauncherEmail string
	Status        EventStatus
	TaskType      TaskType
	From          *time.Time
	To            *time.Time
}
