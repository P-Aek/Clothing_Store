package routes

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/controllers"
	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/services"
	"clothing-store-api/internal/utils"
)

type routeOrderRepository struct {
	orders map[primitive.ObjectID]models.Order
}

func newRouteOrderRepository() *routeOrderRepository {
	return &routeOrderRepository{orders: map[primitive.ObjectID]models.Order{}}
}

func (r *routeOrderRepository) CreateFromCart(_ context.Context, userID primitive.ObjectID, items []models.OrderItem, total float64) (models.Order, error) {
	order := models.Order{ID: primitive.NewObjectID(), UserID: userID, Items: items, TotalPrice: total, Status: models.OrderStatusPending}
	r.orders[order.ID] = order
	return order, nil
}
func (r *routeOrderRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return models.Order{}, repositories.ErrOrderNotFound
	}
	return order, nil
}
func (r *routeOrderRepository) ListByUserID(_ context.Context, userID primitive.ObjectID) ([]models.Order, error) {
	orders := []models.Order{}
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
func (r *routeOrderRepository) ListAll(context.Context) ([]models.Order, error) {
	orders := make([]models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return orders, nil
}
func (r *routeOrderRepository) UpdateStatus(_ context.Context, id primitive.ObjectID, status string, updatedAt time.Time) (models.Order, error) {
	order, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Order{}, err
	}
	order.Status = status
	order.UpdatedAt = updatedAt
	r.orders[id] = order
	return order, nil
}

func orderTestApp(orderController *controllers.OrderController) *fiber.App {
	app := fiber.New()
	Register(app, nil, nil, nil, nil, nil, orderController, "test-secret")
	return app
}

func orderToken(t *testing.T, userID primitive.ObjectID, role string) string {
	t.Helper()
	token, err := utils.GenerateJWT(userID.Hex(), role, "test-secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestOrderRoutesRequireAuthentication(t *testing.T) {
	service := services.NewOrderService(newRouteOrderRepository(), newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	res, err := orderTestApp(controllers.NewOrderController(service)).Test(httptest.NewRequest("GET", "/api/orders/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestOrderRoutesRestrictAdminListAndAllowUserList(t *testing.T) {
	userID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	userRequest := httptest.NewRequest("GET", "/api/orders/", nil)
	userRequest.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	userResponse, err := app.Test(userRequest)
	if err != nil || userResponse.StatusCode != fiber.StatusOK {
		t.Fatalf("user list status = %v, error = %v", userResponse.StatusCode, err)
	}

	adminRequest := httptest.NewRequest("GET", "/api/admin/orders/", nil)
	adminRequest.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	adminResponse, err := app.Test(adminRequest)
	if err != nil || adminResponse.StatusCode != fiber.StatusForbidden {
		t.Fatalf("admin list status = %v, error = %v", adminResponse.StatusCode, err)
	}
}

func TestOrderRoutesAllowAdminStatusUpdate(t *testing.T) {
	orderID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	orders.orders[orderID] = models.Order{ID: orderID, UserID: primitive.NewObjectID(), Status: models.OrderStatusPending}
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	req := httptest.NewRequest("PUT", "/api/admin/orders/"+orderID.Hex()+"/status", strings.NewReader(`{"status":"shipped"}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, primitive.NewObjectID(), "admin"))
	res, err := app.Test(req)
	if err != nil || res.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %v, error = %v", res.StatusCode, err)
	}
}
