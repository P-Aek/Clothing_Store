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
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"clothing-store-api/internal/config"
	"clothing-store-api/internal/controllers"
	"clothing-store-api/internal/database"
	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/routes"
	"clothing-store-api/internal/services"
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
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := dbClient.Disconnect(closeCtx); err != nil {
			log.Printf("disconnect MongoDB: %v", err)
		}
	}()

	userRepository := repositories.NewMongoUserRepository(dbClient.Database(cfg.MongoDatabase))
	indexCtx, cancelIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	if err := userRepository.EnsureIndexes(indexCtx); err != nil {
		cancelIndexes()
		log.Fatalf("ensure user indexes: %v", err)
	}
	cancelIndexes()
	authService := services.NewAuthService(userRepository, cfg.JWTSecret)
	healthController := controllers.NewHealthController(dbClient)
	authController := controllers.NewAuthController(authService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip}\n",
	}))
	routes.Register(app, healthController, authController, cfg.JWTSecret)

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
