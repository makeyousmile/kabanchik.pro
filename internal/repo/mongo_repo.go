package repo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"kabanchik.pro/internal/model"
)

const opTimeout = 5 * time.Second

var ErrNotFound = errors.New("not found")

// MongoStore implements Store using MongoDB.
type MongoStore struct {
	db *mongo.Database
}

func NewMongoStore(db *mongo.Database) *MongoStore {
	return &MongoStore{db: db}
}

func (s *MongoStore) users() *mongo.Collection    { return s.db.Collection("users") }
func (s *MongoStore) services() *mongo.Collection { return s.db.Collection("services") }
func (s *MongoStore) orders() *mongo.Collection   { return s.db.Collection("orders") }

func (s *MongoStore) CreateUser(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt

	_, err := s.users().InsertOne(ctx, user)
	return err
}

func (s *MongoStore) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var user model.User
	err := s.users().FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (s *MongoStore) FindUserByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var user model.User
	err := s.users().FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (s *MongoStore) UpdateUser(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	user.UpdatedAt = time.Now().UTC()
	update := bson.M{"$set": bson.M{
		"name":       user.Name,
		"phone":      user.Phone,
		"city":       user.City,
		"updated_at": user.UpdatedAt,
	}}
	res, err := s.users().UpdateByID(ctx, user.ID, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) CreateService(ctx context.Context, service *model.Service) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	service.CreatedAt = time.Now().UTC()
	service.UpdatedAt = service.CreatedAt

	_, err := s.services().InsertOne(ctx, service)
	return err
}

func (s *MongoStore) ListServices(ctx context.Context, filter ServiceFilter) ([]model.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	query := bson.M{}
	if filter.Category != "" {
		query["category"] = filter.Category
	}
	if filter.City != "" {
		query["city"] = filter.City
	}
	if filter.Query != "" {
		query["title"] = bson.M{"$regex": filter.Query, "$options": "i"}
	}
	if filter.MinPrice > 0 || filter.MaxPrice > 0 {
		price := bson.M{}
		if filter.MinPrice > 0 {
			price["$gte"] = filter.MinPrice
		}
		if filter.MaxPrice > 0 {
			price["$lte"] = filter.MaxPrice
		}
		query["price"] = price
	}

	cursor, err := s.services().Find(ctx, query, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []model.Service
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (s *MongoStore) GetServiceByID(ctx context.Context, id bson.ObjectID) (*model.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var service model.Service
	err := s.services().FindOne(ctx, bson.M{"_id": id}).Decode(&service)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &service, err
}

func (s *MongoStore) UpdateService(ctx context.Context, service *model.Service) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	service.UpdatedAt = time.Now().UTC()
	update := bson.M{"$set": bson.M{
		"title":       service.Title,
		"description": service.Description,
		"category":    service.Category,
		"city":        service.City,
		"price":       service.Price,
		"updated_at":  service.UpdatedAt,
	}}
	res, err := s.services().UpdateOne(ctx, bson.M{"_id": service.ID, "provider_id": service.ProviderID}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) DeleteService(ctx context.Context, id bson.ObjectID, providerID bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	res, err := s.services().DeleteOne(ctx, bson.M{"_id": id, "provider_id": providerID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) CreateOrder(ctx context.Context, order *model.Order) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = order.CreatedAt
	order.Status = model.OrderNew

	_, err := s.orders().InsertOne(ctx, order)
	return err
}

func (s *MongoStore) ListOrders(ctx context.Context, filter OrderFilter) ([]model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	query := bson.M{}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.ClientID != nil {
		query["client_id"] = *filter.ClientID
	}
	if filter.ProviderID != nil {
		query["provider_id"] = *filter.ProviderID
	}

	cursor, err := s.orders().Find(ctx, query, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *MongoStore) GetOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var order model.Order
	err := s.orders().FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &order, err
}

func (s *MongoStore) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, actorID bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now().UTC()}}
	res, err := s.orders().UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MongoStore) AddOrderMessage(ctx context.Context, id bson.ObjectID, msg model.OrderMessage, actorID bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	msg.CreatedAt = time.Now().UTC()
	update := bson.M{
		"$push": bson.M{"messages": msg},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}
	res, err := s.orders().UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
