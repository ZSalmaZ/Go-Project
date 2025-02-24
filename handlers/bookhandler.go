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

// Handle Books
func (h *Handler) HandleBooks(w http.ResponseWriter, r *http.Request) {
	// Set a request timeout (5 seconds max for execution)
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
			// Run search books in a goroutine only if search params are present
			title := r.URL.Query().Get("title")
			authorFirstName := r.URL.Query().Get("author_first_name")
			authorLastName := r.URL.Query().Get("author_last_name")

			if title != "" || authorFirstName != "" || authorLastName != "" {
				h.handleSearchBooks(ctx, w, r) // ✅ Run search in a separate thread
			} else {
				h.handleGetAllBooks(ctx, w, r) // Direct call for normal GET (faster)
			}

		case http.MethodPost:
			h.handleCreateBook(ctx, w, r)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) HandleBook(w http.ResponseWriter, r *http.Request) {
	// Create a context with a 5-second timeout
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
			h.respondWithError(w, http.StatusBadRequest, "Invalid book ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetBook(ctx, w, id, w)

		case http.MethodPut:
			h.handleUpdateBook(ctx, w, r, id)

		case http.MethodDelete:
			h.handleDeleteBook(ctx, w, id, w)

		default:
			h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// Generic Response Helpers
func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, map[string]string{"error": message})
}

// Implement CRUD operations for Books
func (h *Handler) handleGetAllBooks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	books, err := h.Store.GetAllBooks(ctx) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve books")
		return
	}

	h.respondWithJSON(w, http.StatusOK, books)
}

func (h *Handler) handleGetBook(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	book, err := h.Store.GetBook(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Book not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, book)
}

func (h *Handler) handleCreateBook(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newBook, err := h.Store.CreateBook(ctx, book) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create book")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, newBook)
}

func (h *Handler) handleUpdateBook(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.Store.UpdateBook(ctx, id, book) // ✅ Now uses context
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update book")
		return
	}

	h.respondWithJSON(w, http.StatusOK, "Book updated successfully")
}

func (h *Handler) handleDeleteBook(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteBook(ctx, id) // ✅ Now uses context
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete book")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Book deleted successfully")
}

// func (h *Handler) handleSearchBooks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
// 	criteria := m.SearchCriteriaBooks{
// 		Title:           r.URL.Query().Get("title"),
// 		AuthorFirstName: r.URL.Query().Get("author_first_name"),
// 		AuthorName:      r.URL.Query().Get("author_last_name"),
// 	}

// 	books, err := h.Store.SearchBooks(ctx, criteria) // ✅ Now uses context
// 	if err == sql.ErrNoRows {
// 		h.respondWithError(w, http.StatusNotFound, "No books found")
// 		return
// 	} else if err != nil {
// 		h.respondWithError(w, http.StatusInternalServerError, "Failed to search books")
// 		return
// 	}

// 	h.respondWithJSON(w, http.StatusOK, books)
// }

func (h *Handler) handleSearchBooks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	authorFirstName := r.URL.Query().Get("author_first_name")
	authorName := r.URL.Query().Get("author_last_name")

	// Parse min_price and max_price query parameters.
	minPriceStr := r.URL.Query().Get("min_price")
	maxPriceStr := r.URL.Query().Get("max_price")
	var minPrice, maxPrice float64
	var err error
	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid min_price")
			return
		}
	}
	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid max_price")
			return
		}
	}

	criteria := m.SearchCriteriaBooks{
		Title:           title,
		AuthorFirstName: authorFirstName,
		AuthorName:      authorName,
		MinPrice:        minPrice,
		MaxPrice:        maxPrice,
	}

	books, err := h.Store.SearchBooks(ctx, criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No books found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search books")
		return
	}

	h.respondWithJSON(w, http.StatusOK, books)
}
