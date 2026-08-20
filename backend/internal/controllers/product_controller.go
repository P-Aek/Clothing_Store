package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
)

type ProductController struct {
	service *services.ProductService
}

func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{service: service}
}

type productVariantRequest struct {
	ID    primitive.ObjectID `json:"id,omitempty"`
	Color string             `json:"color"`
	Size  string             `json:"size"`
	Stock int                `json:"stock"`
}

type productRequest struct {
	CategoryID  primitive.ObjectID      `json:"categoryId"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Price       float64                 `json:"price"`
	Images      []string                `json:"images"`
	Variants    []productVariantRequest `json:"variants"`
}

func (h *ProductController) List(c *fiber.Ctx) error {
	var categoryID *primitive.ObjectID
	if value := c.Query("categoryId"); value != "" {
		id, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid category id")
		}
		categoryID = &id
	}
	products, err := h.service.List(c.UserContext(), categoryID)
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"products": products})
}

func (h *ProductController) Get(c *fiber.Ctx) error {
	id, err := productID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	product, err := h.service.Get(c.UserContext(), id)
	if errors.Is(err, repositories.ErrProductNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"product": product})
}

func (h *ProductController) Create(c *fiber.Ctx) error {
	var request productRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	product, err := h.service.Create(c.UserContext(), toProductInput(request))
	if errors.Is(err, services.ErrInvalidProductInput) {
		return fiber.NewError(fiber.StatusBadRequest, "product data is invalid")
	}
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.NewError(fiber.StatusBadRequest, "category not found")
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"product": product})
}

func (h *ProductController) Update(c *fiber.Ctx) error {
	id, err := productID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	var request productRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	product, err := h.service.Update(c.UserContext(), id, toProductInput(request))
	if errors.Is(err, services.ErrInvalidProductInput) {
		return fiber.NewError(fiber.StatusBadRequest, "product data is invalid")
	}
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return fiber.NewError(fiber.StatusBadRequest, "category not found")
	}
	if errors.Is(err, repositories.ErrProductNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"product": product})
}

func (h *ProductController) Delete(c *fiber.Ctx) error {
	id, err := productID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product id")
	}
	if err := h.service.Delete(c.UserContext(), id); errors.Is(err, repositories.ErrProductNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func toProductInput(request productRequest) services.ProductInput {
	variants := make([]services.ProductVariantInput, 0, len(request.Variants))
	for _, variant := range request.Variants {
		variants = append(variants, services.ProductVariantInput{ID: variant.ID, Color: variant.Color, Size: variant.Size, Stock: variant.Stock})
	}
	return services.ProductInput{
		CategoryID: request.CategoryID, Name: request.Name, Description: request.Description,
		Price: request.Price, Images: request.Images, Variants: variants,
	}
}

func productID(c *fiber.Ctx) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(c.Params("id"))
}
