package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/middleware"
	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthController) Register(c *fiber.Ctx) error {
	var request registerRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	created, err := h.service.Register(c.UserContext(), services.RegisterInput{
		Name: request.Name, Email: request.Email, Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			return fiber.NewError(fiber.StatusBadRequest, "name, email, or password is invalid")
		}
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return fiber.NewError(fiber.StatusConflict, "email is already registered")
		}
		return fiber.ErrInternalServerError
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": created})
}

func (h *AuthController) Login(c *fiber.Ctx) error {
	var request loginRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	token, loggedIn, err := h.service.Login(c.UserContext(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"token": token, "user": loggedIn})
}

func (h *AuthController) Me(c *fiber.Ctx) error {
	id, ok := c.Locals(middleware.UserIDKey).(primitive.ObjectID)
	if !ok || id.IsZero() {
		return fiber.ErrUnauthorized
	}
	role, _ := c.Locals(middleware.RoleKey).(string)
	return c.JSON(fiber.Map{"userId": id.Hex(), "role": role})
}

func (h *AuthController) Logout(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
