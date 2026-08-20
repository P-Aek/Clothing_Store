package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/utils"
)

const (
	ClaimsKey = "auth_claims"
	UserIDKey = "authenticated_user_id"
	RoleKey   = "authenticated_user_role"
)

func JWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const unauthorized = "unauthorized"
		header := strings.SplitN(c.Get(fiber.HeaderAuthorization), " ", 2)
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") || header[1] == "" {
			return fiber.NewError(fiber.StatusUnauthorized, unauthorized)
		}
		claims, err := utils.ParseJWT(header[1], secret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, unauthorized)
		}
		id, err := primitive.ObjectIDFromHex(claims.Subject)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, unauthorized)
		}
		c.Locals(ClaimsKey, claims)
		c.Locals(UserIDKey, id)
		c.Locals(RoleKey, claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(RoleKey).(string)
		if !ok {
			return fiber.ErrUnauthorized
		}
		if _, ok := allowed[role]; !ok {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}
