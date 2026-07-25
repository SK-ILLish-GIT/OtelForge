package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/otelforge/otelforge/internal/auth"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/events"
	"github.com/otelforge/otelforge/internal/models"
)

type EventsHandler struct {
	store  *db.Store
	events *events.Service
}

func NewEventsHandler(store *db.Store, svc *events.Service) *EventsHandler {
	return &EventsHandler{store: store, events: svc}
}

type createEventRequest struct {
	Name          string   `json:"name"`
	LauncherName  string   `json:"launcherName"`
	LauncherEmail string   `json:"launcherEmail"`
	TaskType      string   `json:"taskType"`
	ConfigContent *string  `json:"configContent"`
	InstanceIDs   []string `json:"instanceIds"`
}

func (h *EventsHandler) Create(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	var req createEventRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid json")
	}
	if req.LauncherName == "" {
		req.LauncherName = user.Email
	}
	if req.LauncherEmail == "" {
		req.LauncherEmail = user.Email
	}
	ev, err := h.events.Create(c.Context(), events.CreateInput{
		OwnerID:       user.ID,
		Name:          req.Name,
		LauncherName:  req.LauncherName,
		LauncherEmail: req.LauncherEmail,
		TaskType:      models.TaskType(req.TaskType),
		ConfigContent: req.ConfigContent,
		InstanceIDs:   req.InstanceIDs,
	})
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}
	return writeJSON(c, fiber.StatusAccepted, ev)
}

func (h *EventsHandler) List(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	f := models.EventFilter{}
	if user.Role != models.RoleAdmin {
		f.OwnerID = user.ID
	} else {
		f.LauncherEmail = c.Query("launcherEmail")
		if s := c.Query("status"); s != "" {
			f.Status = models.EventStatus(s)
		}
		if t := c.Query("taskType"); t != "" {
			f.TaskType = models.TaskType(t)
		}
		if from := c.Query("from"); from != "" {
			if t, err := time.Parse(time.RFC3339, from); err == nil {
				f.From = &t
			}
		}
		if to := c.Query("to"); to != "" {
			if t, err := time.Parse(time.RFC3339, to); err == nil {
				f.To = &t
			}
		}
	}
	list, err := h.store.ListEvents(c.Context(), f)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	return writeJSON(c, fiber.StatusOK, list)
}

func (h *EventsHandler) Get(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	id := c.Params("id")
	ev, err := h.store.GetEvent(c.Context(), id)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, "not found")
	}
	if user.Role != models.RoleAdmin && ev.OwnerID != user.ID {
		return writeError(c, fiber.StatusForbidden, "forbidden")
	}
	jobs, err := h.store.ListJobsForEvent(c.Context(), id)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	ev.Jobs = jobs
	return writeJSON(c, fiber.StatusOK, ev)
}

func (h *EventsHandler) RerunFailed(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	id := c.Params("id")
	admin := user.Role == models.RoleAdmin
	ev, err := h.events.RerunFailed(c.Context(), id, user.ID, admin)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}
	return writeJSON(c, fiber.StatusOK, ev)
}

type cloneRequest struct {
	Name          string   `json:"name"`
	LauncherName  string   `json:"launcherName"`
	LauncherEmail string   `json:"launcherEmail"`
	ConfigContent *string  `json:"configContent"`
	InstanceIDs   []string `json:"instanceIds"`
}

func (h *EventsHandler) Clone(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	id := c.Params("id")
	var req cloneRequest
	_ = c.BodyParser(&req)
	ev, err := h.events.Clone(c.Context(), id, user.ID, events.CreateInput{
		Name: req.Name, LauncherName: req.LauncherName, LauncherEmail: req.LauncherEmail,
		ConfigContent: req.ConfigContent, InstanceIDs: req.InstanceIDs,
	})
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}
	return writeJSON(c, fiber.StatusAccepted, ev)
}

func (h *EventsHandler) RollbackJob(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	eventID := c.Params("id")
	jobID := c.Params("jobId")
	ev, err := h.store.GetEvent(c.Context(), eventID)
	if err != nil || (ev.OwnerID != user.ID && user.Role != models.RoleAdmin) {
		return writeError(c, fiber.StatusForbidden, "forbidden")
	}
	job, err := h.store.GetJob(c.Context(), jobID)
	if err != nil || job.EventID != eventID {
		return writeError(c, fiber.StatusNotFound, "not found")
	}
	scope := "instance"
	rollbackEv, err := h.events.Create(c.Context(), events.CreateInput{
		OwnerID: user.ID, Name: "Rollback " + ev.Name,
		LauncherName: user.Email, LauncherEmail: user.Email,
		TaskType: models.TaskRollback, InstanceIDs: []string{job.InstanceID},
		RelatedEventID: &eventID, RollbackScope: &scope,
	})
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}
	return writeJSON(c, fiber.StatusAccepted, rollbackEv)
}

type AdminHandler struct {
	store *db.Store
}

func NewAdminHandler(store *db.Store) *AdminHandler {
	return &AdminHandler{store: store}
}

func (h *AdminHandler) ListEvents(c *fiber.Ctx) error {
	f := models.EventFilter{
		LauncherEmail: c.Query("launcherEmail"),
	}
	if s := c.Query("status"); s != "" {
		f.Status = models.EventStatus(s)
	}
	if t := c.Query("taskType"); t != "" {
		f.TaskType = models.TaskType(t)
	}
	list, err := h.store.ListEvents(c.Context(), f)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	return writeJSON(c, fiber.StatusOK, list)
}

func (h *AdminHandler) ListInstances(c *fiber.Ctx) error {
	list, err := h.store.ListInstances(c.Context(), "", true)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	out := make([]instanceDTO, len(list))
	for i, inst := range list {
		out[i] = toInstanceDTO(inst)
	}
	return writeJSON(c, fiber.StatusOK, out)
}
