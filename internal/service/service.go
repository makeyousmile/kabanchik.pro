package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"kabanchik.pro/internal/model"
	"kabanchik.pro/internal/repo"
)

var ErrUnauthorized = errors.New("unauthorized")

// AppService contains business logic.
type AppService struct {
	store repo.Store
}

func New(store repo.Store) *AppService {
	return &AppService{store: store}
}

func (s *AppService) Register(ctx context.Context, email, password string, role model.UserRole) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, errors.New("email and password required")
	}
	if role != model.RoleClient && role != model.RoleProvider {
		return nil, errors.New("invalid role")
	}

	_, err := s.store.FindUserByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AppService) Login(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.store.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrUnauthorized
	}
	return user, nil
}

func (s *AppService) GetUser(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	return s.store.FindUserByID(ctx, id)
}

func (s *AppService) UpdateUser(ctx context.Context, user *model.User) error {
	return s.store.UpdateUser(ctx, user)
}

func (s *AppService) CreateService(ctx context.Context, service *model.Service) error {
	return s.store.CreateService(ctx, service)
}

func (s *AppService) ListServices(ctx context.Context, filter repo.ServiceFilter) ([]model.Service, error) {
	return s.store.ListServices(ctx, filter)
}

func (s *AppService) GetService(ctx context.Context, id bson.ObjectID) (*model.Service, error) {
	return s.store.GetServiceByID(ctx, id)
}

func (s *AppService) UpdateService(ctx context.Context, service *model.Service) error {
	return s.store.UpdateService(ctx, service)
}

func (s *AppService) DeleteService(ctx context.Context, id bson.ObjectID, providerID bson.ObjectID) error {
	return s.store.DeleteService(ctx, id, providerID)
}

func (s *AppService) CreateOrder(ctx context.Context, order *model.Order) error {
	return s.store.CreateOrder(ctx, order)
}

func (s *AppService) ListOrders(ctx context.Context, filter repo.OrderFilter) ([]model.Order, error) {
	return s.store.ListOrders(ctx, filter)
}

func (s *AppService) GetOrder(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	return s.store.GetOrderByID(ctx, id)
}

func (s *AppService) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, actorID bson.ObjectID) error {
	return s.store.UpdateOrderStatus(ctx, id, status, actorID)
}

func (s *AppService) AddOrderMessage(ctx context.Context, id bson.ObjectID, msg model.OrderMessage, actorID bson.ObjectID) error {
	return s.store.AddOrderMessage(ctx, id, msg, actorID)
}
