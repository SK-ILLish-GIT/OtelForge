package handlers

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/otelforge/otelforge/internal/auth"
	"github.com/otelforge/otelforge/internal/crypto"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/models"
)

type InstancesHandler struct {
	store         *db.Store
	encryptionKey string
}

func NewInstancesHandler(store *db.Store, encryptionKey string) *InstancesHandler {
	return &InstancesHandler{store: store, encryptionKey: encryptionKey}
}

type instanceDTO struct {
	ID         string `json:"id"`
	OwnerID    string `json:"ownerId,omitempty"`
	OwnerEmail string `json:"ownerEmail,omitempty"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	SSHUser    string `json:"sshUser"`
	CreatedAt  string `json:"createdAt"`
}

func toInstanceDTO(i models.Instance) instanceDTO {
	return instanceDTO{
		ID: i.ID, OwnerID: i.OwnerID, OwnerEmail: i.OwnerEmail,
		Name: i.Name, Host: i.Host, Port: i.Port, SSHUser: i.SSHUser,
		CreatedAt: i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *InstancesHandler) List(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	admin := user.Role == models.RoleAdmin
	list, err := h.store.ListInstances(c.Context(), user.ID, admin)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	out := make([]instanceDTO, len(list))
	for i, inst := range list {
		out[i] = toInstanceDTO(inst)
	}
	return writeJSON(c, fiber.StatusOK, out)
}

type createInstanceRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	SSHUser    string `json:"sshUser"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

func (h *InstancesHandler) encryptCredential(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	return crypto.Encrypt([]byte(value), h.encryptionKey)
}

func (h *InstancesHandler) Create(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	var req createInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid json")
	}
	if req.Port == 0 {
		req.Port = 22
	}
	if req.Password == "" && req.PrivateKey == "" {
		return writeError(c, fiber.StatusBadRequest, "password or privateKey required")
	}
	passEnc, err := h.encryptCredential(req.Password)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	keyEnc, err := h.encryptCredential(req.PrivateKey)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	inst, err := h.store.CreateInstance(c.Context(), user.ID, req.Name, req.Host, req.Port, req.SSHUser, passEnc, keyEnc)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	return writeJSON(c, fiber.StatusCreated, toInstanceDTO(*inst))
}

type updateInstanceRequest struct {
	SSHUser    string `json:"sshUser"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

func (h *InstancesHandler) Update(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	id := c.Params("id")
	var req updateInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid json")
	}
	if req.SSHUser == "" {
		return writeError(c, fiber.StatusBadRequest, "sshUser required")
	}
	if req.Password == "" && req.PrivateKey == "" {
		return writeError(c, fiber.StatusBadRequest, "password or privateKey required")
	}
	passEnc, err := h.encryptCredential(req.Password)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	keyEnc, err := h.encryptCredential(req.PrivateKey)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err.Error())
	}
	inst, err := h.store.UpdateInstanceSSH(c.Context(), id, user.ID, req.SSHUser, passEnc, keyEnc)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err.Error())
	}
	return writeJSON(c, fiber.StatusOK, toInstanceDTO(*inst))
}

type bulkRequest struct {
	CSV string `json:"csv"`
}

func (h *InstancesHandler) BulkCreate(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	var req bulkRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid json")
	}
	reader := csv.NewReader(strings.NewReader(req.CSV))
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		return writeError(c, fiber.StatusBadRequest, "invalid csv")
	}
	var created []instanceDTO
	var errs []string
	for _, row := range rows[1:] {
		if len(row) < 5 {
			errs = append(errs, "row too short")
			continue
		}
		port, _ := strconv.Atoi(row[2])
		if port == 0 {
			port = 22
		}
		enc, err := crypto.Encrypt([]byte(row[4]), h.encryptionKey)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		inst, err := h.store.CreateInstance(c.Context(), user.ID, row[0], row[1], port, row[3], enc, nil)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		created = append(created, toInstanceDTO(*inst))
	}
	return writeJSON(c, fiber.StatusOK, map[string]any{"created": len(created), "instances": created, "errors": errs})
}

func (h *InstancesHandler) Delete(c *fiber.Ctx) error {
	user := auth.CurrentUser(c)
	id := c.Params("id")
	if err := h.store.DeleteInstance(c.Context(), id, user.ID); err != nil {
		return writeError(c, fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
