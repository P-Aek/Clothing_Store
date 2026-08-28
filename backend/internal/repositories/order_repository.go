package repositories

import (
	"context"
	"errors"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"clothing-store-api/internal/models"
)

type OrderRepository interface {
	CreateFromCart(context.Context, primitive.ObjectID, []models.OrderItem, float64) (models.Order, error)
	FindByID(context.Context, primitive.ObjectID) (models.Order, error)
	ListByUserID(context.Context, primitive.ObjectID, int, int) (models.OrderListResponse, error)
	ListAll(context.Context, int, int) (models.OrderListResponse, error)
	UpdateStatus(context.Context, primitive.ObjectID, string, time.Time) (models.Order, error)
	Cancel(context.Context, primitive.ObjectID, *primitive.ObjectID, time.Time) (models.Order, error)
}

type MongoOrderRepository struct {
	database *mongo.Database
	orders   *mongo.Collection
	products *mongo.Collection
	carts    *mongo.Collection
}

func NewMongoOrderRepository(database *mongo.Database) *MongoOrderRepository {
	return &MongoOrderRepository{
		database: database,
		orders:   database.Collection("orders"),
		products: database.Collection("products"),
		carts:    database.Collection("carts"),
	}
}

func (r *MongoOrderRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.orders.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_by_user_created")},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}, Options: options.Index().SetName("orders_by_status_created")},
	})
	return err
}

func (r *MongoOrderRepository) CreateFromCart(ctx context.Context, userID primitive.ObjectID, items []models.OrderItem, total float64) (models.Order, error) {
	now := time.Now().UTC()
	order := models.Order{ID: primitive.NewObjectID(), UserID: userID, Items: items, TotalPrice: total, Status: models.OrderStatusPending, CreatedAt: now, UpdatedAt: now}

	session, err := r.database.Client().StartSession()
	if err != nil {
		return models.Order{}, err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		var currentCart models.Cart
		if findErr := r.carts.FindOne(sc, bson.M{"userId": userID}).Decode(&currentCart); findErr != nil {
			return nil, ErrCartChanged
		}
		if !sameCartItems(currentCart.Items, items) {
			return nil, ErrCartChanged
		}
		for _, item := range items {
			result, updateErr := r.products.UpdateOne(sc,
				bson.M{"_id": item.ProductID, "active": true, "variants": bson.M{"$elemMatch": bson.M{"_id": item.VariantID, "stock": bson.M{"$gte": item.Quantity}}}},
				bson.M{"$inc": bson.M{"variants.$[variant].stock": -item.Quantity}},
				options.Update().SetArrayFilters(options.ArrayFilters{Filters: []interface{}{bson.M{"variant._id": item.VariantID}}}),
			)
			if updateErr != nil {
				return nil, updateErr
			}
			if result.MatchedCount == 0 {
				return nil, ErrOrderStockUnavailable
			}
		}
		if _, insertErr := r.orders.InsertOne(sc, order); insertErr != nil {
			return nil, insertErr
		}
		if _, deleteErr := r.carts.DeleteOne(sc, bson.M{"userId": userID}); deleteErr != nil {
			return nil, deleteErr
		}
		return nil, nil
	})
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func sameCartItems(cartItems []models.CartItem, orderItems []models.OrderItem) bool {
	if len(cartItems) != len(orderItems) {
		return false
	}
	for index, cartItem := range cartItems {
		orderItem := orderItems[index]
		if cartItem.ProductID != orderItem.ProductID || cartItem.VariantID != orderItem.VariantID || cartItem.Quantity != orderItem.Quantity {
			return false
		}
	}
	return true
}

func (r *MongoOrderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (models.Order, error) {
	var order models.Order
	err := r.orders.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (r *MongoOrderRepository) ListByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) (models.OrderListResponse, error) {
	return r.list(ctx, bson.M{"userId": userID}, page, limit)
}

func (r *MongoOrderRepository) ListAll(ctx context.Context, page, limit int) (models.OrderListResponse, error) {
	return r.list(ctx, bson.M{}, page, limit)
}

func (r *MongoOrderRepository) list(ctx context.Context, filter bson.M, page, limit int) (models.OrderListResponse, error) {
	totalItems, err := r.orders.CountDocuments(ctx, filter)
	if err != nil {
		return models.OrderListResponse{}, err
	}
	skip := int64(page-1) * int64(limit)
	cursor, err := r.orders.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit)))
	if err != nil {
		return models.OrderListResponse{}, err
	}
	defer cursor.Close(ctx)
	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return models.OrderListResponse{}, err
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return models.OrderListResponse{Orders: orders, Pagination: models.Pagination{
		Page: page, Limit: limit, TotalItems: totalItems,
		TotalPages: int(math.Ceil(float64(totalItems) / float64(limit))),
	}}, nil
}

func (r *MongoOrderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, updatedAt time.Time) (models.Order, error) {
	result, err := r.orders.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status, "updatedAt": updatedAt}})
	if err != nil {
		return models.Order{}, err
	}
	if result.MatchedCount == 0 {
		return models.Order{}, ErrOrderNotFound
	}
	return r.FindByID(ctx, id)
}

// Cancel atomically verifies the order, restores every ordered variant's stock,
// and marks the order cancelled. ownerID is nil for an authorized admin request.
func (r *MongoOrderRepository) Cancel(ctx context.Context, id primitive.ObjectID, ownerID *primitive.ObjectID, updatedAt time.Time) (models.Order, error) {
	session, err := r.database.Client().StartSession()
	if err != nil {
		return models.Order{}, err
	}
	defer session.EndSession(ctx)

	var cancelled models.Order
	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		var order models.Order
		if findErr := r.orders.FindOne(sc, bson.M{"_id": id}).Decode(&order); findErr != nil {
			if errors.Is(findErr, mongo.ErrNoDocuments) {
				return nil, ErrOrderNotFound
			}
			return nil, findErr
		}
		if ownerID != nil && order.UserID != *ownerID {
			return nil, ErrOrderNotOwned
		}
		switch order.Status {
		case models.OrderStatusCancelled:
			return nil, ErrOrderAlreadyCancelled
		case models.OrderStatusPending:
			// Continue with cancellation.
		default:
			return nil, ErrOrderCannotBeCancelled
		}

		for _, item := range order.Items {
			if item.ProductID.IsZero() || item.VariantID.IsZero() || item.Quantity < 1 {
				return nil, ErrOrderStockRestoreFailed
			}
			result, restoreErr := r.products.UpdateOne(sc,
				bson.M{"_id": item.ProductID, "variants": bson.M{"$elemMatch": bson.M{"_id": item.VariantID}}},
				bson.M{"$inc": bson.M{"variants.$[variant].stock": item.Quantity}},
				options.Update().SetArrayFilters(options.ArrayFilters{Filters: []interface{}{bson.M{"variant._id": item.VariantID}}}),
			)
			if restoreErr != nil {
				return nil, restoreErr
			}
			if result.MatchedCount == 0 {
				return nil, ErrOrderStockRestoreFailed
			}
		}

		result, updateErr := r.orders.UpdateOne(sc,
			bson.M{"_id": id, "status": models.OrderStatusPending},
			bson.M{"$set": bson.M{"status": models.OrderStatusCancelled, "updatedAt": updatedAt}},
		)
		if updateErr != nil {
			return nil, updateErr
		}
		if result.MatchedCount == 0 {
			return nil, ErrOrderCannotBeCancelled
		}

		order.Status = models.OrderStatusCancelled
		order.UpdatedAt = updatedAt
		cancelled = order
		return nil, nil
	})
	if err != nil {
		return models.Order{}, err
	}
	return cancelled, nil
}
