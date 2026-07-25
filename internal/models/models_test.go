package models

import "testing"

func TestTaskType_RequiresConfig(t *testing.T) {
	if !TaskDeployConfig.RequiresConfig() {
		t.Fatal("deploy should require config")
	}
	if TaskSSHTest.RequiresConfig() {
		t.Fatal("ssh test should not require config")
	}
}
