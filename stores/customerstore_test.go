package stores

import (
	"context"
	"testing"

	m "project.com/myproject/models"
)

func TestCreateCustomer(t *testing.T) {
	customer := m.Customer{
		Name:       "Alice Johnson",
		Email:      "alice@example.com",
		Street:     "123 Main St",
		City:       "Metropolis",
		State:      "CA",
		PostalCode: "12345",
		Country:    "USA",
	}

	createdCustomer, err := store.CreateCustomer(context.Background(), customer)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}
	if createdCustomer.ID == 0 {
		t.Fatalf("Expected valid customer ID, got 0")
	}
}

func TestGetCustomer(t *testing.T) {
	customer, err := store.GetCustomer(context.Background(), 1) // Assuming ID 1 exists
	if err != nil {
		t.Fatalf("Failed to get customer: %v", err)
	}
	if customer.ID != 1 {
		t.Fatalf("Expected customer ID 1, got %d", customer.ID)
	}
}

func TestGetCustomer_NotFound(t *testing.T) {
	_, err := store.GetCustomer(context.Background(), 9999) // Non-existent ID
	if err == nil {
		t.Fatalf("Expected error for non-existent customer, got nil")
	}
}
