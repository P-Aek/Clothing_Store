package controllers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
)

type OrderController struct {
	service *services.OrderService
}

func NewOrderController(service *services.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (h *OrderController) Create(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	order, err := h.service.Create(c.UserContext(), userID)
	if mapped := mapOrderError(err); mapped != nil {
		return mapped
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"order": order})
}

func (h *OrderController) List(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	page, limit := paginationQuery(c)
	orders, err := h.service.ListByUser(c.UserContext(), userID, page, limit)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(orders)
}

func (h *OrderController) Get(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	orderID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
	}
	order, err := h.service.GetForUser(c.UserContext(), userID, orderID)
	if mapped := mapOrderError(err); mapped != nil {
		return mapped
	}
	return c.JSON(fiber.Map{"order": order})
}

func (h *OrderController) Cancel(c *fiber.Ctx) error {
	userID, err := authenticatedUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	orderID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
	}
	order, err := h.service.CancelOrder(c.UserContext(), userID, orderID)
	if mapped := mapOrderError(err); mapped != nil {
		return mapped
	}
	return c.JSON(fiber.Map{"message": "order cancelled successfully", "order": order})
}

func (h *OrderController) AdminList(c *fiber.Ctx) error {
	page, limit := paginationQuery(c)
	orders, err := h.service.ListAll(c.UserContext(), page, limit)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(orders)
}

func paginationQuery(c *fiber.Ctx) (int, int) {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func (h *OrderController) AdminUpdateStatus(c *fiber.Ctx) error {
	orderID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
	}
	var request struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	order, err := h.service.UpdateStatus(c.UserContext(), orderID, strings.ToLower(strings.TrimSpace(request.Status)))
	if mapped := mapOrderError(err); mapped != nil {
		return mapped
	}
	return c.JSON(fiber.Map{"order": order})
}

func mapOrderError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, services.ErrCartEmpty):
		return fiber.NewError(fiber.StatusConflict, "cart is empty")
	case errors.Is(err, services.ErrProductVariantNotFound):
		return fiber.NewError(fiber.StatusConflict, "a product variant is no longer available")
	case errors.Is(err, repositories.ErrOrderStockUnavailable):
		return fiber.NewError(fiber.StatusConflict, "requested stock is no longer available")
	case errors.Is(err, repositories.ErrOrderNotOwned):
		return fiber.NewError(fiber.StatusForbidden, "order does not belong to authenticated user")
	case errors.Is(err, repositories.ErrOrderAlreadyCancelled):
		return fiber.NewError(fiber.StatusBadRequest, "order already cancelled")
	case errors.Is(err, repositories.ErrOrderCannotBeCancelled):
		return fiber.NewError(fiber.StatusBadRequest, "order cannot be cancelled in its current status")
	case errors.Is(err, repositories.ErrCartChanged):
		return fiber.NewError(fiber.StatusConflict, "cart changed during checkout; please try again")
	case errors.Is(err, services.ErrInvalidOrderStatus):
		return fiber.NewError(fiber.StatusBadRequest, "order status is invalid")
	case errors.Is(err, repositories.ErrOrderNotFound):
		return fiber.ErrNotFound
	case errors.Is(err, repositories.ErrProductNotFound):
		return fiber.NewError(fiber.StatusConflict, "a product in the cart is no longer available")
	case errors.Is(err, services.ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "order data is invalid")
	default:
		return fiber.ErrInternalServerError
	}
}
