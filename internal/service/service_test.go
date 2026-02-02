package service

import (
	"context"
	"errors"
	"testing"

	"kabanchik.pro/internal/model"
	"kabanchik.pro/internal/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeStore struct {
	usersByEmail map[string]*model.User
}

func newFakeStore() *fakeStore {
	return &fakeStore{usersByEmail: make(map[string]*model.User)}
}

func (f *fakeStore) CreateUser(ctx context.Context, user *model.User) error {
	if _, exists := f.usersByEmail[user.Email]; exists {
		return errors.New("exists")
	}
	user.ID = bson.NewObjectID()
	f.usersByEmail[user.Email] = user
	return nil
}

func (f *fakeStore) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, ok := f.usersByEmail[email]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return user, nil
}

func (f *fakeStore) FindUserByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	for _, user := range f.usersByEmail {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (f *fakeStore) UpdateUser(ctx context.Context, user *model.User) error { return nil }

func (f *fakeStore) CreateService(ctx context.Context, service *model.Service) error { return nil }
func (f *fakeStore) ListServices(ctx context.Context, filter repo.ServiceFilter) ([]model.Service, error) {
	return nil, nil
}
func (f *fakeStore) GetServiceByID(ctx context.Context, id bson.ObjectID) (*model.Service, error) {
	return nil, repo.ErrNotFound
}
func (f *fakeStore) UpdateService(ctx context.Context, service *model.Service) error { return nil }
func (f *fakeStore) DeleteService(ctx context.Context, id bson.ObjectID, providerID bson.ObjectID) error {
	return nil
}

func (f *fakeStore) CreateOrder(ctx context.Context, order *model.Order) error { return nil }
func (f *fakeStore) ListOrders(ctx context.Context, filter repo.OrderFilter) ([]model.Order, error) {
	return nil, nil
}
func (f *fakeStore) GetOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	return nil, repo.ErrNotFound
}
func (f *fakeStore) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, actorID bson.ObjectID) error {
	return nil
}
func (f *fakeStore) AddOrderMessage(ctx context.Context, id bson.ObjectID, msg model.OrderMessage, actorID bson.ObjectID) error {
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	store := newFakeStore()
	svc := New(store)

	user, err := svc.Register(context.Background(), "test@example.com", "secret", model.RoleClient)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.ID.IsZero() {
		t.Fatal("expected user id")
	}

	_, err = svc.Login(context.Background(), "test@example.com", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	_, err = svc.Login(context.Background(), "test@example.com", "wrong")
	if err == nil {
		t.Fatal("expected login error")
	}
}
