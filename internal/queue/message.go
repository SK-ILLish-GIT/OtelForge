package queue

type JobMessage struct {
	JobID      string `json:"jobId"`
	EventID    string `json:"eventId"`
	InstanceID string `json:"instanceId"`
	TaskType   string `json:"taskType"`
}
