package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	OrderNew      OrderStatus = "new"
	OrderAccepted OrderStatus = "accepted"
	OrderDone     OrderStatus = "done"
	OrderCanceled OrderStatus = "canceled"
)

type OrderMessage struct {
	SenderID  bson.ObjectID `bson:"sender_id" json:"sender_id"`
	Text      string        `bson:"text" json:"text"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

type Order struct {
	ID         bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	ServiceID  bson.ObjectID  `bson:"service_id" json:"service_id"`
	ClientID   bson.ObjectID  `bson:"client_id" json:"client_id"`
	ProviderID bson.ObjectID  `bson:"provider_id" json:"provider_id"`
	Status     OrderStatus    `bson:"status" json:"status"`
	Details    string         `bson:"details" json:"details"`
	Messages   []OrderMessage `bson:"messages" json:"messages"`
	CreatedAt  time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `bson:"updated_at" json:"updated_at"`
}
