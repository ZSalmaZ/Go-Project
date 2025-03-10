package stores

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	m "project.com/myproject/models"
)

func TestCreateBook(t *testing.T) {
	if store == nil {
		t.Fatal("❌ store is nil; TestMain did not initialize it")
	}

	// Step 1: Create author first
	author := m.Author{
		FirstName: "John",
		LastName:  "Doe",
		Bio:       "Test Bio",
	}
	createdAuthor, err := store.CreateAuthor(context.Background(), author)
	if err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}

	// Step 2: Now create book linked to author
	book := m.Book{
		Title:       "Test Book",
		Author:      createdAuthor,
		PublishedAt: time.Now(),
		Price:       19.99,
		Stock:       10,
		Genres:      []string{"Fiction"},
	}
	createdBook, err := store.CreateBook(context.Background(), book)
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}
	if createdBook.ID == 0 {
		t.Fatalf("Expected valid book ID, got 0")
	}
}

func TestGetBook(t *testing.T) {
	// Assuming book ID 1 exists for test
	book, err := store.GetBook(context.Background(), 1)
	if err != nil {
		t.Fatalf("Failed to get book: %v", err)
	}
	if book.ID != 1 {
		t.Fatalf("Expected book ID 1, got %d", book.ID)
	}
}

func TestGetBook_NotFound(t *testing.T) {
	_, err := store.GetBook(context.Background(), 9999) // Non-existent ID
	if err == nil {
		t.Fatalf("Expected error for non-existent book, got nil")
	}
}
