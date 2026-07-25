package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/otelforge/otelforge/internal/auth"
	"github.com/otelforge/otelforge/internal/crypto"
	"github.com/otelforge/otelforge/internal/db"
)

type AuthHandler struct {
	store  *db.Store
	tokens *auth.TokenService
}

func NewAuthHandler(store *db.Store, tokens *auth.TokenService) *AuthHandler {
	return &AuthHandler{store: store, tokens: tokens}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type userDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid json")
	}
	user, err := h.store.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "invalid credentials")
	}
	if err := crypto.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return writeError(c, fiber.StatusUnauthorized, "invalid credentials")
	}
	token, err := h.tokens.IssueToken(user)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "token error")
	}
	return writeJSON(c, fiber.StatusOK, loginResponse{
		Token: token,
		User:  userDTO{ID: user.ID, Email: user.Email, Role: string(user.Role)},
	})
}
