package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRole string

const (
	RoleClient   UserRole = "client"
	RoleProvider UserRole = "provider"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Role         UserRole           `bson:"role" json:"role"`
	Name         string             `bson:"name,omitempty" json:"name,omitempty"`
	Phone        string             `bson:"phone,omitempty" json:"phone,omitempty"`
	City         string             `bson:"city,omitempty" json:"city,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
