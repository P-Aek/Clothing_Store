package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

var (
	ErrInvalidCartInput       = errors.New("invalid cart input")
	ErrProductVariantNotFound = errors.New("product variant not found")
	ErrInsufficientStock      = errors.New("insufficient stock")
)

const MaxCartItemQuantity = 99

type CartService struct {
	carts    repositories.CartRepository
	products repositories.ProductRepository
	now      func() time.Time
}

func NewCartService(carts repositories.CartRepository, products repositories.ProductRepository) *CartService {
	return &CartService{carts: carts, products: products, now: time.Now}
}

func (s *CartService) Get(ctx context.Context, userID primitive.ObjectID) (models.Cart, error) {
	cart, err := s.carts.FindByUserID(ctx, userID)
	if errors.Is(err, repositories.ErrCartNotFound) {
		return models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
	}
	return cart, err
}

func (s *CartService) AddItem(ctx context.Context, userID, productID, variantID primitive.ObjectID, quantity int) (models.Cart, error) {
	if err := validateCartItemIDs(userID, productID, variantID); err != nil {
		return models.Cart{}, err
	}
	if quantity < 1 || quantity > MaxCartItemQuantity {
		return models.Cart{}, ErrInvalidCartInput
	}
	product, variant, err := s.findVariant(ctx, productID, variantID)
	if err != nil {
		return models.Cart{}, err
	}
	existing, err := s.Get(ctx, userID)
	if err != nil {
		return models.Cart{}, err
	}
	currentQuantity := cartItemQuantity(existing, product.ID, variant.ID)
	if currentQuantity+quantity > variant.Stock {
		return models.Cart{}, ErrInsufficientStock
	}
	if err := s.carts.UpsertItem(ctx, userID, models.CartItem{ProductID: product.ID, VariantID: variant.ID, Quantity: quantity}, s.now().UTC()); err != nil {
		return models.Cart{}, err
	}
	return s.Get(ctx, userID)
}

func (s *CartService) UpdateItem(ctx context.Context, userID, productID, variantID primitive.ObjectID, quantity int) (models.Cart, error) {
	if err := validateCartItemIDs(userID, productID, variantID); err != nil {
		return models.Cart{}, err
	}
	if quantity < 1 || quantity > MaxCartItemQuantity {
		return models.Cart{}, ErrInvalidCartInput
	}
	product, variant, err := s.findVariant(ctx, productID, variantID)
	if err != nil {
		return models.Cart{}, err
	}
	if quantity > variant.Stock {
		return models.Cart{}, ErrInsufficientStock
	}
	if err := s.carts.UpdateItemQuantity(ctx, userID, product.ID, variant.ID, quantity, s.now().UTC()); err != nil {
		if errors.Is(err, repositories.ErrCartNotFound) {
			return models.Cart{}, repositories.ErrCartItemNotFound
		}
		return models.Cart{}, err
	}
	return s.Get(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID, productID, variantID primitive.ObjectID) error {
	if err := validateCartItemIDs(userID, productID, variantID); err != nil {
		return err
	}
	err := s.carts.RemoveItem(ctx, userID, productID, variantID, s.now().UTC())
	if errors.Is(err, repositories.ErrCartNotFound) {
		return repositories.ErrCartItemNotFound
	}
	return err
}

func (s *CartService) findVariant(ctx context.Context, productID, variantID primitive.ObjectID) (models.Product, models.ProductVariant, error) {
	product, err := s.products.FindByID(ctx, productID)
	if errors.Is(err, repositories.ErrProductNotFound) {
		return models.Product{}, models.ProductVariant{}, repositories.ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, models.ProductVariant{}, err
	}
	for _, variant := range product.Variants {
		if variant.ID == variantID {
			return product, variant, nil
		}
	}
	return models.Product{}, models.ProductVariant{}, ErrProductVariantNotFound
}

func validateCartItemIDs(userID, productID, variantID primitive.ObjectID) error {
	if userID.IsZero() || productID.IsZero() || variantID.IsZero() {
		return ErrInvalidCartInput
	}
	return nil
}

func cartItemQuantity(cart models.Cart, productID, variantID primitive.ObjectID) int {
	for _, item := range cart.Items {
		if item.ProductID == productID && item.VariantID == variantID {
			return item.Quantity
		}
	}
	return 0
}
