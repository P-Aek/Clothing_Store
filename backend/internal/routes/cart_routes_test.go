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

func TestCartRoutesRequireAuthentication(t *testing.T) {
	carts := newRouteCartRepository()
	products := &routeProductRepository{products: map[primitive.ObjectID]models.Product{}}
	cartController := controllers.NewCartController(services.NewCartService(carts, products))
	app := fiber.New()
	Register(app, nil, nil, nil, nil, cartController, nil, "test-secret")

	res, err := app.Test(httptest.NewRequest("GET", "/api/cart/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestCartRoutesAddItemForAuthenticatedUser(t *testing.T) {
	userID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	carts := newRouteCartRepository()
	products := &routeProductRepository{products: map[primitive.ObjectID]models.Product{}}
	products.products[productID] = models.Product{ID: productID, Active: true, Variants: []models.ProductVariant{{ID: variantID, Stock: 3}}}
	cartController := controllers.NewCartController(services.NewCartService(carts, products))
	app := fiber.New()
	Register(app, nil, nil, nil, nil, cartController, nil, "test-secret")
	token, err := utils.GenerateJWT(userID.Hex(), "customer", "test-secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(`{"productId":"`+productID.Hex()+`","variantId":"`+variantID.Hex()+`","quantity":1}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
}

type routeCartRepository struct {
	carts map[primitive.ObjectID]models.Cart
}

func newRouteCartRepository() *routeCartRepository {
	return &routeCartRepository{carts: map[primitive.ObjectID]models.Cart{}}
}
func (r *routeCartRepository) FindByUserID(_ context.Context, userID primitive.ObjectID) (models.Cart, error) {
	cart, ok := r.carts[userID]
	if !ok {
		return models.Cart{}, repositories.ErrCartNotFound
	}
	return cart, nil
}
func (r *routeCartRepository) UpsertItem(_ context.Context, userID primitive.ObjectID, item models.CartItem, now time.Time) error {
	cart := r.carts[userID]
	if cart.UserID.IsZero() {
		cart = models.Cart{ID: primitive.NewObjectID(), UserID: userID, CreatedAt: now}
	}
	cart.Items = append(cart.Items, item)
	cart.UpdatedAt = now
	r.carts[userID] = cart
	return nil
}
func (r *routeCartRepository) UpdateItemQuantity(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, int, time.Time) error {
	return repositories.ErrCartItemNotFound
}
func (r *routeCartRepository) RemoveItem(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, time.Time) error {
	return repositories.ErrCartItemNotFound
}
