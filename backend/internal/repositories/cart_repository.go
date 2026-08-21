package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"clothing-store-api/internal/models"
)

type CartRepository interface {
	FindByUserID(context.Context, primitive.ObjectID) (models.Cart, error)
	UpsertItem(context.Context, primitive.ObjectID, models.CartItem, time.Time) error
	UpdateItemQuantity(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, int, time.Time) error
	RemoveItem(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, time.Time) error
}

type MongoCartRepository struct {
	collection *mongo.Collection
}

func NewMongoCartRepository(database *mongo.Database) *MongoCartRepository {
	return &MongoCartRepository{collection: database.Collection("carts")}
}

func (r *MongoCartRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_cart_user"),
	})
	return err
}

func (r *MongoCartRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) (models.Cart, error) {
	var cart models.Cart
	err := r.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&cart)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Cart{}, ErrCartNotFound
	}
	if err != nil {
		return models.Cart{}, err
	}
	if cart.Items == nil {
		cart.Items = []models.CartItem{}
	}
	return cart, nil
}

func (r *MongoCartRepository) UpsertItem(ctx context.Context, userID primitive.ObjectID, item models.CartItem, updatedAt time.Time) error {
	filter := bson.M{"userId": userID, "items": bson.M{"$elemMatch": bson.M{"productId": item.ProductID, "variantId": item.VariantID}}}
	update := bson.M{"$inc": bson.M{"items.$.quantity": item.Quantity}, "$set": bson.M{"updatedAt": updatedAt}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount > 0 {
		return nil
	}

	filter = bson.M{"userId": userID}
	update = bson.M{
		"$set":         bson.M{"updatedAt": updatedAt},
		"$setOnInsert": bson.M{"userId": userID, "createdAt": updatedAt},
		"$push":        bson.M{"items": item},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *MongoCartRepository) UpdateItemQuantity(ctx context.Context, userID, productID, variantID primitive.ObjectID, quantity int, updatedAt time.Time) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"userId": userID,
		"items":  bson.M{"$elemMatch": bson.M{"productId": productID, "variantId": variantID}},
	}, bson.M{
		"$set": bson.M{"items.$.quantity": quantity, "updatedAt": updatedAt},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (r *MongoCartRepository) RemoveItem(ctx context.Context, userID, productID, variantID primitive.ObjectID, updatedAt time.Time) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"userId": userID,
		"items":  bson.M{"$elemMatch": bson.M{"productId": productID, "variantId": variantID}},
	}, bson.M{
		"$pull": bson.M{"items": bson.M{"productId": productID, "variantId": variantID}},
		"$set":  bson.M{"updatedAt": updatedAt},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrCartItemNotFound
	}
	return nil
}
