package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
)

type CategoryController struct {
	service *services.CategoryService
}

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

type categoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *CategoryController) List(c *fiber.Ctx) error {
	categories, err := h.service.List(c.UserContext())
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"categories": categories})
}

func (h *CategoryController) Get(c *fiber.Ctx) error {
	id, err := categoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
	}
	category, err := h.service.Get(c.UserContext(), id)
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"category": category})
}

func (h *CategoryController) Create(c *fiber.Ctx) error {
	var request categoryRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	category, err := h.service.Create(c.UserContext(), services.CategoryInput{Name: request.Name, Slug: request.Slug})
	if errors.Is(err, services.ErrInvalidCategoryInput) {
		return fiber.NewError(fiber.StatusBadRequest, "name or slug is invalid")
	}
	if errors.Is(err, repositories.ErrCategorySlugAlreadyExists) {
		return fiber.NewError(fiber.StatusConflict, "category slug already exists")
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"category": category})
}

func (h *CategoryController) Update(c *fiber.Ctx) error {
	id, err := categoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
	}
	var request categoryRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	category, err := h.service.Update(c.UserContext(), id, services.CategoryInput{Name: request.Name, Slug: request.Slug})
	if errors.Is(err, services.ErrInvalidCategoryInput) {
		return fiber.NewError(fiber.StatusBadRequest, "name or slug is invalid")
	}
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.ErrNotFound
	}
	if errors.Is(err, repositories.ErrCategorySlugAlreadyExists) {
		return fiber.NewError(fiber.StatusConflict, "category slug already exists")
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"category": category})
}

func (h *CategoryController) Delete(c *fiber.Ctx) error {
	id, err := categoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
	}
	if err := h.service.Delete(c.UserContext(), id); errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func categoryID(c *fiber.Ctx) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(c.Params("id"))
}
