package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ProviderID  bson.ObjectID `bson:"provider_id" json:"provider_id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	Category    string        `bson:"category" json:"category"`
	City        string        `bson:"city" json:"city"`
	Price       int64         `bson:"price" json:"price"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}
