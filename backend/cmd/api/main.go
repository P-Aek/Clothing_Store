package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"clothing-store-api/internal/config"
	"clothing-store-api/internal/database"
)

func main() {
	// Loading .env is convenient locally; deployed environments provide variables directly.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load environment: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	dbClient, err := database.Connect(context.Background(), cfg.MongoURI)
	if err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := dbClient.Disconnect(closeCtx); err != nil {
			log.Printf("disconnect MongoDB: %v", err)
		}
	}()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("API listening on port %s", cfg.Port)
		serverErrors <- app.Listen(":" + cfg.Port)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("start API: %v", err)
		}
	case <-shutdown:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("shutdown API: %v", err)
		}
	}
}
