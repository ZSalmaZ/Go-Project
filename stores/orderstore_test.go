package stores

import (
	"context"
	"testing"

	m "project.com/myproject/models"
)

func TestCreateOrder(t *testing.T) {
	order := m.Order{
		Customer:   m.Customer{ID: 1},
		TotalPrice: 59.99,
		Status:     "Pending",
		Items: []m.OrderItem{
			{Book: m.Book{ID: 1}, Quantity: 2},
		},
	}

	createdOrder, err := store.CreateOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}
	if createdOrder.ID == 0 {
		t.Fatalf("Expected valid order ID, got 0")
	}
}

func TestGetOrder(t *testing.T) {
	order, err := store.GetOrder(context.Background(), 1) // Assuming order ID 1 exists
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}
	if order.ID != 1 {
		t.Fatalf("Expected order ID 1, got %d", order.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	_, err := store.GetOrder(context.Background(), 9999)
	if err == nil {
		t.Fatalf("Expected error for non-existent order, got nil")
	}
}
