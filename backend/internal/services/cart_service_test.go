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

type memoryCartRepository struct {
	carts map[primitive.ObjectID]models.Cart
}

func newMemoryCartRepository() *memoryCartRepository {
	return &memoryCartRepository{carts: map[primitive.ObjectID]models.Cart{}}
}

func (r *memoryCartRepository) FindByUserID(_ context.Context, userID primitive.ObjectID) (models.Cart, error) {
	cart, ok := r.carts[userID]
	if !ok {
		return models.Cart{}, repositories.ErrCartNotFound
	}
	return cart, nil
}

func (r *memoryCartRepository) UpsertItem(_ context.Context, userID primitive.ObjectID, item models.CartItem, updatedAt time.Time) error {
	cart := r.carts[userID]
	if cart.UserID.IsZero() {
		cart = models.Cart{ID: primitive.NewObjectID(), UserID: userID, CreatedAt: updatedAt}
	}
	found := false
	for index := range cart.Items {
		if cart.Items[index].ProductID == item.ProductID && cart.Items[index].VariantID == item.VariantID {
			cart.Items[index].Quantity += item.Quantity
			found = true
		}
	}
	if !found {
		cart.Items = append(cart.Items, item)
	}
	cart.UpdatedAt = updatedAt
	r.carts[userID] = cart
	return nil
}

func (r *memoryCartRepository) UpdateItemQuantity(_ context.Context, userID, productID, variantID primitive.ObjectID, quantity int, updatedAt time.Time) error {
	cart, err := r.FindByUserID(context.Background(), userID)
	if err != nil {
		return err
	}
	for index := range cart.Items {
		if cart.Items[index].ProductID == productID && cart.Items[index].VariantID == variantID {
			cart.Items[index].Quantity = quantity
			cart.UpdatedAt = updatedAt
			r.carts[userID] = cart
			return nil
		}
	}
	return repositories.ErrCartItemNotFound
}

func (r *memoryCartRepository) RemoveItem(_ context.Context, userID, productID, variantID primitive.ObjectID, updatedAt time.Time) error {
	cart, err := r.FindByUserID(context.Background(), userID)
	if err != nil {
		return err
	}
	for index := range cart.Items {
		if cart.Items[index].ProductID == productID && cart.Items[index].VariantID == variantID {
			cart.Items = append(cart.Items[:index], cart.Items[index+1:]...)
			cart.UpdatedAt = updatedAt
			r.carts[userID] = cart
			return nil
		}
	}
	return repositories.ErrCartItemNotFound
}

func TestCartServiceAddUpdateRemoveAndEmptyCart(t *testing.T) {
	userID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	products := newMemoryProductRepository()
	products.products[productID] = models.Product{
		ID: productID, Active: true,
		Variants: []models.ProductVariant{{ID: variantID, Color: "Blue", Size: "M", Stock: 5}},
	}
	carts := newMemoryCartRepository()
	service := NewCartService(carts, products)
	service.now = func() time.Time { return time.Unix(2_000_000_000, 0) }

	empty, err := service.Get(context.Background(), userID)
	if err != nil || empty.UserID != userID || len(empty.Items) != 0 {
		t.Fatalf("Get() empty result = %+v, error = %v", empty, err)
	}
	cart, err := service.AddItem(context.Background(), userID, productID, variantID, 2)
	if err != nil || len(cart.Items) != 1 || cart.Items[0].Quantity != 2 {
		t.Fatalf("AddItem() result = %+v, error = %v", cart, err)
	}
	cart, err = service.UpdateItem(context.Background(), userID, productID, variantID, 4)
	if err != nil || cart.Items[0].Quantity != 4 {
		t.Fatalf("UpdateItem() result = %+v, error = %v", cart, err)
	}
	if err := service.RemoveItem(context.Background(), userID, productID, variantID); err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	cart, err = service.Get(context.Background(), userID)
	if err != nil || len(cart.Items) != 0 {
		t.Fatalf("cart after removal = %+v, error = %v", cart, err)
	}
}

func TestCartServiceRejectsInvalidVariantQuantityStockAndMissingItem(t *testing.T) {
	userID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	variantID := primitive.NewObjectID()
	products := newMemoryProductRepository()
	products.products[productID] = models.Product{ID: productID, Active: true, Variants: []models.ProductVariant{{ID: variantID, Stock: 2}}}
	service := NewCartService(newMemoryCartRepository(), products)

	checks := []struct {
		name string
		call func() error
		want error
	}{
		{name: "invalid quantity", call: func() error {
			_, err := service.AddItem(context.Background(), userID, productID, variantID, 0)
			return err
		}, want: ErrInvalidCartInput},
		{name: "missing variant", call: func() error {
			_, err := service.AddItem(context.Background(), userID, productID, primitive.NewObjectID(), 1)
			return err
		}, want: ErrProductVariantNotFound},
		{name: "insufficient stock", call: func() error {
			_, err := service.AddItem(context.Background(), userID, productID, variantID, 3)
			return err
		}, want: ErrInsufficientStock},
		{name: "missing cart item", call: func() error { return service.RemoveItem(context.Background(), userID, productID, variantID) }, want: repositories.ErrCartItemNotFound},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.call(); !errors.Is(err, check.want) {
				t.Fatalf("error = %v, want %v", err, check.want)
			}
		})
	}
}
