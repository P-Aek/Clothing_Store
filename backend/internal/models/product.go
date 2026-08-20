package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductVariant struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Color string             `bson:"color" json:"color"`
	Size  string             `bson:"size" json:"size"`
	Stock int                `bson:"stock" json:"stock"`
}

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID  primitive.ObjectID `bson:"categoryId" json:"categoryId"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Images      []string           `bson:"images" json:"images"`
	Variants    []ProductVariant   `bson:"variants" json:"variants"`
	Active      bool               `bson:"active" json:"active"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
