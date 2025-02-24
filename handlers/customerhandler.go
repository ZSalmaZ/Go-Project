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

// Handle Customers
func (h *Handler) HandleCustomers(w http.ResponseWriter, r *http.Request) {
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
			name := r.URL.Query().Get("name")
			email := r.URL.Query().Get("email")

			if name != "" || email != "" {
				h.handleSearchCustomers(ctx, w, r)
			} else {
				h.handleGetAllCustomers(ctx, w, r)
			}
		case http.MethodPost:
			h.handleCreateCustomer(ctx, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) HandleCustomer(w http.ResponseWriter, r *http.Request) {
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
			h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetCustomer(ctx, w, id, w)
		case http.MethodPut:
			h.handleUpdateCustomer(ctx, w, r, id)
		case http.MethodDelete:
			h.handleDeleteCustomer(ctx, w, id, w)
		default:
			h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// CRUD operations for Customers
func (h *Handler) handleGetAllCustomers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	customers, err := h.Store.GetAllCustomers(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve customers")
		return
	}
	h.respondWithJSON(w, http.StatusOK, customers)
}

func (h *Handler) handleGetCustomer(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	customer, err := h.Store.GetCustomer(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Customer not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, customer)
}

func (h *Handler) handleCreateCustomer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	newCustomer, err := h.Store.CreateCustomer(ctx, customer)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newCustomer)
}

func (h *Handler) handleUpdateCustomer(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var customer m.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateCustomer(ctx, id, customer)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update customer")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Customer updated successfully")
}

func (h *Handler) handleDeleteCustomer(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteCustomer(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete customer")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Customer deleted successfully")
}

func (h *Handler) handleSearchCustomers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaCustomers{
		Name:  r.URL.Query().Get("name"),
		Email: r.URL.Query().Get("email"),
	}

	customers, err := h.Store.SearchCustomers(ctx, criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No customers found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search customers")
		return
	}

	h.respondWithJSON(w, http.StatusOK, customers)
}
