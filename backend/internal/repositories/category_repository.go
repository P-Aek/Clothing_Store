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

type CategoryRepository interface {
	Create(context.Context, models.Category) (models.Category, error)
	List(context.Context) ([]models.Category, error)
	FindByID(context.Context, primitive.ObjectID) (models.Category, error)
	Update(context.Context, primitive.ObjectID, models.Category) (models.Category, error)
	Delete(context.Context, primitive.ObjectID, time.Time) error
}

type MongoCategoryRepository struct {
	collection *mongo.Collection
}

func NewMongoCategoryRepository(database *mongo.Database) *MongoCategoryRepository {
	return &MongoCategoryRepository{collection: database.Collection("categories")}
}

func (r *MongoCategoryRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_category_slug"),
	})
	return err
}

func (r *MongoCategoryRepository) Create(ctx context.Context, category models.Category) (models.Category, error) {
	result, err := r.collection.InsertOne(ctx, category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.Category{}, ErrCategorySlugAlreadyExists
		}
		return models.Category{}, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return models.Category{}, errors.New("database returned an invalid category id")
	}
	category.ID = id
	return category, nil
}

func (r *MongoCategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"active": true}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	if categories == nil {
		categories = []models.Category{}
	}
	return categories, nil
}

func (r *MongoCategoryRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Category, error) {
	var category models.Category
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "active": true}).Decode(&category)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Category{}, ErrCategoryNotFound
	}
	if err != nil {
		return models.Category{}, err
	}
	return category, nil
}

func (r *MongoCategoryRepository) Update(ctx context.Context, id primitive.ObjectID, category models.Category) (models.Category, error) {
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "active": true}, bson.M{"$set": bson.M{
		"name": category.Name, "slug": category.Slug, "updatedAt": category.UpdatedAt,
	}})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.Category{}, ErrCategorySlugAlreadyExists
		}
		return models.Category{}, err
	}
	if result.MatchedCount == 0 {
		return models.Category{}, ErrCategoryNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *MongoCategoryRepository) Delete(ctx context.Context, id primitive.ObjectID, updatedAt time.Time) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "active": true}, bson.M{"$set": bson.M{"active": false, "updatedAt": updatedAt}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
