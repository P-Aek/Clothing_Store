package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

type memoryOrderRepository struct {
	orders       map[primitive.ObjectID]models.Order
	stockFailure bool
}

func newMemoryOrderRepository() *memoryOrderRepository {
	return &memoryOrderRepository{orders: map[primitive.ObjectID]models.Order{}}
}

func (r *memoryOrderRepository) CreateFromCart(_ context.Context, userID primitive.ObjectID, items []models.OrderItem, total float64) (models.Order, error) {
	if r.stockFailure {
		return models.Order{}, repositories.ErrOrderStockUnavailable
	}
	order := models.Order{ID: primitive.NewObjectID(), UserID: userID, Items: items, TotalPrice: total, Status: models.OrderStatusPending}
	r.orders[order.ID] = order
	return order, nil
}

func (r *memoryOrderRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return models.Order{}, repositories.ErrOrderNotFound
	}
	return order, nil
}

func (r *memoryOrderRepository) ListByUserID(_ context.Context, userID primitive.ObjectID) ([]models.Order, error) {
	orders := []models.Order{}
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (r *memoryOrderRepository) ListAll(context.Context) ([]models.Order, error) {
	orders := make([]models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *memoryOrderRepository) UpdateStatus(_ context.Context, id primitive.ObjectID, status string, updatedAt time.Time) (models.Order, error) {
	order, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Order{}, err
	}
	order.Status = status
	order.UpdatedAt = updatedAt
	r.orders[id] = order
	return order, nil
}

func TestOrderServiceCreatesSnapshotAndTotalFromCart(t *testing.T) {
	userID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	products := newMemoryProductRepository()
	products.products[productID] = models.Product{
		ID: productID, Name: "Blue Shirt", Price: 25.50, Active: true,
		Variants: []models.ProductVariant{{ID: variantID, Color: "Blue", Size: "M", Stock: 4}},
	}
	carts := newMemoryCartRepository()
	carts.carts[userID] = models.Cart{UserID: userID, Items: []models.CartItem{{ProductID: productID, VariantID: variantID, Quantity: 2}}}
	orders := newMemoryOrderRepository()
	service := NewOrderService(orders, carts, products)

	order, err := service.Create(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != models.OrderStatusPending || order.TotalPrice != 51 || len(order.Items) != 1 {
		t.Fatalf("order = %+v", order)
	}
	item := order.Items[0]
	if item.ProductName != "Blue Shirt" || item.Color != "Blue" || item.Size != "M" || item.Subtotal != 51 {
		t.Fatalf("order item = %+v", item)
	}
}

func TestOrderServiceRejectsEmptyCartAndStockFailure(t *testing.T) {
	userID := primitive.NewObjectID()
	carts := newMemoryCartRepository()
	orders := newMemoryOrderRepository()
	products := newMemoryProductRepository()
	service := NewOrderService(orders, carts, products)

	if _, err := service.Create(context.Background(), userID); !errors.Is(err, ErrCartEmpty) {
		t.Fatalf("empty cart error = %v", err)
	}

	productID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	products.products[productID] = models.Product{ID: productID, Name: "Shirt", Price: 10, Active: true, Variants: []models.ProductVariant{{ID: variantID, Stock: 1}}}
	carts.carts[userID] = models.Cart{UserID: userID, Items: []models.CartItem{{ProductID: productID, VariantID: variantID, Quantity: 1}}}
	orders.stockFailure = true
	if _, err := service.Create(context.Background(), userID); !errors.Is(err, repositories.ErrOrderStockUnavailable) {
		t.Fatalf("stock error = %v", err)
	}
}

func TestOrderServiceProtectsOwnershipAndValidatesStatus(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	order := models.Order{ID: primitive.NewObjectID(), UserID: ownerID, Status: models.OrderStatusPending}
	orders := newMemoryOrderRepository()
	orders.orders[order.ID] = order
	service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

	if _, err := service.GetForUser(context.Background(), otherID, order.ID); !errors.Is(err, repositories.ErrOrderNotFound) {
		t.Fatalf("ownership error = %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), order.ID, "unknown"); !errors.Is(err, ErrInvalidOrderStatus) {
		t.Fatalf("status error = %v", err)
	}
	updated, err := service.UpdateStatus(context.Background(), order.ID, models.OrderStatusShipped)
	if err != nil || updated.Status != models.OrderStatusShipped {
		t.Fatalf("updated order = %+v, error = %v", updated, err)
	}
}
