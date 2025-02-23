package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	m "project.com/myproject/models"
)

// Handle Customers
func (h *Handler) HandleCustomers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Check if search parameters are provided
		name := r.URL.Query().Get("name")
		email := r.URL.Query().Get("email")

		if name != "" || email != "" {
			h.handleSearchCustomers(w, r)
		} else {
			h.handleGetAllCustomers(w, r)
		}
	case http.MethodPost:
		h.handleCreateCustomer(w, r)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) HandleCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetCustomer(w, id, w)
	case http.MethodPut:
		h.handleUpdateCustomer(w, r, id)
	case http.MethodDelete:
		h.handleDeleteCustomer(w, id, w)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// CRUD operations for Customers
func (h *Handler) handleGetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.Store.GetAllCustomers()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve customers")
		return
	}
	h.respondWithJSON(w, http.StatusOK, customers)
}

func (h *Handler) handleGetCustomer(w http.ResponseWriter, id int, res http.ResponseWriter) {
	customer, err := h.Store.GetCustomer(id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Customer not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, customer)
}

func (h *Handler) handleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	newCustomer, err := h.Store.CreateCustomer(customer)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newCustomer)
}

func (h *Handler) handleUpdateCustomer(w http.ResponseWriter, r *http.Request, id int) {
	var customer m.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateCustomer(id, customer)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update customer")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Customer updated successfully")
}

func (h *Handler) handleDeleteCustomer(w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteCustomer(id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete customer")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Customer deleted successfully")
}

func (h *Handler) handleSearchCustomers(w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaCustomers{
		Name:  r.URL.Query().Get("name"),
		Email: r.URL.Query().Get("email"),
	}

	customers, err := h.Store.SearchCustomers(criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No customers found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search customers")
		return
	}

	h.respondWithJSON(w, http.StatusOK, customers)
}
