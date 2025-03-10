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

func setupCustomerTestRouter(t *testing.T) (*mux.Router, *Handler) {
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
	protected.HandleFunc("/customers", h.HandleCustomers).Methods("GET", "POST")
	protected.HandleFunc("/customers/{id}", h.HandleCustomer).Methods("GET", "PUT", "DELETE")
	return r, h
}

func TestHandleGetCustomers(t *testing.T) {
	router, _ := setupCustomerTestRouter(t)
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	req := httptest.NewRequest("GET", "/api/customers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}
}

func TestHandleCreateCustomer(t *testing.T) {
	router, _ := setupCustomerTestRouter(t)
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	customer := m.Customer{
		Name:       "Test User",
		Email:      "testuser@example.com",
		Street:     "456 Road",
		City:       "Townsville",
		State:      "TS",
		PostalCode: "67890",
		Country:    "USA",
	}
	body, _ := json.Marshal(customer)

	req := httptest.NewRequest("POST", "/api/customers", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", rec.Code)
	}
}
