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
	orders                  map[primitive.ObjectID]models.Order
	stock                   map[primitive.ObjectID]int
	stockFailure            bool
	cancelFailureAfterItems int
}

func newMemoryOrderRepository() *memoryOrderRepository {
	return &memoryOrderRepository{
		orders: map[primitive.ObjectID]models.Order{},
		stock:  map[primitive.ObjectID]int{},
	}
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

func (r *memoryOrderRepository) ListByUserID(_ context.Context, userID primitive.ObjectID, page, limit int) (models.OrderListResponse, error) {
	orders := []models.Order{}
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return models.OrderListResponse{Orders: orders, Pagination: models.Pagination{Page: page, Limit: limit, TotalItems: int64(len(orders)), TotalPages: 1}}, nil
}

func (r *memoryOrderRepository) ListAll(_ context.Context, page, limit int) (models.OrderListResponse, error) {
	orders := make([]models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return models.OrderListResponse{Orders: orders, Pagination: models.Pagination{Page: page, Limit: limit, TotalItems: int64(len(orders)), TotalPages: 1}}, nil
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

func (r *memoryOrderRepository) Cancel(_ context.Context, id primitive.ObjectID, ownerID *primitive.ObjectID, updatedAt time.Time) (models.Order, error) {
	order, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Order{}, err
	}
	if ownerID != nil && order.UserID != *ownerID {
		return models.Order{}, repositories.ErrOrderNotOwned
	}
	if order.Status == models.OrderStatusCancelled {
		return models.Order{}, repositories.ErrOrderAlreadyCancelled
	}
	if order.Status != models.OrderStatusPending {
		return models.Order{}, repositories.ErrOrderCannotBeCancelled
	}

	// Copy state to model the commit-or-rollback behavior of the MongoDB transaction.
	nextStock := make(map[primitive.ObjectID]int, len(r.stock))
	for variantID, quantity := range r.stock {
		nextStock[variantID] = quantity
	}
	for index, item := range order.Items {
		nextStock[item.VariantID] += item.Quantity
		if r.cancelFailureAfterItems > 0 && index+1 == r.cancelFailureAfterItems {
			return models.Order{}, errors.New("stock restoration failed")
		}
	}

	order.Status = models.OrderStatusCancelled
	order.UpdatedAt = updatedAt
	r.stock = nextStock
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

func TestOrderServiceCancelsOwnPendingOrderAndRestoresStockOnce(t *testing.T) {
	ownerID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	order := models.Order{
		ID: primitive.NewObjectID(), UserID: ownerID, Status: models.OrderStatusPending,
		Items: []models.OrderItem{{ProductID: primitive.NewObjectID(), VariantID: variantID, Quantity: 2}},
	}
	orders := newMemoryOrderRepository()
	orders.orders[order.ID] = order
	orders.stock[variantID] = 8
	service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

	cancelled, err := service.CancelOrder(context.Background(), ownerID, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != models.OrderStatusCancelled || orders.stock[variantID] != 10 {
		t.Fatalf("order status = %q, stock = %d", cancelled.Status, orders.stock[variantID])
	}

	if _, err := service.CancelOrder(context.Background(), ownerID, order.ID); !errors.Is(err, repositories.ErrOrderAlreadyCancelled) {
		t.Fatalf("second cancellation error = %v", err)
	}
	if orders.stock[variantID] != 10 {
		t.Fatalf("stock after duplicate cancellation = %d, want 10", orders.stock[variantID])
	}
}

func TestOrderServiceRejectsCancellationByAnotherCustomer(t *testing.T) {
	ownerID := primitive.NewObjectID()
	order := models.Order{ID: primitive.NewObjectID(), UserID: ownerID, Status: models.OrderStatusPending}
	orders := newMemoryOrderRepository()
	orders.orders[order.ID] = order
	service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

	if _, err := service.CancelOrder(context.Background(), primitive.NewObjectID(), order.ID); !errors.Is(err, repositories.ErrOrderNotOwned) {
		t.Fatalf("ownership error = %v", err)
	}
	if orders.orders[order.ID].Status != models.OrderStatusPending {
		t.Fatalf("order status changed to %q", orders.orders[order.ID].Status)
	}
}

func TestOrderServiceRejectsNonPendingCancellation(t *testing.T) {
	for _, status := range []string{models.OrderStatusProcessing, models.OrderStatusShipped, models.OrderStatusDelivered} {
		t.Run(status, func(t *testing.T) {
			ownerID := primitive.NewObjectID()
			order := models.Order{ID: primitive.NewObjectID(), UserID: ownerID, Status: status}
			orders := newMemoryOrderRepository()
			orders.orders[order.ID] = order
			service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

			if _, err := service.CancelOrder(context.Background(), ownerID, order.ID); !errors.Is(err, repositories.ErrOrderCannotBeCancelled) {
				t.Fatalf("cancellation error = %v", err)
			}
			if orders.orders[order.ID].Status != status {
				t.Fatalf("order status changed to %q", orders.orders[order.ID].Status)
			}
		})
	}
}

func TestOrderServiceCancellationRollsBackWhenStockRestorationFails(t *testing.T) {
	ownerID := primitive.NewObjectID()
	firstVariantID := primitive.NewObjectID()
	secondVariantID := primitive.NewObjectID()
	order := models.Order{
		ID: primitive.NewObjectID(), UserID: ownerID, Status: models.OrderStatusPending,
		Items: []models.OrderItem{
			{ProductID: primitive.NewObjectID(), VariantID: firstVariantID, Quantity: 2},
			{ProductID: primitive.NewObjectID(), VariantID: secondVariantID, Quantity: 3},
		},
	}
	orders := newMemoryOrderRepository()
	orders.orders[order.ID] = order
	orders.stock[firstVariantID] = 8
	orders.stock[secondVariantID] = 7
	orders.cancelFailureAfterItems = 2
	service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

	if _, err := service.CancelOrder(context.Background(), ownerID, order.ID); err == nil {
		t.Fatal("expected stock restoration failure")
	}
	if orders.orders[order.ID].Status != models.OrderStatusPending {
		t.Fatalf("order status = %q, want pending", orders.orders[order.ID].Status)
	}
	if orders.stock[firstVariantID] != 8 || orders.stock[secondVariantID] != 7 {
		t.Fatalf("stock changed after rollback: first=%d second=%d", orders.stock[firstVariantID], orders.stock[secondVariantID])
	}
}

func TestOrderServiceAdminCancellationUsesTransactionalCancellation(t *testing.T) {
	ownerID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	order := models.Order{
		ID: primitive.NewObjectID(), UserID: ownerID, Status: models.OrderStatusPending,
		Items: []models.OrderItem{{ProductID: primitive.NewObjectID(), VariantID: variantID, Quantity: 2}},
	}
	orders := newMemoryOrderRepository()
	orders.orders[order.ID] = order
	orders.stock[variantID] = 8
	service := NewOrderService(orders, newMemoryCartRepository(), newMemoryProductRepository())

	updated, err := service.UpdateStatus(context.Background(), order.ID, models.OrderStatusCancelled)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.OrderStatusCancelled || orders.stock[variantID] != 10 {
		t.Fatalf("order status = %q, stock = %d", updated.Status, orders.stock[variantID])
	}
}
