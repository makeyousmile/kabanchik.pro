package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kabanchik.pro/internal/auth"
	"kabanchik.pro/internal/model"
	"kabanchik.pro/internal/repo"
	"kabanchik.pro/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type memoryStore struct {
	users    map[bson.ObjectID]*model.User
	services map[bson.ObjectID]*model.Service
	orders   map[bson.ObjectID]*model.Order
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:    make(map[bson.ObjectID]*model.User),
		services: make(map[bson.ObjectID]*model.Service),
		orders:   make(map[bson.ObjectID]*model.Order),
	}
}

func (m *memoryStore) CreateUser(ctx context.Context, user *model.User) error {
	user.ID = bson.NewObjectID()
	m.users[user.ID] = user
	return nil
}

func (m *memoryStore) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (m *memoryStore) FindUserByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return user, nil
}

func (m *memoryStore) UpdateUser(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *memoryStore) CreateService(ctx context.Context, service *model.Service) error {
	service.ID = bson.NewObjectID()
	m.services[service.ID] = service
	return nil
}

func (m *memoryStore) ListServices(ctx context.Context, filter repo.ServiceFilter) ([]model.Service, error) {
	items := make([]model.Service, 0, len(m.services))
	for _, svc := range m.services {
		items = append(items, *svc)
	}
	return items, nil
}

func (m *memoryStore) GetServiceByID(ctx context.Context, id bson.ObjectID) (*model.Service, error) {
	svc, ok := m.services[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return svc, nil
}

func (m *memoryStore) UpdateService(ctx context.Context, service *model.Service) error {
	m.services[service.ID] = service
	return nil
}

func (m *memoryStore) DeleteService(ctx context.Context, id bson.ObjectID, providerID bson.ObjectID) error {
	delete(m.services, id)
	return nil
}

func (m *memoryStore) CreateOrder(ctx context.Context, order *model.Order) error {
	order.ID = bson.NewObjectID()
	m.orders[order.ID] = order
	return nil
}

func (m *memoryStore) ListOrders(ctx context.Context, filter repo.OrderFilter) ([]model.Order, error) {
	items := make([]model.Order, 0, len(m.orders))
	for _, order := range m.orders {
		items = append(items, *order)
	}
	return items, nil
}

func (m *memoryStore) GetOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return order, nil
}

func (m *memoryStore) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, actorID bson.ObjectID) error {
	order, ok := m.orders[id]
	if !ok {
		return repo.ErrNotFound
	}
	order.Status = status
	m.orders[id] = order
	return nil
}

func (m *memoryStore) AddOrderMessage(ctx context.Context, id bson.ObjectID, msg model.OrderMessage, actorID bson.ObjectID) error {
	order, ok := m.orders[id]
	if !ok {
		return repo.ErrNotFound
	}
	order.Messages = append(order.Messages, msg)
	m.orders[id] = order
	return nil
}

func TestAuthFlow(t *testing.T) {
	store := newMemoryStore()
	svc := service.New(store)
	secret := []byte("test-secret")
	api := NewAPI(svc, secret, time.Hour)

	mux := http.NewServeMux()
	api.Register(mux)

	registerBody := map[string]any{"email": "u1@example.com", "password": "secret", "role": "client"}
	buf, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(buf))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register code: %d", rec.Code)
	}

	loginBody := map[string]any{"email": "u1@example.com", "password": "secret"}
	buf, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(buf))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login code: %d", rec.Code)
	}

	var loginResp struct {
		Token string `json:"token"`
		User  model.User `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login json: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatalf("expected token")
	}

	_, err := auth.ParseToken(loginResp.Token, secret)
	if err != nil {
		t.Fatalf("token parse: %v", err)
	}
}
