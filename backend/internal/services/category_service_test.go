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

type memoryCategoryRepository struct {
	categories map[primitive.ObjectID]models.Category
}

func newMemoryCategoryRepository() *memoryCategoryRepository {
	return &memoryCategoryRepository{categories: map[primitive.ObjectID]models.Category{}}
}

func (r *memoryCategoryRepository) Create(_ context.Context, category models.Category) (models.Category, error) {
	for _, existing := range r.categories {
		if existing.Slug == category.Slug && existing.Active {
			return models.Category{}, repositories.ErrCategorySlugAlreadyExists
		}
	}
	r.categories[category.ID] = category
	return category, nil
}

func (r *memoryCategoryRepository) List(_ context.Context) ([]models.Category, error) {
	result := make([]models.Category, 0)
	for _, category := range r.categories {
		if category.Active {
			result = append(result, category)
		}
	}
	return result, nil
}

func (r *memoryCategoryRepository) FindByID(_ context.Context, id primitive.ObjectID) (models.Category, error) {
	category, ok := r.categories[id]
	if !ok || !category.Active {
		return models.Category{}, repositories.ErrCategoryNotFound
	}
	return category, nil
}

func (r *memoryCategoryRepository) Update(_ context.Context, id primitive.ObjectID, category models.Category) (models.Category, error) {
	existing, err := r.FindByID(context.Background(), id)
	if err != nil {
		return models.Category{}, err
	}
	for otherID, other := range r.categories {
		if otherID != id && other.Active && other.Slug == category.Slug {
			return models.Category{}, repositories.ErrCategorySlugAlreadyExists
		}
	}
	existing.Name, existing.Slug, existing.UpdatedAt = category.Name, category.Slug, category.UpdatedAt
	r.categories[id] = existing
	return existing, nil
}

func (r *memoryCategoryRepository) Delete(_ context.Context, id primitive.ObjectID, updatedAt time.Time) error {
	category, err := r.FindByID(context.Background(), id)
	if err != nil {
		return err
	}
	category.Active, category.UpdatedAt = false, updatedAt
	r.categories[id] = category
	return nil
}

func TestCategoryServiceCreateNormalizesAndValidatesInput(t *testing.T) {
	repo := newMemoryCategoryRepository()
	service := NewCategoryService(repo)
	service.now = func() time.Time { return time.Unix(2_000_000_000, 0) }

	category, err := service.Create(context.Background(), CategoryInput{Name: "  T-Shirts ", Slug: " T-SHIRTS "})
	if err != nil {
		t.Fatal(err)
	}
	if category.Name != "T-Shirts" || category.Slug != "t-shirts" || !category.Active {
		t.Fatalf("unexpected category: %+v", category)
	}
	if _, err := service.Create(context.Background(), CategoryInput{Name: "x", Slug: "bad slug"}); !errors.Is(err, ErrInvalidCategoryInput) {
		t.Fatalf("validation error = %v", err)
	}
}

func TestCategoryServiceDuplicateAndNotFound(t *testing.T) {
	repo := newMemoryCategoryRepository()
	service := NewCategoryService(repo)
	input := CategoryInput{Name: "Shoes", Slug: "shoes"}
	created, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), input); !errors.Is(err, repositories.ErrCategorySlugAlreadyExists) {
		t.Fatalf("duplicate error = %v", err)
	}
	if _, err := service.Get(context.Background(), primitive.NewObjectID()); !errors.Is(err, repositories.ErrCategoryNotFound) {
		t.Fatalf("not found error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(context.Background(), created.ID); !errors.Is(err, repositories.ErrCategoryNotFound) {
		t.Fatalf("deleted category error = %v", err)
	}
}
