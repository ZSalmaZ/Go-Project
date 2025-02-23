package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	m "project.com/myproject/models"
)

// Handle Orders
func (h *Handler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Check if search parameters are provided
		customerName := r.URL.Query().Get("customer_name")
		status := r.URL.Query().Get("status")

		if customerName != "" || status != "" {
			h.handleSearchOrders(w, r)
		} else {
			h.handleGetAllOrders(w, r)
		}
	case http.MethodPost:
		h.handleCreateOrder(w, r)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) HandleOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetOrder(w, id, w)
	case http.MethodPut:
		h.handleUpdateOrder(w, r, id)
	case http.MethodDelete:
		h.handleDeleteOrder(w, id, w)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// CRUD operations for Orders
func (h *Handler) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.Store.GetAllOrders()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve orders")
		return
	}
	h.respondWithJSON(w, http.StatusOK, orders)
}

func (h *Handler) handleGetOrder(w http.ResponseWriter, id int, res http.ResponseWriter) {
	order, err := h.Store.GetOrder(id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Order not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, order)
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var order m.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate the order data
	if order.Customer.ID == 0 || len(order.Items) == 0 || order.TotalPrice <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Missing required order fields")
		return
	}

	// Check if the customer exists in the system
	customerExists, _ := h.Store.GetCustomer(order.Customer.ID)
	if customerExists.ID == 0 {
		h.respondWithError(w, http.StatusBadRequest, "Customer does not exist")
		return
	}

	newOrder, err := h.Store.CreateOrder(order)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newOrder)
}

func (h *Handler) handleUpdateOrder(w http.ResponseWriter, r *http.Request, id int) {
	var order m.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateOrder(id, order)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Order updated successfully")
}

func (h *Handler) handleDeleteOrder(w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteOrder(id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete order")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Order deleted successfully")
}

func (h *Handler) handleSearchOrders(w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaOrders{
		CustomerName: r.URL.Query().Get("customer_name"),
		Status:       r.URL.Query().Get("status"),
	}

	orders, err := h.Store.SearchOrders(criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No orders found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search orders")
		return
	}

	h.respondWithJSON(w, http.StatusOK, orders)
}
