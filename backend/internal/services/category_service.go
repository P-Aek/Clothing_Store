package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

var ErrInvalidCategoryInput = errors.New("invalid category input")

type CategoryService struct {
	repository repositories.CategoryRepository
	now        func() time.Time
}

func NewCategoryService(repository repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repository: repository, now: time.Now}
}

type CategoryInput struct {
	Name string
	Slug string
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *CategoryService) Create(ctx context.Context, input CategoryInput) (models.Category, error) {
	name, slug, err := validateCategoryInput(input)
	if err != nil {
		return models.Category{}, err
	}
	now := s.now().UTC()
	return s.repository.Create(ctx, models.Category{
		ID: primitive.NewObjectID(), Name: name, Slug: slug, Active: true,
		CreatedAt: now, UpdatedAt: now,
	})
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	return s.repository.List(ctx)
}

func (s *CategoryService) Get(ctx context.Context, id primitive.ObjectID) (models.Category, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *CategoryService) Update(ctx context.Context, id primitive.ObjectID, input CategoryInput) (models.Category, error) {
	name, slug, err := validateCategoryInput(input)
	if err != nil {
		return models.Category{}, err
	}
	return s.repository.Update(ctx, id, models.Category{Name: name, Slug: slug, UpdatedAt: s.now().UTC()})
}

func (s *CategoryService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repository.Delete(ctx, id, s.now().UTC())
}

func validateCategoryInput(input CategoryInput) (string, string, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if len(name) < 2 || len(name) > 100 || len(slug) < 2 || len(slug) > 100 || !slugPattern.MatchString(slug) {
		return "", "", fmt.Errorf("%w: name or slug is invalid", ErrInvalidCategoryInput)
	}
	return name, slug, nil
}
