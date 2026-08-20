package repositories

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"clothing-store-api/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, newUser models.User) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{collection: database.Collection("users")}
}

func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_user_email"),
	})
	return err
}

func (r *MongoUserRepository) Create(ctx context.Context, newUser models.User) (models.User, error) {
	result, err := r.collection.InsertOne(ctx, newUser)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.User{}, ErrEmailAlreadyExists
		}
		return models.User{}, err
	}
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return models.User{}, errors.New("database returned an invalid user id")
	}
	newUser.ID = insertedID
	return newUser, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var found models.User
	err := r.collection.FindOne(ctx, bson.M{"email": strings.ToLower(strings.TrimSpace(email))}).Decode(&found)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return found, nil
}
