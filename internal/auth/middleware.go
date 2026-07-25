package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/models"
)

const userLocalKey = "user"

type Middleware struct {
	tokens *TokenService
	store  *db.Store
}

func NewMiddleware(tokens *TokenService, store *db.Store) *Middleware {
	return &Middleware{tokens: tokens, store: store}
}

func (m *Middleware) RequireAuth(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
	}
	claims, err := m.tokens.ParseToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
	}
	user, err := m.store.GetUserByID(c.Context(), claims.Subject)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
	}
	c.Locals(userLocalKey, *user)
	return c.Next()
}

func (m *Middleware) RequireAdmin(c *fiber.Ctx) error {
	user, ok := UserFromFiber(c)
	if !ok || user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).SendString("forbidden")
	}
	return c.Next()
}

func UserFromFiber(c *fiber.Ctx) (models.User, bool) {
	user, ok := c.Locals(userLocalKey).(models.User)
	return user, ok
}

func CurrentUser(c *fiber.Ctx) models.User {
	user, _ := UserFromFiber(c)
	return user
}
