package controllers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type fakeMongoPinger struct {
	err error
}

func (f fakeMongoPinger) Ping(context.Context, *readpref.ReadPref) error {
	return f.err
}

func TestHealthReturnsConnectedWhenMongoDBIsAvailable(t *testing.T) {
	controller := &HealthController{mongoClient: fakeMongoPinger{}}
	app := fiber.New()
	app.Get("/health", controller.Health)

	response, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, response.StatusCode)
	}
}

func TestHealthReturnsUnavailableWhenMongoDBIsDown(t *testing.T) {
	controller := &HealthController{mongoClient: fakeMongoPinger{err: errors.New("database unavailable")}}
	app := fiber.New()
	app.Get("/health", controller.Health)

	response, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", fiber.StatusServiceUnavailable, response.StatusCode)
	}
}
