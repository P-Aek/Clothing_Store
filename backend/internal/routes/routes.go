package routes

import (
	"github.com/gofiber/fiber/v2"

	"clothing-store-api/internal/controllers"
	"clothing-store-api/internal/middleware"
)

func Register(app *fiber.App, healthController *controllers.HealthController, authController *controllers.AuthController, categoryController *controllers.CategoryController, jwtSecret string) {
	app.Get("/health", healthController.Health)

	authRoutes := app.Group("/api/auth")
	authRoutes.Post("/register", authController.Register)
	authRoutes.Post("/login", authController.Login)
	authRoutes.Get("/me", middleware.JWT(jwtSecret), authController.Me)

	categoryRoutes := app.Group("/api/categories")
	categoryRoutes.Get("/", categoryController.List)
	categoryRoutes.Get("/:id", categoryController.Get)
	adminCategoryRoutes := categoryRoutes.Group("", middleware.JWT(jwtSecret), middleware.RequireRole("admin"))
	adminCategoryRoutes.Post("/", categoryController.Create)
	adminCategoryRoutes.Put("/:id", categoryController.Update)
	adminCategoryRoutes.Delete("/:id", categoryController.Delete)
}
