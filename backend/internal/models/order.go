package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

type OrderItem struct {
	ProductID   primitive.ObjectID `bson:"productId" json:"productId"`
	VariantID   primitive.ObjectID `bson:"variantId" json:"variantId"`
	ProductName string             `bson:"productName" json:"productName"`
	Color       string             `bson:"color" json:"color"`
	Size        string             `bson:"size" json:"size"`
	Price       float64            `bson:"price" json:"price"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Subtotal    float64            `bson:"subtotal" json:"subtotal"`
}

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Items      []OrderItem        `bson:"items" json:"items"`
	TotalPrice float64            `bson:"totalPrice" json:"totalPrice"`
	Status     string             `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}
