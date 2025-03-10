// handlers/authorhandler_test.go
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

func setupAuthorTestRouter(t *testing.T) (*mux.Router, *Handler) {
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
	protected.HandleFunc("/authors", h.HandleAuthors).Methods("GET", "POST")
	protected.HandleFunc("/authors/{id}", h.HandleAuthor).Methods("GET", "PUT", "DELETE")

	return r, h
}

func TestHandleGetAuthors(t *testing.T) {
	router, _ := setupAuthorTestRouter(t)

	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	req := httptest.NewRequest("GET", "/api/authors", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}
}

func TestHandleCreateAuthor(t *testing.T) {
	router, _ := setupAuthorTestRouter(t)

	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	token, _ := jwtManager.Generate("testuser")

	author := m.Author{
		FirstName: "Jane",
		LastName:  "Smith",
		Bio:       "An aspiring novelist.",
	}
	body, _ := json.Marshal(author)

	req := httptest.NewRequest("POST", "/api/authors", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", rec.Code)
	}
}

func TestHandleGetAuthors_Unauthorized(t *testing.T) {
	router, _ := setupAuthorTestRouter(t)

	req := httptest.NewRequest("GET", "/api/authors", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 Unauthorized, got %d", rec.Code)
	}
}
