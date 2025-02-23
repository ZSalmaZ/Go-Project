package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	m "project.com/myproject/models"
)

// Handle Books
func (h *Handler) HandleBooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Check if search parameters are provided
		title := r.URL.Query().Get("title")
		authorFirstName := r.URL.Query().Get("author_first_name")
		authorLastName := r.URL.Query().Get("author_last_name")

		if title != "" || authorFirstName != "" || authorLastName != "" {
			h.handleSearchBooks(w, r)
		} else {
			h.handleGetAllBooks(w, r)
		}
	case http.MethodPost:
		h.handleCreateBook(w, r)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) HandleBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetBook(w, id, w)
	case http.MethodPut:
		h.handleUpdateBook(w, r, id)
	case http.MethodDelete:
		h.handleDeleteBook(w, id, w)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
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
func (h *Handler) handleGetAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.Store.GetAllBooks()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve books")
		return
	}
	h.respondWithJSON(w, http.StatusOK, books)
}

func (h *Handler) handleGetBook(w http.ResponseWriter, id int, res http.ResponseWriter) {
	book, err := h.Store.GetBook(id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Book not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, book)
}

func (h *Handler) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	newBook, err := h.Store.CreateBook(book)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create book")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newBook)
}

func (h *Handler) handleUpdateBook(w http.ResponseWriter, r *http.Request, id int) {
	var book m.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateBook(id, book)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update book")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Book updated successfully")
}

func (h *Handler) handleDeleteBook(w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteBook(id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete book")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Book deleted successfully")
}

func (h *Handler) handleSearchBooks(w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaBooks{
		Title:           r.URL.Query().Get("title"),
		AuthorFirstName: r.URL.Query().Get("author_first_name"),
		AuthorName:      r.URL.Query().Get("author_last_name"),
	}

	books, err := h.Store.SearchBooks(criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No books found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search books")
		return
	}

	h.respondWithJSON(w, http.StatusOK, books)
}
