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

// Handle Orders
func (h *Handler) HandleOrders(w http.ResponseWriter, r *http.Request) {
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
			customerName := r.URL.Query().Get("customer_name")
			status := r.URL.Query().Get("status")

			if customerName != "" || status != "" {
				h.handleSearchOrders(ctx, w, r)
			} else {
				h.handleGetAllOrders(ctx, w, r)
			}
		case http.MethodPost:
			h.handleCreateOrder(ctx, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) HandleOrder(w http.ResponseWriter, r *http.Request) {
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
			h.respondWithError(w, http.StatusBadRequest, "Invalid order ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGetOrder(ctx, w, id, w)
		case http.MethodPut:
			h.handleUpdateOrder(ctx, w, r, id)
		case http.MethodDelete:
			h.handleDeleteOrder(ctx, w, id, w)
		default:
			h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// CRUD operations for Orders
func (h *Handler) handleGetAllOrders(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	orders, err := h.Store.GetAllOrders(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve orders")
		return
	}
	h.respondWithJSON(w, http.StatusOK, orders)
}

func (h *Handler) handleGetOrder(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	order, err := h.Store.GetOrder(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusNotFound, "Order not found")
		return
	}
	h.respondWithJSON(res, http.StatusOK, order)
}

func (h *Handler) handleCreateOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
	customerExists, _ := h.Store.GetCustomer(ctx, order.Customer.ID)
	if customerExists.ID == 0 {
		h.respondWithError(w, http.StatusBadRequest, "Customer does not exist")
		return
	}

	newOrder, err := h.Store.CreateOrder(ctx, order)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, newOrder)
}

func (h *Handler) handleUpdateOrder(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) {
	var order m.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	err := h.Store.UpdateOrder(ctx, id, order)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order")
		return
	}
	h.respondWithJSON(w, http.StatusOK, "Order updated successfully")
}

func (h *Handler) handleDeleteOrder(ctx context.Context, w http.ResponseWriter, id int, res http.ResponseWriter) {
	err := h.Store.DeleteOrder(ctx, id)
	if err != nil {
		h.respondWithError(res, http.StatusInternalServerError, "Failed to delete order")
		return
	}
	h.respondWithJSON(res, http.StatusOK, "Order deleted successfully")
}

func (h *Handler) handleSearchOrders(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	criteria := m.SearchCriteriaOrders{
		CustomerName: r.URL.Query().Get("customer_name"),
		Status:       r.URL.Query().Get("status"),
	}

	orders, err := h.Store.SearchOrders(ctx, criteria)
	if err == sql.ErrNoRows {
		h.respondWithError(w, http.StatusNotFound, "No orders found")
		return
	} else if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search orders")
		return
	}

	h.respondWithJSON(w, http.StatusOK, orders)
}
