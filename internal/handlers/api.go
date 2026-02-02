package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kabanchik.pro/internal/auth"
	"kabanchik.pro/internal/model"
	"kabanchik.pro/internal/repo"
	"kabanchik.pro/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type API struct {
	svc       *service.AppService
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAPI(svc *service.AppService, jwtSecret []byte, jwtTTL time.Duration) *API {
	return &API{svc: svc, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health", health)

	mux.HandleFunc("/api/v1/auth/register", api.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", api.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", api.handleLogout)

	mux.HandleFunc("/api/v1/me", api.handleMe)

	mux.HandleFunc("/api/v1/services", api.handleServices)
	mux.HandleFunc("/api/v1/services/", api.handleServiceByID)

	mux.HandleFunc("/api/v1/orders", api.handleOrders)
	mux.HandleFunc("/api/v1/orders/", api.handleOrderByID)
}

func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string         `json:"email"`
		Password string         `json:"password"`
		Role     model.UserRole `json:"role"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := api.svc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (api *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := api.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.CreateToken(user.ID.Hex(), string(user.Role), api.jwtSecret, api.jwtTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	resp := map[string]any{
		"token": token,
		"user":  user,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := api.requireUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, user)
	case http.MethodPatch:
		var req struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
			City  string `json:"city"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		user.Name = strings.TrimSpace(req.Name)
		user.Phone = strings.TrimSpace(req.Phone)
		user.City = strings.TrimSpace(req.City)
		if err := api.svc.UpdateUser(r.Context(), user); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, user)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		filter := repo.ServiceFilter{
			Category: r.URL.Query().Get("category"),
			City:     r.URL.Query().Get("city"),
			Query:    r.URL.Query().Get("q"),
		}
		filter.MinPrice, _ = parseInt64(r.URL.Query().Get("min_price"))
		filter.MaxPrice, _ = parseInt64(r.URL.Query().Get("max_price"))

		services, err := api.svc.ListServices(r.Context(), filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, services)
	case http.MethodPost:
		user, err := api.requireRole(r, model.RoleProvider)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Category    string `json:"category"`
			City        string `json:"city"`
			Price       int64  `json:"price"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		service := &model.Service{
			ProviderID:  user.ID,
			Title:       strings.TrimSpace(req.Title),
			Description: strings.TrimSpace(req.Description),
			Category:    strings.TrimSpace(req.Category),
			City:        strings.TrimSpace(req.City),
			Price:       req.Price,
		}
		if service.Title == "" {
			writeError(w, http.StatusBadRequest, "title required")
			return
		}
		if err := api.svc.CreateService(r.Context(), service); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, service)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) handleServiceByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseObjectID(r.URL.Path, "/api/v1/services/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		serviceItem, err := api.svc.GetService(r.Context(), id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, serviceItem)
	case http.MethodPatch:
		user, err := api.requireRole(r, model.RoleProvider)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Category    string `json:"category"`
			City        string `json:"city"`
			Price       int64  `json:"price"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		serviceItem := &model.Service{
			ID:          id,
			ProviderID:  user.ID,
			Title:       strings.TrimSpace(req.Title),
			Description: strings.TrimSpace(req.Description),
			Category:    strings.TrimSpace(req.Category),
			City:        strings.TrimSpace(req.City),
			Price:       req.Price,
		}
		if err := api.svc.UpdateService(r.Context(), serviceItem); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, serviceItem)
	case http.MethodDelete:
		user, err := api.requireRole(r, model.RoleProvider)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if err := api.svc.DeleteService(r.Context(), id, user.ID); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) handleOrders(w http.ResponseWriter, r *http.Request) {
	user, err := api.requireUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		filter := repo.OrderFilter{}
		if status := r.URL.Query().Get("status"); status != "" {
			filter.Status = model.OrderStatus(status)
		}
		if user.Role == model.RoleProvider {
			filter.ProviderID = &user.ID
		} else {
			filter.ClientID = &user.ID
		}
		orders, err := api.svc.ListOrders(r.Context(), filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, orders)
	case http.MethodPost:
		if user.Role != model.RoleClient {
			writeError(w, http.StatusForbidden, "only clients can create orders")
			return
		}
		var req struct {
			ServiceID string `json:"service_id"`
			Details   string `json:"details"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		serviceID, err := bson.ObjectIDFromHex(req.ServiceID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid service_id")
			return
		}
		serviceItem, err := api.svc.GetService(r.Context(), serviceID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "service not found")
			return
		}
		order := &model.Order{
			ServiceID:  serviceItem.ID,
			ClientID:   user.ID,
			ProviderID: serviceItem.ProviderID,
			Details:    strings.TrimSpace(req.Details),
		}
		if err := api.svc.CreateOrder(r.Context(), order); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, order)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseObjectID(r.URL.Path, "/api/v1/orders/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := api.requireUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		order, err := api.svc.GetOrder(r.Context(), id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		if !api.canAccessOrder(user, order) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodPatch:
		order, err := api.svc.GetOrder(r.Context(), id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		if !api.canAccessOrder(user, order) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req struct {
			Status model.OrderStatus `json:"status"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Status == "" {
			writeError(w, http.StatusBadRequest, "status required")
			return
		}
		if !api.canUpdateOrderStatus(user, order, req.Status) {
			writeError(w, http.StatusForbidden, "status transition not allowed")
			return
		}
		if err := api.svc.UpdateOrderStatus(r.Context(), id, req.Status, user.ID); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repo.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/message") {
			order, err := api.svc.GetOrder(r.Context(), id)
			if err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, repo.ErrNotFound) {
					status = http.StatusNotFound
				}
				writeError(w, status, err.Error())
				return
			}
			if !api.canAccessOrder(user, order) {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
			var req struct {
				Text string `json:"text"`
			}
			if err := readJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			msg := model.OrderMessage{SenderID: user.ID, Text: strings.TrimSpace(req.Text)}
			if msg.Text == "" {
				writeError(w, http.StatusBadRequest, "text required")
				return
			}
			if err := api.svc.AddOrderMessage(r.Context(), id, msg, user.ID); err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, repo.ErrNotFound) {
					status = http.StatusNotFound
				}
				writeError(w, status, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusBadRequest, "unknown action")
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *API) requireUser(r *http.Request) (*model.User, error) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
	if token == "" {
		return nil, service.ErrUnauthorized
	}
	claims, err := auth.ParseToken(strings.TrimSpace(token), api.jwtSecret)
	if err != nil {
		return nil, service.ErrUnauthorized
	}
	id, err := bson.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, service.ErrUnauthorized
	}
	user, err := api.svc.GetUser(r.Context(), id)
	if err != nil {
		return nil, service.ErrUnauthorized
	}
	return user, nil
}

func (api *API) requireRole(r *http.Request, role model.UserRole) (*model.User, error) {
	user, err := api.requireUser(r)
	if err != nil {
		return nil, err
	}
	if user.Role != role {
		return nil, service.ErrUnauthorized
	}
	return user, nil
}

func (api *API) canAccessOrder(user *model.User, order *model.Order) bool {
	if user.Role == model.RoleProvider {
		return user.ID == order.ProviderID
	}
	return user.ID == order.ClientID
}

func (api *API) canUpdateOrderStatus(user *model.User, order *model.Order, next model.OrderStatus) bool {
	switch user.Role {
	case model.RoleProvider:
		if user.ID != order.ProviderID {
			return false
		}
		return (order.Status == model.OrderNew && next == model.OrderAccepted) ||
			(order.Status == model.OrderAccepted && next == model.OrderDone)
	case model.RoleClient:
		if user.ID != order.ClientID {
			return false
		}
		return next == model.OrderCanceled
	default:
		return false
	}
}

func readJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseObjectID(path, prefix string) (bson.ObjectID, error) {
	idPart := strings.TrimPrefix(path, prefix)
	idPart = strings.TrimSuffix(idPart, "/message")
	return bson.ObjectIDFromHex(strings.Trim(idPart, "/"))
}
