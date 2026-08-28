package routes

import (
	"context"
	"encoding/json"
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
	orders    map[primitive.ObjectID]models.Order
	lastPage  int
	lastLimit int
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
func (r *routeOrderRepository) ListByUserID(_ context.Context, userID primitive.ObjectID, page, limit int) (models.OrderListResponse, error) {
	r.lastPage, r.lastLimit = page, limit
	orders := []models.Order{}
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return models.OrderListResponse{Orders: orders, Pagination: models.Pagination{Page: page, Limit: limit, TotalItems: int64(len(orders)), TotalPages: 1}}, nil
}
func (r *routeOrderRepository) ListAll(_ context.Context, page, limit int) (models.OrderListResponse, error) {
	r.lastPage, r.lastLimit = page, limit
	orders := make([]models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return models.OrderListResponse{Orders: orders, Pagination: models.Pagination{Page: page, Limit: limit, TotalItems: int64(len(orders)), TotalPages: 1}}, nil
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

func (r *routeOrderRepository) Cancel(_ context.Context, id primitive.ObjectID, ownerID *primitive.ObjectID, updatedAt time.Time) (models.Order, error) {
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
	order.Status = models.OrderStatusCancelled
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

func TestOrderListUsesPaginationAndOnlyReturnsAuthenticatedUserOrders(t *testing.T) {
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	orders.orders[primitive.NewObjectID()] = models.Order{UserID: userID}
	orders.orders[primitive.NewObjectID()] = models.Order{UserID: otherUserID}
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	req := httptest.NewRequest("GET", "/api/orders?page=2&limit=5", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	res, err := app.Test(req)
	if err != nil || res.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %v, error = %v", res.StatusCode, err)
	}
	var response models.OrderListResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if orders.lastPage != 2 || orders.lastLimit != 5 {
		t.Fatalf("repository pagination = (%d, %d), want (2, 5)", orders.lastPage, orders.lastLimit)
	}
	if len(response.Orders) != 1 || response.Orders[0].UserID != userID {
		t.Fatalf("orders = %+v, want only authenticated user's order", response.Orders)
	}
	if response.Pagination.Page != 2 || response.Pagination.Limit != 5 {
		t.Fatalf("pagination = %+v", response.Pagination)
	}
}

func TestOrderListNormalizesPaginationQuery(t *testing.T) {
	userID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	req := httptest.NewRequest("GET", "/api/orders?page=0&limit=101", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	res, err := app.Test(req)
	if err != nil || res.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %v, error = %v", res.StatusCode, err)
	}
	if orders.lastPage != 1 || orders.lastLimit != 100 {
		t.Fatalf("repository pagination = (%d, %d), want (1, 100)", orders.lastPage, orders.lastLimit)
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

func TestOrderRoutesAllowCustomerToCancelOwnPendingOrder(t *testing.T) {
	userID := primitive.NewObjectID()
	orderID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	orders.orders[orderID] = models.Order{ID: orderID, UserID: userID, Status: models.OrderStatusPending}
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	req := httptest.NewRequest("PUT", "/api/orders/"+orderID.Hex()+"/cancel", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	res, err := app.Test(req)
	if err != nil || res.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %v, error = %v", res.StatusCode, err)
	}
	var response struct {
		Message string       `json:"message"`
		Order   models.Order `json:"order"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Message != "order cancelled successfully" || response.Order.Status != models.OrderStatusCancelled {
		t.Fatalf("response = %+v", response)
	}
}

func TestOrderCancelRouteEnforcesAuthenticationOwnershipAndStatus(t *testing.T) {
	ownerID := primitive.NewObjectID()
	orderID := primitive.NewObjectID()
	orders := newRouteOrderRepository()
	orders.orders[orderID] = models.Order{ID: orderID, UserID: ownerID, Status: models.OrderStatusPending}
	service := services.NewOrderService(orders, newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	tests := []struct {
		name   string
		token  string
		status string
		want   int
	}{
		{name: "missing token", want: fiber.StatusUnauthorized},
		{name: "different customer", token: orderToken(t, primitive.NewObjectID(), "customer"), want: fiber.StatusForbidden},
		{name: "processing", token: orderToken(t, ownerID, "customer"), status: models.OrderStatusProcessing, want: fiber.StatusBadRequest},
		{name: "shipped", token: orderToken(t, ownerID, "customer"), status: models.OrderStatusShipped, want: fiber.StatusBadRequest},
		{name: "delivered", token: orderToken(t, ownerID, "customer"), status: models.OrderStatusDelivered, want: fiber.StatusBadRequest},
		{name: "already cancelled", token: orderToken(t, ownerID, "customer"), status: models.OrderStatusCancelled, want: fiber.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			order := orders.orders[orderID]
			if test.status == "" {
				order.Status = models.OrderStatusPending
			} else {
				order.Status = test.status
			}
			orders.orders[orderID] = order

			req := httptest.NewRequest("PUT", "/api/orders/"+orderID.Hex()+"/cancel", nil)
			if test.token != "" {
				req.Header.Set(fiber.HeaderAuthorization, "Bearer "+test.token)
			}
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != test.want {
				t.Fatalf("status = %d, want %d", res.StatusCode, test.want)
			}
		})
	}
}

func TestOrderCancelRouteReturnsNotFound(t *testing.T) {
	userID := primitive.NewObjectID()
	service := services.NewOrderService(newRouteOrderRepository(), newRouteCartRepository(), &routeProductRepository{products: map[primitive.ObjectID]models.Product{}})
	app := orderTestApp(controllers.NewOrderController(service))

	req := httptest.NewRequest("PUT", "/api/orders/"+primitive.NewObjectID().Hex()+"/cancel", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+orderToken(t, userID, "customer"))
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusNotFound)
	}
}
