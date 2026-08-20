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

type memoryProductRepository struct {
	products map[primitive.ObjectID]models.Product
}

func newMemoryProductRepository() *memoryProductRepository {
	return &memoryProductRepository{products: map[primitive.ObjectID]models.Product{}}
}

func (r *memoryProductRepository) Create(_ context.Context, product models.Product) (models.Product, error) {
	r.products[product.ID] = product
	return product, nil
}
func (r *memoryProductRepository) List(_ context.Context, categoryID *primitive.ObjectID) ([]models.Product, error) {
	result := make([]models.Product, 0)
	for _, product := range r.products {
		if product.Active && (categoryID == nil || product.CategoryID == *categoryID) {
			result = append(result, product)
		}
	}
	return result, nil
}
func (r *memoryProductRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Product, error) {
	product, ok := r.products[id]
	if !ok || !product.Active {
		return models.Product{}, repositories.ErrProductNotFound
	}
	return product, nil
}
func (r *memoryProductRepository) Update(_ context.Context, id primitive.ObjectID, product models.Product) (models.Product, error) {
	existing, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Product{}, err
	}
	product.ID, product.CreatedAt, product.Active = id, existing.CreatedAt, true
	r.products[id] = product
	return product, nil
}
func (r *memoryProductRepository) Delete(_ context.Context, id primitive.ObjectID, updatedAt time.Time) error {
	product, err := r.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	product.Active, product.UpdatedAt = false, updatedAt
	r.products[id] = product
	return nil
}

type memoryProductCategoryRepository struct {
	categories map[primitive.ObjectID]models.Category
}

func newMemoryProductCategoryRepository() *memoryProductCategoryRepository {
	return &memoryProductCategoryRepository{categories: map[primitive.ObjectID]models.Category{}}
}
func (r *memoryProductCategoryRepository) Create(_ context.Context, category models.Category) (models.Category, error) {
	r.categories[category.ID] = category
	return category, nil
}
func (r *memoryProductCategoryRepository) List(_ context.Context) ([]models.Category, error) {
	return nil, nil
}
func (r *memoryProductCategoryRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Category, error) {
	category, ok := r.categories[id]
	if !ok || !category.Active {
		return models.Category{}, repositories.ErrCategoryNotFound
	}
	return category, nil
}
func (r *memoryProductCategoryRepository) Update(context.Context, primitive.ObjectID, models.Category) (models.Category, error) {
	return models.Category{}, repositories.ErrCategoryNotFound
}
func (r *memoryProductCategoryRepository) Delete(context.Context, primitive.ObjectID, time.Time) error {
	return repositories.ErrCategoryNotFound
}

func TestProductServiceCreateValidatesCategoryAndVariants(t *testing.T) {
	products := newMemoryProductRepository()
	categories := newMemoryProductCategoryRepository()
	category := models.Category{ID: primitive.NewObjectID(), Name: "Shirts", Slug: "shirts", Active: true}
	categories.categories[category.ID] = category
	service := NewProductService(products, categories)
	service.now = func() time.Time { return time.Unix(2_000_000_000, 0) }

	created, err := service.Create(context.Background(), ProductInput{
		CategoryID: category.ID, Name: "  Oxford Shirt ", Description: "Cotton shirt", Price: 49.99,
		Images: []string{" https://example.com/shirt.jpg "}, Variants: []ProductVariantInput{{Color: "Blue", Size: "M", Stock: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Oxford Shirt" || created.Images[0] != "https://example.com/shirt.jpg" || len(created.Variants) != 1 || created.Variants[0].ID.IsZero() {
		t.Fatalf("unexpected product: %+v", created)
	}
	if _, err := service.Create(context.Background(), ProductInput{CategoryID: category.ID, Name: "x", Price: 1, Variants: []ProductVariantInput{{Color: "Red", Size: "S", Stock: -1}}}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("validation error = %v", err)
	}
	if _, err := service.Create(context.Background(), ProductInput{CategoryID: primitive.NewObjectID(), Name: "Hat", Price: 10, Variants: []ProductVariantInput{{Color: "Black", Size: "M"}}}); !errors.Is(err, repositories.ErrCategoryNotFound) {
		t.Fatalf("category error = %v", err)
	}
}

func TestProductServiceListFilterAndDelete(t *testing.T) {
	products := newMemoryProductRepository()
	categories := newMemoryProductCategoryRepository()
	category := models.Category{ID: primitive.NewObjectID(), Name: "Shoes", Slug: "shoes", Active: true}
	categories.categories[category.ID] = category
	service := NewProductService(products, categories)
	product, err := service.Create(context.Background(), ProductInput{CategoryID: category.ID, Name: "Sneakers", Price: 80, Variants: []ProductVariantInput{{Color: "White", Size: "42"}}})
	if err != nil {
		t.Fatal(err)
	}
	filtered, err := service.List(context.Background(), &category.ID)
	if err != nil || len(filtered) != 1 || filtered[0].ID != product.ID {
		t.Fatalf("List() result = %+v, error = %v", filtered, err)
	}
	if err := service.Delete(context.Background(), product.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(context.Background(), product.ID); !errors.Is(err, repositories.ErrProductNotFound) {
		t.Fatalf("deleted product error = %v", err)
	}
}
