package stores

import (
	"context"
	"testing"
	"time"

	m "project.com/myproject/models"
)

func TestAuthorIntegration(t *testing.T) {
	ctx := context.Background()

	author := m.Author{
		FirstName: "UniqueAuthorFirst",
		LastName:  "UniqueAuthorLast",
		Bio:       "A test bio",
	}

	createdAuthor, err := store.CreateAuthor(ctx, author)
	if err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}

	fetchedAuthor, err := store.GetAuthor(ctx, createdAuthor.ID)
	if err != nil {
		t.Fatalf("Failed to get author: %v", err)
	}

	if fetchedAuthor.FirstName != author.FirstName {
		t.Fatalf("Expected first name %s, got %s", author.FirstName, fetchedAuthor.FirstName)
	}
}

func TestBookIntegration(t *testing.T) {
	ctx := context.Background()

	author := m.Author{
		FirstName: "BookAuthorFirst",
		LastName:  "BookAuthorLast",
		Bio:       "Book Author Bio",
	}
	createdAuthor, _ := store.CreateAuthor(ctx, author)

	book := m.Book{
		Title:       "Test Book Integration",
		Author:      createdAuthor,
		PublishedAt: time.Now(),
		Price:       10.99,
		Stock:       7,
		Genres:      []string{"Science"},
	}

	createdBook, err := store.CreateBook(ctx, book)
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	fetchedBook, err := store.GetBook(ctx, createdBook.ID)
	if err != nil {
		t.Fatalf("Failed to fetch book: %v", err)
	}

	if fetchedBook.Title != book.Title {
		t.Fatalf("Expected title %s, got %s", book.Title, fetchedBook.Title)
	}
}

func TestCustomerIntegration(t *testing.T) {
	ctx := context.Background()

	customer := m.Customer{
		Name:       "Test Customer",
		Email:      "unique_customer@example.com", // unique email to avoid conflicts
		Street:     "123 Test St",
		City:       "Test City",
		State:      "TS",
		PostalCode: "12345",
		Country:    "TestCountry",
	}

	createdCustomer, err := store.CreateCustomer(ctx, customer)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	fetchedCustomer, err := store.GetCustomer(ctx, createdCustomer.ID)
	if err != nil {
		t.Fatalf("Failed to get customer: %v", err)
	}

	if fetchedCustomer.Email != customer.Email {
		t.Fatalf("Expected email %s, got %s", customer.Email, fetchedCustomer.Email)
	}
}

func TestOrderIntegration(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create Customer
	customer := m.Customer{
		Name:       "Order Customer",
		Email:      "unique_order_customer@example.com",
		Street:     "456 Order St",
		City:       "Order City",
		State:      "OC",
		PostalCode: "45678",
		Country:    "OrderCountry",
	}
	createdCustomer, _ := store.CreateCustomer(ctx, customer)

	// Step 2: Create Author and Book
	author := m.Author{
		FirstName: "Order Author",
		LastName:  "Order Last",
		Bio:       "Order Bio",
	}
	createdAuthor, _ := store.CreateAuthor(ctx, author)

	book := m.Book{
		Title:       "Order Test Book",
		Author:      createdAuthor,
		PublishedAt: time.Now(),
		Price:       12.99,
		Stock:       10,
		Genres:      []string{"History"},
	}
	createdBook, _ := store.CreateBook(ctx, book)

	// Step 3: Create Order
	order := m.Order{
		Customer:   createdCustomer,
		TotalPrice: 12.99,
		Status:     "Pending",
		Items: []m.OrderItem{
			{Book: createdBook, Quantity: 1},
		},
	}

	createdOrder, err := store.CreateOrder(ctx, order)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Step 4: Fetch Order
	fetchedOrder, err := store.GetOrder(ctx, createdOrder.ID)
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}

	if fetchedOrder.Customer.Email != customer.Email {
		t.Fatalf("Expected customer email %s, got %s", customer.Email, fetchedOrder.Customer.Email)
	}
}
