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

const productTestCategoryHex = "507f1f77bcf86cd799439011"

func productTestApp() *fiber.App {
	categories := &routeCategoryRepository{categories: map[primitive.ObjectID]models.Category{}}
	categoryID, _ := primitive.ObjectIDFromHex(productTestCategoryHex)
	category := models.Category{ID: categoryID, Name: "Shirts", Slug: "shirts", Active: true}
	categories.categories[category.ID] = category
	products := &routeProductRepository{products: map[primitive.ObjectID]models.Product{}}
	controller := controllers.NewProductController(services.NewProductService(products, categories))
	app := fiber.New()
	Register(app, nil, nil, nil, controller, nil, "test-secret")
	return app
}

func productRouteToken(t *testing.T, role string) string {
	t.Helper()
	token, err := utils.GenerateJWT(primitive.NewObjectID().Hex(), role, "test-secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestProductRoutesProtectAdminMutations(t *testing.T) {
	app := productTestApp()
	for _, test := range []struct {
		name   string
		header string
		want   int
	}{
		{name: "missing token", want: fiber.StatusUnauthorized},
		{name: "customer token", header: "Bearer " + productRouteToken(t, "customer"), want: fiber.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/products/", strings.NewReader(`{"name":"Oxford Shirt","categoryId":"507f1f77bcf86cd799439011","price":49.99,"variants":[{"color":"Blue","size":"M","stock":2}]}`))
			req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
			if test.header != "" {
				req.Header.Set(fiber.HeaderAuthorization, test.header)
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

func TestProductRoutesRejectInvalidCategoryFilter(t *testing.T) {
	app := productTestApp()
	res, err := app.Test(httptest.NewRequest("GET", "/api/products?categoryId=invalid", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestProductRoutesAllowAdminCreate(t *testing.T) {
	app := productTestApp()
	req := httptest.NewRequest("POST", "/api/products/", strings.NewReader(`{"name":"Oxford Shirt","categoryId":"507f1f77bcf86cd799439011","price":49.99,"variants":[{"color":"Blue","size":"M","stock":2}]}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+productRouteToken(t, "admin"))
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

type routeCategoryRepository struct {
	categories map[primitive.ObjectID]models.Category
}

func (r *routeCategoryRepository) Create(_ context.Context, category models.Category) (models.Category, error) {
	r.categories[category.ID] = category
	return category, nil
}
func (r *routeCategoryRepository) List(context.Context) ([]models.Category, error) { return nil, nil }
func (r *routeCategoryRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Category, error) {
	category, ok := r.categories[id]
	if !ok || !category.Active {
		return models.Category{}, repositories.ErrCategoryNotFound
	}
	return category, nil
}
func (r *routeCategoryRepository) Update(context.Context, primitive.ObjectID, models.Category) (models.Category, error) {
	return models.Category{}, repositories.ErrCategoryNotFound
}
func (r *routeCategoryRepository) Delete(context.Context, primitive.ObjectID, time.Time) error {
	return repositories.ErrCategoryNotFound
}

type routeProductRepository struct {
	products map[primitive.ObjectID]models.Product
}

func (r *routeProductRepository) Create(_ context.Context, product models.Product) (models.Product, error) {
	r.products[product.ID] = product
	return product, nil
}
func (r *routeProductRepository) List(context.Context, *primitive.ObjectID) ([]models.Product, error) {
	return []models.Product{}, nil
}
func (r *routeProductRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Product, error) {
	product, ok := r.products[id]
	if !ok || !product.Active {
		return models.Product{}, repositories.ErrProductNotFound
	}
	return product, nil
}
func (r *routeProductRepository) Update(context.Context, primitive.ObjectID, models.Product) (models.Product, error) {
	return models.Product{}, repositories.ErrProductNotFound
}
func (r *routeProductRepository) Delete(context.Context, primitive.ObjectID, time.Time) error {
	return repositories.ErrProductNotFound
}
