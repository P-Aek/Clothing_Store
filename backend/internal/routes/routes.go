package routes

import (
	"github.com/gofiber/fiber/v2"

	"clothing-store-api/internal/controllers"
	"clothing-store-api/internal/middleware"
)

func Register(app *fiber.App, healthController *controllers.HealthController, authController *controllers.AuthController, jwtSecret string) {
	app.Get("/health", healthController.Health)

	authRoutes := app.Group("/api/auth")
	authRoutes.Post("/register", authController.Register)
	authRoutes.Post("/login", authController.Login)
	authRoutes.Get("/me", middleware.JWT(jwtSecret), authController.Me)
}
