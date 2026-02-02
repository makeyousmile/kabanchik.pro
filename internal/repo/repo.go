package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"kabanchik.pro/internal/model"
)

type Store interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error

	CreateService(ctx context.Context, service *model.Service) error
	ListServices(ctx context.Context, filter ServiceFilter) ([]model.Service, error)
	GetServiceByID(ctx context.Context, id bson.ObjectID) (*model.Service, error)
	UpdateService(ctx context.Context, service *model.Service) error
	DeleteService(ctx context.Context, id bson.ObjectID, providerID bson.ObjectID) error

	CreateOrder(ctx context.Context, order *model.Order) error
	ListOrders(ctx context.Context, filter OrderFilter) ([]model.Order, error)
	GetOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, actorID bson.ObjectID) error
	AddOrderMessage(ctx context.Context, id bson.ObjectID, msg model.OrderMessage, actorID bson.ObjectID) error
}

type ServiceFilter struct {
	Category string
	City     string
	MinPrice int64
	MaxPrice int64
	Query    string
}

type OrderFilter struct {
	Status     model.OrderStatus
	ClientID   *bson.ObjectID
	ProviderID *bson.ObjectID
}
