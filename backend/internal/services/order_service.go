package services

import (
	"context"
	"errors"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

var (
	ErrCartEmpty          = errors.New("cart is empty")
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

type OrderService struct {
	orders   repositories.OrderRepository
	carts    repositories.CartRepository
	products repositories.ProductRepository
	now      func() time.Time
}

func NewOrderService(orders repositories.OrderRepository, carts repositories.CartRepository, products repositories.ProductRepository) *OrderService {
	return &OrderService{orders: orders, carts: carts, products: products, now: time.Now}
}

func (s *OrderService) Create(ctx context.Context, userID primitive.ObjectID) (models.Order, error) {
	if userID.IsZero() {
		return models.Order{}, ErrInvalidInput
	}
	cart, err := s.carts.FindByUserID(ctx, userID)
	if errors.Is(err, repositories.ErrCartNotFound) {
		return models.Order{}, ErrCartEmpty
	}
	if err != nil {
		return models.Order{}, err
	}
	if len(cart.Items) == 0 {
		return models.Order{}, ErrCartEmpty
	}

	items := make([]models.OrderItem, 0, len(cart.Items))
	var total float64
	for _, cartItem := range cart.Items {
		if cartItem.Quantity < 1 {
			return models.Order{}, ErrInvalidInput
		}
		product, err := s.products.FindByID(ctx, cartItem.ProductID)
		if err != nil {
			return models.Order{}, err
		}
		var variant models.ProductVariant
		for _, candidate := range product.Variants {
			if candidate.ID == cartItem.VariantID {
				variant = candidate
				break
			}
		}
		if variant.ID.IsZero() {
			return models.Order{}, ErrProductVariantNotFound
		}
		subtotal := math.Round(product.Price*float64(cartItem.Quantity)*100) / 100
		item := models.OrderItem{
			ProductID: product.ID, VariantID: variant.ID, ProductName: product.Name,
			Color: variant.Color, Size: variant.Size, Price: product.Price,
			Quantity: cartItem.Quantity, Subtotal: subtotal,
		}
		items = append(items, item)
		total += subtotal
	}
	return s.orders.CreateFromCart(ctx, userID, items, math.Round(total*100)/100)
}

func (s *OrderService) ListByUser(ctx context.Context, userID primitive.ObjectID, page, limit int) (models.OrderListResponse, error) {
	if userID.IsZero() {
		return models.OrderListResponse{}, ErrInvalidInput
	}
	page, limit = normalizePagination(page, limit)
	return s.orders.ListByUserID(ctx, userID, page, limit)
}

func (s *OrderService) GetForUser(ctx context.Context, userID, orderID primitive.ObjectID) (models.Order, error) {
	if userID.IsZero() || orderID.IsZero() {
		return models.Order{}, ErrInvalidInput
	}
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}
	if order.UserID != userID {
		return models.Order{}, repositories.ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID primitive.ObjectID) (models.Order, error) {
	if userID.IsZero() || orderID.IsZero() {
		return models.Order{}, ErrInvalidInput
	}
	return s.orders.Cancel(ctx, orderID, &userID, s.now().UTC())
}

func (s *OrderService) ListAll(ctx context.Context, page, limit int) (models.OrderListResponse, error) {
	page, limit = normalizePagination(page, limit)
	return s.orders.ListAll(ctx, page, limit)
}

func normalizePagination(page, limit int) (int, int) {
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

func (s *OrderService) UpdateStatus(ctx context.Context, orderID primitive.ObjectID, status string) (models.Order, error) {
	if orderID.IsZero() {
		return models.Order{}, ErrInvalidInput
	}
	if !validOrderStatus(status) {
		return models.Order{}, ErrInvalidOrderStatus
	}
	if status == models.OrderStatusCancelled {
		return s.orders.Cancel(ctx, orderID, nil, s.now().UTC())
	}
	return s.orders.UpdateStatus(ctx, orderID, status, s.now().UTC())
}

func validOrderStatus(status string) bool {
	switch status {
	case models.OrderStatusPending, models.OrderStatusProcessing, models.OrderStatusShipped, models.OrderStatusDelivered, models.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
