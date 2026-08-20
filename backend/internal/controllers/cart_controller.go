package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/middleware"
	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
)

type CartController struct {
	service *services.CartService
}

func NewCartController(service *services.CartService) *CartController {
	return &CartController{service: service}
}

type cartItemRequest struct {
	ProductID primitive.ObjectID `json:"productId"`
	VariantID primitive.ObjectID `json:"variantId"`
	Quantity  int                `json:"quantity"`
}

func (h *CartController) Get(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	cart, err := h.service.Get(c.UserContext(), userID)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(fiber.Map{"cart": cart})
}

func (h *CartController) AddItem(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	var request cartItemRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cart, err := h.service.AddItem(c.UserContext(), userID, request.ProductID, request.VariantID, request.Quantity)
	if mapped := mapCartError(err); mapped != nil {
		return mapped
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"cart": cart})
}

func (h *CartController) UpdateItem(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	productID, variantID, err := cartItemIDs(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product or variant id")
	}
	var request struct {
		Quantity int `json:"quantity"`
	}
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cart, err := h.service.UpdateItem(c.UserContext(), userID, productID, variantID, request.Quantity)
	if mapped := mapCartError(err); mapped != nil {
		return mapped
	}
	return c.JSON(fiber.Map{"cart": cart})
}

func (h *CartController) RemoveItem(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	productID, variantID, err := cartItemIDs(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product or variant id")
	}
	if err := h.service.RemoveItem(c.UserContext(), userID, productID, variantID); err != nil {
		if mapped := mapCartError(err); mapped != nil {
			return mapped
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func authenticatedUserID(c *fiber.Ctx) (primitive.ObjectID, error) {
	id, ok := c.Locals(middleware.UserIDKey).(primitive.ObjectID)
	if !ok || id.IsZero() {
		return primitive.NilObjectID, errors.New("authenticated user is missing")
	}
	return id, nil
}

func cartItemIDs(c *fiber.Ctx) (primitive.ObjectID, primitive.ObjectID, error) {
	productID, err := primitive.ObjectIDFromHex(c.Params("productId"))
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	variantID, err := primitive.ObjectIDFromHex(c.Params("variantId"))
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	return productID, variantID, nil
}

func mapCartError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, services.ErrInvalidCartInput):
		return fiber.NewError(fiber.StatusBadRequest, "cart item data is invalid")
	case errors.Is(err, repositories.ErrProductNotFound):
		return fiber.ErrNotFound
	case errors.Is(err, services.ErrProductVariantNotFound):
		return fiber.NewError(fiber.StatusBadRequest, "product variant not found")
	case errors.Is(err, services.ErrInsufficientStock):
		return fiber.NewError(fiber.StatusConflict, "requested quantity is not available")
	case errors.Is(err, repositories.ErrCartItemNotFound):
		return fiber.ErrNotFound
	default:
		return fiber.ErrInternalServerError
	}
}
