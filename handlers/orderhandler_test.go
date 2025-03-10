package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"project.com/myproject/auth"
	m "project.com/myproject/models"
	"project.com/myproject/stores"
)

func setupOrderTestRouter(t *testing.T) (*mux.Router, *Handler) {
	db, err := sql.Open("postgres", "postgres://postgres:secret@localhost:5432/bookstore?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	store := stores.NewPostgresStore(db)
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)
	h := NewHandler(store)
	r := mux.NewRouter()
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Middleware)
	protected.HandleFunc("/orders", h.HandleOrders).Methods("GET", "POST")
	protected.HandleFunc("/orders/{id}", h.HandleOrder).Methods("GET", "PUT", "DELETE")
	return r, h
}

func TestHandleGetOrders(t *testing.T) {
	router, _ := setupOrderTestRouter(t)
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	req := httptest.NewRequest("GET", "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}
}

func TestHandleCreateOrder(t *testing.T) {
	router, _ := setupOrderTestRouter(t)
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	order := m.Order{
		Customer:   m.Customer{ID: 1},
		TotalPrice: 39.99,
		Status:     "Pending",
		Items: []m.OrderItem{
			{Book: m.Book{ID: 1}, Quantity: 1},
		},
	}
	body, _ := json.Marshal(order)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", rec.Code)
	}
}
