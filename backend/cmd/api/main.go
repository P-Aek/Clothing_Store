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
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	categoryRepository := repositories.NewMongoCategoryRepository(dbClient.Database(cfg.MongoDatabase))
	categoryIndexCtx, cancelCategoryIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	if err := categoryRepository.EnsureIndexes(categoryIndexCtx); err != nil {
		cancelCategoryIndexes()
		log.Fatalf("ensure category indexes: %v", err)
	}
	cancelCategoryIndexes()
	productRepository := repositories.NewMongoProductRepository(dbClient.Database(cfg.MongoDatabase))
	productIndexCtx, cancelProductIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	if err := productRepository.EnsureIndexes(productIndexCtx); err != nil {
		cancelProductIndexes()
		log.Fatalf("ensure product indexes: %v", err)
	}
	cancelProductIndexes()
	cartRepository := repositories.NewMongoCartRepository(dbClient.Database(cfg.MongoDatabase))
	cartIndexCtx, cancelCartIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	if err := cartRepository.EnsureIndexes(cartIndexCtx); err != nil {
		cancelCartIndexes()
		log.Fatalf("ensure cart indexes: %v", err)
	}
	cancelCartIndexes()
	orderRepository := repositories.NewMongoOrderRepository(dbClient.Database(cfg.MongoDatabase))
	orderIndexCtx, cancelOrderIndexes := context.WithTimeout(context.Background(), 10*time.Second)
	if err := orderRepository.EnsureIndexes(orderIndexCtx); err != nil {
		cancelOrderIndexes()
		log.Fatalf("ensure order indexes: %v", err)
	}
	cancelOrderIndexes()
	authService := services.NewAuthService(userRepository, cfg.JWTSecret)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository, categoryRepository)
	cartService := services.NewCartService(cartRepository, productRepository)
	orderService := services.NewOrderService(orderRepository, cartRepository, productRepository)
	healthController := controllers.NewHealthController(dbClient)
	authController := controllers.NewAuthController(authService)
	categoryController := controllers.NewCategoryController(categoryService)
	productController := controllers.NewProductController(productService)
	cartController := controllers.NewCartController(cartService)
	orderController := controllers.NewOrderController(orderService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.FrontendURL,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip}\n",
	}))
	routes.Register(app, healthController, authController, categoryController, productController, cartController, orderController, cfg.JWTSecret)

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
