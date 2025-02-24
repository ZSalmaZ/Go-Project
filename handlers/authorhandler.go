package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	m "project.com/myproject/models"
)

// ✅ Handle Authors with Goroutines and Cancellation
func (h *Handler) HandleAuthors(w http.ResponseWriter, r *http.Request) {
	// Set a request timeout of 5 seconds
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("❌ Request cancelled")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
		switch r.Method {
		case http.MethodGet:
			firstName := r.URL.Query().Get("first_name")
			lastName := r.URL.Query().Get("last_name")

			if firstName != "" || lastName != "" {
				h.handleSearchAuthors(ctx, w, r)
			} else {
				h.handleGetAllAuthors(ctx, w, r)
			}
		case http.MethodPost:
			h.handleCreateAuthor(ctx, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) HandleAuthor(w http.ResponseWriter, r *http.Request) {
	// Set a request timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("❌ Request cancelled")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid author ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetAuthor(ctx, w, id, w)
		case http.MethodPut:
			h.handleUpdateAuthor(ctx, w, r, id)
		case http.MethodDelete:
			h.handleDeleteAuthor(ctx, w, id, w)
		default:
			h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// CRUD operations for Authors
func (h *Handler) handleGetAllAuthors(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	authors, err := h.Store.GetAllAuthors(ctx) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve authors")
		return
	}
	h.respondWithJSON(w, http.StatusOK, authors)
}

func (h *Handler) handleGetAuthor(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	author, err := h.Store.GetAuthor(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Author not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, author)
}

func (h *Handler) handleCreateAuthor(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var author m.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newAuthor, err := h.Store.CreateAuthor(ctx, author) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create author")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newAuthor)
}

func (h *Handler) handleUpdateAuthor(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var author m.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.Store.UpdateAuthor(ctx, id, author) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update author")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Author updated successfully")
}

func (h *Handler) handleDeleteAuthor(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteAuthor(ctx, id) // ✅ Now uses context
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete author")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Author deleted successfully")
}

func (h *Handler) handleSearchAuthors(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaAuthors{
		FirstName: r.URL.Query().Get("first_name"),
		LastName:  r.URL.Query().Get("last_name"),
	}

	authors, err := h.Store.SearchAuthors(ctx, criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No authors found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search authors")
		return
	}

	h.respondWithJSON(w, http.StatusOK, authors)
}
