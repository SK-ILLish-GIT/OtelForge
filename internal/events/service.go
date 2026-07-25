package events

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/models"
	"github.com/otelforge/otelforge/internal/queue"
	"github.com/otelforge/otelforge/internal/yamlutil"
)

type Service struct {
	store *db.Store
	queue *queue.Queue
}

func NewService(store *db.Store, q *queue.Queue) *Service {
	return &Service{store: store, queue: q}
}

type CreateInput struct {
	OwnerID       string
	Name          string
	LauncherName  string
	LauncherEmail string
	TaskType      models.TaskType
	ConfigContent *string
	InstanceIDs   []string
	RelatedEventID *string
	RollbackScope  *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*models.Event, error) {
	if len(in.InstanceIDs) == 0 {
		return nil, fmt.Errorf("at least one instance required")
	}
	if in.TaskType.RequiresConfig() {
		if in.ConfigContent == nil || *in.ConfigContent == "" {
			return nil, fmt.Errorf("config required for task")
		}
		if err := yamlutil.ValidateYAML([]byte(*in.ConfigContent)); err != nil {
			return nil, err
		}
	}
	for _, iid := range in.InstanceIDs {
		ok, err := s.store.UserOwnsInstance(ctx, in.OwnerID, iid)
		if err != nil || !ok {
			return nil, fmt.Errorf("instance not owned: %s", iid)
		}
	}

	var checksum *string
	if in.ConfigContent != nil {
		sum := sha256.Sum256([]byte(*in.ConfigContent))
		h := hex.EncodeToString(sum[:])
		checksum = &h
	}

	ev, jobs, err := s.store.CreateEventWithJobs(ctx, db.CreateEventParams{
		OwnerID:        in.OwnerID,
		Name:           in.Name,
		LauncherName:   in.LauncherName,
		LauncherEmail:  in.LauncherEmail,
		TaskType:       in.TaskType,
		ConfigContent:  in.ConfigContent,
		ConfigChecksum: checksum,
		RelatedEventID: in.RelatedEventID,
		RollbackScope:  in.RollbackScope,
		InstanceIDs:    in.InstanceIDs,
	})
	if err != nil {
		return nil, err
	}

	if err := s.store.UpdateEventStatus(ctx, ev.ID, models.EventRunning); err != nil {
		return nil, err
	}
	ev.Status = models.EventRunning

	for _, job := range jobs {
		if err := s.queue.Publish(ctx, queue.JobMessage{
			JobID:      job.ID,
			EventID:    job.EventID,
			InstanceID: job.InstanceID,
			TaskType:   string(in.TaskType),
		}); err != nil {
			return nil, err
		}
	}
	return ev, nil
}

func (s *Service) RerunFailed(ctx context.Context, eventID, ownerID string, admin bool) (*models.Event, error) {
	if !admin {
		ok, err := s.store.UserOwnsEvent(ctx, ownerID, eventID)
		if err != nil || !ok {
			return nil, fmt.Errorf("event not found")
		}
	}
	jobs, err := s.store.ResetFailedJobs(ctx, eventID)
	if err != nil {
		return nil, err
	}
	ev, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	for _, job := range jobs {
		if err := s.queue.Publish(ctx, queue.JobMessage{
			JobID:      job.ID,
			EventID:    job.EventID,
			InstanceID: job.InstanceID,
			TaskType:   string(ev.TaskType),
		}); err != nil {
			return nil, err
		}
	}
	return ev, nil
}

func (s *Service) Clone(ctx context.Context, sourceID, ownerID string, overrides CreateInput) (*models.Event, error) {
	src, err := s.store.GetEvent(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	ok, err := s.store.UserOwnsEvent(ctx, ownerID, sourceID)
	if err != nil || !ok {
		return nil, fmt.Errorf("event not found")
	}

	in := CreateInput{
		OwnerID:       ownerID,
		Name:          src.Name + " (clone)",
		LauncherName:  src.LauncherName,
		LauncherEmail: src.LauncherEmail,
		TaskType:      src.TaskType,
		ConfigContent: src.ConfigContent,
		InstanceIDs:   src.InstanceIDs,
	}
	if overrides.Name != "" {
		in.Name = overrides.Name
	}
	if overrides.LauncherName != "" {
		in.LauncherName = overrides.LauncherName
	}
	if overrides.LauncherEmail != "" {
		in.LauncherEmail = overrides.LauncherEmail
	}
	if overrides.ConfigContent != nil {
		in.ConfigContent = overrides.ConfigContent
	}
	if len(overrides.InstanceIDs) > 0 {
		in.InstanceIDs = overrides.InstanceIDs
	}
	return s.Create(ctx, in)
}
