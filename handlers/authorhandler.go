package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	m "project.com/myproject/models"
)

// Handle Authors
func (h *Handler) HandleAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		log.Println("Request cancelled")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
		var wg sync.WaitGroup
		switch r.Method {
		case http.MethodGet:
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Check if search parameters are provided
				firstName := r.URL.Query().Get("first_name")
				lastName := r.URL.Query().Get("last_name")

				if firstName != "" || lastName != "" {
					h.handleSearchAuthors(w, r)
				} else {
					h.handleGetAllAuthors(w, r)
				}
			}()
		case http.MethodPost:
			wg.Add(1)
			go func() {
				defer wg.Done()
				h.handleCreateAuthor(w, r)
			}()
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		wg.Wait()
	}
}

func (h *Handler) HandleAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid author ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetAuthor(w, id, w)
	case http.MethodPut:
		h.handleUpdateAuthor(w, r, id)
	case http.MethodDelete:
		h.handleDeleteAuthor(w, id, w)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// CRUD operations for Authors
func (h *Handler) handleGetAllAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := h.Store.GetAllAuthors()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve authors")
		return
	}
	h.respondWithJSON(w, http.StatusOK, authors)
}

func (h *Handler) handleGetAuthor(w http.ResponseWriter, id int, res http.ResponseWriter) {
	author, err := h.Store.GetAuthor(id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Author not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, author)
}

func (h *Handler) handleCreateAuthor(w http.ResponseWriter, r *http.Request) {
	var author m.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	newAuthor, err := h.Store.CreateAuthor(author)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create author")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newAuthor)
}

func (h *Handler) handleUpdateAuthor(w http.ResponseWriter, r *http.Request, id int) {
	var author m.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateAuthor(id, author)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update author")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Author updated successfully")
}

func (h *Handler) handleDeleteAuthor(w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteAuthor(id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete author")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Author deleted successfully")
}

func (h *Handler) handleSearchAuthors(w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaAuthors{
		FirstName: r.URL.Query().Get("first_name"),
		LastName:  r.URL.Query().Get("last_name"),
	}

	authors, err := h.Store.SearchAuthors(criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No authors found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search authors")
		return
	}

	h.respondWithJSON(w, http.StatusOK, authors)
}
