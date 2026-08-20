package services

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

var ErrInvalidProductInput = errors.New("invalid product input")

type ProductService struct {
	products   repositories.ProductRepository
	categories repositories.CategoryRepository
	now        func() time.Time
}

func NewProductService(products repositories.ProductRepository, categories repositories.CategoryRepository) *ProductService {
	return &ProductService{products: products, categories: categories, now: time.Now}
}

type ProductVariantInput struct {
	ID    primitive.ObjectID
	Color string
	Size  string
	Stock int
}

type ProductInput struct {
	CategoryID  primitive.ObjectID
	Name        string
	Description string
	Price       float64
	Images      []string
	Variants    []ProductVariantInput
}

func (s *ProductService) Create(ctx context.Context, input ProductInput) (models.Product, error) {
	product, err := s.buildProduct(ctx, input, primitive.NilObjectID, time.Time{})
	if err != nil {
		return models.Product{}, err
	}
	return s.products.Create(ctx, product)
}

func (s *ProductService) List(ctx context.Context, categoryID *primitive.ObjectID) ([]models.Product, error) {
	if categoryID != nil {
		if _, err := s.categories.FindByID(ctx, *categoryID); err != nil {
			return nil, err
		}
	}
	return s.products.List(ctx, categoryID)
}

func (s *ProductService) Get(ctx context.Context, id primitive.ObjectID) (models.Product, error) {
	return s.products.FindByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id primitive.ObjectID, input ProductInput) (models.Product, error) {
	product, err := s.buildProduct(ctx, input, id, time.Time{})
	if err != nil {
		return models.Product{}, err
	}
	return s.products.Update(ctx, id, product)
}

func (s *ProductService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.products.Delete(ctx, id, s.now().UTC())
}

func (s *ProductService) buildProduct(ctx context.Context, input ProductInput, id primitive.ObjectID, createdAt time.Time) (models.Product, error) {
	if input.CategoryID.IsZero() {
		return models.Product{}, ErrInvalidProductInput
	}
	if _, err := s.categories.FindByID(ctx, input.CategoryID); err != nil {
		if errors.Is(err, repositories.ErrCategoryNotFound) {
			return models.Product{}, err
		}
		return models.Product{}, err
	}
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if len(name) < 2 || len(name) > 200 || len(description) > 5000 || input.Price <= 0 || math.IsNaN(input.Price) || math.IsInf(input.Price, 0) {
		return models.Product{}, ErrInvalidProductInput
	}
	images, err := validateImages(input.Images)
	if err != nil {
		return models.Product{}, err
	}
	variants, err := validateVariants(input.Variants)
	if err != nil {
		return models.Product{}, err
	}
	now := s.now().UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	return models.Product{
		ID: id, CategoryID: input.CategoryID, Name: name, Description: description,
		Price: input.Price, Images: images, Variants: variants, Active: true,
		CreatedAt: createdAt, UpdatedAt: now,
	}, nil
}

func validateImages(images []string) ([]string, error) {
	if len(images) > 10 {
		return nil, ErrInvalidProductInput
	}
	result := make([]string, 0, len(images))
	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" || len(image) > 2048 {
			return nil, ErrInvalidProductInput
		}
		result = append(result, image)
	}
	return result, nil
}

func validateVariants(variants []ProductVariantInput) ([]models.ProductVariant, error) {
	if len(variants) == 0 || len(variants) > 100 {
		return nil, ErrInvalidProductInput
	}
	result := make([]models.ProductVariant, 0, len(variants))
	seen := make(map[string]struct{}, len(variants))
	for _, variant := range variants {
		color := strings.TrimSpace(variant.Color)
		size := strings.TrimSpace(variant.Size)
		key := strings.ToLower(color) + "\x00" + strings.ToLower(size)
		if color == "" || len(color) > 50 || size == "" || len(size) > 30 || variant.Stock < 0 {
			return nil, ErrInvalidProductInput
		}
		if _, exists := seen[key]; exists {
			return nil, ErrInvalidProductInput
		}
		seen[key] = struct{}{}
		variantID := variant.ID
		if variantID.IsZero() {
			variantID = primitive.NewObjectID()
		}
		result = append(result, models.ProductVariant{ID: variantID, Color: color, Size: size, Stock: variant.Stock})
	}
	return result, nil
}
