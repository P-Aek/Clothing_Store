package routes

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/controllers"
	"clothing-store-api/internal/utils"
)

func authTestApp() *fiber.App {
	app := fiber.New()
	Register(app, nil, controllers.NewAuthController(nil), nil, nil, nil, nil, "test-secret")
	return app
}

func TestLogoutRequiresAuthentication(t *testing.T) {
	app := authTestApp()

	for _, authorization := range []string{"", "Bearer invalid"} {
		req := httptest.NewRequest("POST", "/api/auth/logout", nil)
		if authorization != "" {
			req.Header.Set(fiber.HeaderAuthorization, authorization)
		}

		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusUnauthorized)
		}
	}
}

func TestLogoutReturnsNoContentForAuthenticatedUser(t *testing.T) {
	app := authTestApp()
	token, err := utils.GenerateJWT(primitive.NewObjectID().Hex(), "customer", "test-secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
	if res.ContentLength > 0 {
		t.Fatalf("content length = %d, want an empty response", res.ContentLength)
	}
}
