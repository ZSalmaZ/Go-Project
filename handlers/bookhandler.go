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
	"project.com/myproject/internal/cache" // Import cache package
	m "project.com/myproject/models"
)

// Handle Books
func (h *Handler) HandleBooks(w http.ResponseWriter, r *http.Request) {
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
			title := r.URL.Query().Get("title")
			authorFirstName := r.URL.Query().Get("author_first_name")
			authorLastName := r.URL.Query().Get("author_last_name")

			if title != "" || authorFirstName != "" || authorLastName != "" {
				h.handleSearchBooks(ctx, w, r)
			} else {
				h.handleGetAllBooks(ctx, w, r)
			}

		case http.MethodPost:
			h.handleCreateBook(ctx, w, r)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) HandleBook(w http.ResponseWriter, r *http.Request) {
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

// respondWithJSON writes the given data as JSON to the response writer
func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondWithError writes an error message as JSON to the response writer
func (h *Handler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, map[string]string{"error": message})
}

// Modify handleGetAllBooks to use caching
func (h *Handler) handleGetAllBooks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cacheKey := "all_books"

	// Check Redis cache
	cachedData, err := cache.GetCache(cacheKey)
	if err == nil {
		log.Println("✅ Cache hit: Returning books from Redis")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedData))
		return
	}

	// If cache miss, fetch from DB
	books, err := h.Store.GetAllBooks(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve books")
		return
	}

	// Store result in cache for 10 minutes
	jsonData, _ := json.Marshal(books)
	cache.SetCache(cacheKey, string(jsonData), 10*time.Minute)

	h.respondWithJSON(w, http.StatusOK, books)
}

// Modify handleGetBook to use caching
func (h *Handler) handleGetBook(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	cacheKey := "book:" + strconv.Itoa(id)

	// Check Redis cache
	cachedData, err := cache.GetCache(cacheKey)
	if err == nil {
		log.Println("✅ Cache hit: Returning book from Redis")
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(cachedData))
		return
	}

	// If cache miss, fetch from DB
	book, err := h.Store.GetBook(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Book not found")
		return
	}

	// Store result in cache for 10 minutes
	jsonData, _ := json.Marshal(book)
	cache.SetCache(cacheKey, string(jsonData), 10*time.Minute)

	h.respondWithJSON(res, http.StatusOK, book)
}

// Modify handleCreateBook to invalidate cache
func (h *Handler) handleCreateBook(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newBook, err := h.Store.CreateBook(ctx, book)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create book")
		return
	}

	// Invalidate book list cache
	cache.DeleteCache("all_books")

	h.respondWithJSON(w, http.StatusCreated, newBook)
}

// Modify handleUpdateBook to invalidate cache
func (h *Handler) handleUpdateBook(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.Store.UpdateBook(ctx, id, book)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update book")
		return
	}

	// Invalidate both individual book and book list cache
	cache.DeleteCache("book:" + strconv.Itoa(id))
	cache.DeleteCache("all_books")

	h.respondWithJSON(w, http.StatusOK, "Book updated successfully")
}

// Modify handleDeleteBook to invalidate cache
func (h *Handler) handleDeleteBook(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteBook(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete book")
		return
	}

	// Invalidate both individual book and book list cache
	cache.DeleteCache("book:" + strconv.Itoa(id))
	cache.DeleteCache("all_books")

	h.respondWithJSON(res, http.StatusOK, "Book deleted successfully")
}

// Modify handleSearchBooks to use caching
func (h *Handler) handleSearchBooks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	authorFirstName := r.URL.Query().Get("author_first_name")
	authorName := r.URL.Query().Get("author_last_name")

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

	cacheKey := "search_books:" + title + ":" + authorFirstName + ":" + authorName
	cachedData, err := cache.GetCache(cacheKey)
	if err == nil {
		log.Println("✅ Cache hit: Returning search results from Redis")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedData))
		return
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

	jsonData, _ := json.Marshal(books)
	cache.SetCache(cacheKey, string(jsonData), 10*time.Minute)

	h.respondWithJSON(w, http.StatusOK, books)
}
