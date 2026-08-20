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

type testCategoryRepository struct {
	categories map[primitive.ObjectID]models.Category
}

func newTestCategoryRepository() *testCategoryRepository {
	return &testCategoryRepository{categories: map[primitive.ObjectID]models.Category{}}
}

func (r *testCategoryRepository) Create(_ context.Context, category models.Category) (models.Category, error) {
	r.categories[category.ID] = category
	return category, nil
}
func (r *testCategoryRepository) List(_ context.Context) ([]models.Category, error) {
	result := make([]models.Category, 0, len(r.categories))
	for _, category := range r.categories {
		if category.Active {
			result = append(result, category)
		}
	}
	return result, nil
}
func (r *testCategoryRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Category, error) {
	category, ok := r.categories[id]
	if !ok || !category.Active {
		return models.Category{}, repositories.ErrCategoryNotFound
	}
	return category, nil
}
func (r *testCategoryRepository) Update(_ context.Context, id primitive.ObjectID, category models.Category) (models.Category, error) {
	existing, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Category{}, err
	}
	existing.Name, existing.Slug = category.Name, category.Slug
	r.categories[id] = existing
	return existing, nil
}
func (r *testCategoryRepository) Delete(_ context.Context, id primitive.ObjectID, _ time.Time) error {
	category, err := r.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	category.Active = false
	r.categories[id] = category
	return nil
}

func categoryTestApp() *fiber.App {
	repo := newTestCategoryRepository()
	controller := controllers.NewCategoryController(services.NewCategoryService(repo))
	app := fiber.New()
	Register(app, nil, nil, controller, "test-secret")
	return app
}

func categoryToken(t *testing.T, role string) string {
	t.Helper()
	token, err := utils.GenerateJWT(primitive.NewObjectID().Hex(), role, "test-secret", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestCategoryRoutesProtectAdminMutations(t *testing.T) {
	app := categoryTestApp()

	for _, test := range []struct {
		name   string
		header string
		want   int
	}{
		{name: "missing token", want: fiber.StatusUnauthorized},
		{name: "customer token", header: "Bearer " + categoryToken(t, "customer"), want: fiber.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/categories/", strings.NewReader(`{"name":"Hats","slug":"hats"}`))
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

func TestCategoryRoutesAllowAdminCreate(t *testing.T) {
	app := categoryTestApp()
	req := httptest.NewRequest("POST", "/api/categories/", strings.NewReader(`{"name":"Hats","slug":"hats"}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+categoryToken(t, "admin"))
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d", res.StatusCode)
	}
}
