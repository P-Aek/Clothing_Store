package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"clothing-store-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func timeNow() time.Time { return time.Now() }

func TestJWTRejectsMissingOrInvalidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", JWT("secret"), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	for _, header := range []string{"", "Basic abc", "Bearer invalid"} {
		req := httptest.NewRequest("GET", "/protected", nil)
		if header != "" {
			req.Header.Set(fiber.HeaderAuthorization, header)
		}
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("status = %d", res.StatusCode)
		}
	}
}

func TestJWTAndRequireRole(t *testing.T) {
	token, err := utils.GenerateJWT("507f1f77bcf86cd799439011", "admin", "secret", timeNow())
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Get("/admin", JWT("secret"), RequireRole("admin"), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d", res.StatusCode)
	}
}
