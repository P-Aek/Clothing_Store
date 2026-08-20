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

type ProductRepository interface {
	Create(context.Context, models.Product) (models.Product, error)
	List(context.Context, *primitive.ObjectID) ([]models.Product, error)
	FindByID(context.Context, primitive.ObjectID) (models.Product, error)
	Update(context.Context, primitive.ObjectID, models.Product) (models.Product, error)
	Delete(context.Context, primitive.ObjectID, time.Time) error
}

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(database *mongo.Database) *MongoProductRepository {
	return &MongoProductRepository{collection: database.Collection("products")}
}

func (r *MongoProductRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "categoryId", Value: 1}, {Key: "active", Value: 1}},
		Options: options.Index().SetName("products_by_category_active"),
	})
	return err
}

func (r *MongoProductRepository) Create(ctx context.Context, product models.Product) (models.Product, error) {
	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return models.Product{}, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return models.Product{}, errors.New("database returned an invalid product id")
	}
	product.ID = id
	return product, nil
}

func (r *MongoProductRepository) List(ctx context.Context, categoryID *primitive.ObjectID) ([]models.Product, error) {
	filter := bson.M{"active": true}
	if categoryID != nil {
		filter["categoryId"] = *categoryID
	}
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	if products == nil {
		products = []models.Product{}
	}
	return products, nil
}

func (r *MongoProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Product, error) {
	var product models.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "active": true}).Decode(&product)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (r *MongoProductRepository) Update(ctx context.Context, id primitive.ObjectID, product models.Product) (models.Product, error) {
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "active": true}, bson.M{"$set": bson.M{
		"categoryId": product.CategoryID, "name": product.Name, "description": product.Description,
		"price": product.Price, "images": product.Images, "variants": product.Variants, "updatedAt": product.UpdatedAt,
	}})
	if err != nil {
		return models.Product{}, err
	}
	if result.MatchedCount == 0 {
		return models.Product{}, ErrProductNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *MongoProductRepository) Delete(ctx context.Context, id primitive.ObjectID, updatedAt time.Time) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "active": true}, bson.M{"$set": bson.M{"active": false, "updatedAt": updatedAt}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrProductNotFound
	}
	return nil
}
