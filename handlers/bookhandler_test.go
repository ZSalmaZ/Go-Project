// handlers/bookhandler_test.go
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

func setupTestRouter(t *testing.T) (*mux.Router, *Handler) {
	db, err := sql.Open("postgres", "postgres://postgres:secret@localhost:5432/bookstore?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	store := stores.NewPostgresStore(db)

	// Mock JWT for testing
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)

	h := NewHandler(store)
	r := mux.NewRouter()

	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Middleware)
	protected.HandleFunc("/books", h.HandleBooks).Methods("GET", "POST")

	return r, h
}

func TestHandleGetBooks(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Mock JWT token
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	req := httptest.NewRequest("GET", "/api/books", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}
}

func TestHandleCreateBook(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Mock JWT token
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	book := m.Book{
		Title:       "API Test Book",
		Author:      m.Author{ID: 1}, // Existing author
		PublishedAt: time.Now(),
		Price:       15.99,
		Stock:       5,
		Genres:      []string{"Science Fiction"},
	}
	body, _ := json.Marshal(book)

	req := httptest.NewRequest("POST", "/api/books", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", rec.Code)
	}
}

func TestHandleGetBooks_Unauthorized(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/books", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 Unauthorized, got %d", rec.Code)
	}
}
