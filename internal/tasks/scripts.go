package tasks

import (
	"fmt"
	"io/fs"

	"github.com/otelforge/otelforge/internal/models"
	"github.com/otelforge/otelforge/scripts"
)

var scriptFiles = map[models.TaskType]string{
	models.TaskDeployConfig:   "bash/deploy.sh",
	models.TaskValidateConfig: "bash/validate.sh",
	models.TaskRestart:        "bash/restart.sh",
	models.TaskCheckStatus:    "bash/status.sh",
	models.TaskFetchLogs:      "bash/logs.sh",
	models.TaskRollback:       "bash/rollback.sh",
	models.TaskStop:           "bash/stop.sh",
	models.TaskSSHTest:        "bash/ssh_test.sh",
	models.TaskInstallAgent:   "bash/install.sh",
}

func ScriptFor(task models.TaskType) ([]byte, error) {
	name, ok := scriptFiles[task]
	if !ok {
		return nil, fmt.Errorf("unknown task type: %s", task)
	}
	return fs.ReadFile(scripts.FS, name)
}

func ScriptName(task models.TaskType) string {
	return scriptFiles[task]
}
