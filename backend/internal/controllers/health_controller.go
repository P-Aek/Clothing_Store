package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoPinger interface {
	Ping(context.Context, *readpref.ReadPref) error
}

type HealthController struct {
	mongoClient mongoPinger
}

func NewHealthController(mongoClient *mongo.Client) *HealthController {
	return &HealthController{mongoClient: mongoClient}
}

func (h *HealthController) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "degraded",
			"database": "disconnected",
		})
	}

	return c.JSON(fiber.Map{
		"status":   "ok",
		"database": "connected",
	})
}
