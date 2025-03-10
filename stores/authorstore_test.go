package stores

import (
	"context"
	"testing"

	m "project.com/myproject/models"
)

func TestCreateAuthor(t *testing.T) {
	if store == nil {
		t.Fatal("❌ store is nil; TestMain did not initialize it")
	}

	author := m.Author{
		FirstName: "John",
		LastName:  "Doe",
		Bio:       "A famous author.",
	}

	createdAuthor, err := store.CreateAuthor(context.Background(), author)
	if err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}
	if createdAuthor.ID == 0 {
		t.Fatalf("Expected valid author ID, got 0")
	}
}

func TestGetAuthor(t *testing.T) {
	if store == nil {
		t.Fatal("❌ store is nil; TestMain did not initialize it")
	}

	author, err := store.GetAuthor(context.Background(), 1) // Assume author with ID 1 exists
	if err != nil {
		t.Fatalf("Failed to get author: %v", err)
	}
	if author.ID != 1 {
		t.Fatalf("Expected author ID 1, got %d", author.ID)
	}
}

func TestGetAuthor_NotFound(t *testing.T) {
	if store == nil {
		t.Fatal("❌ store is nil; TestMain did not initialize it")
	}

	_, err := store.GetAuthor(context.Background(), 9999) // Non-existent ID
	if err == nil {
		t.Fatalf("Expected error for non-existent author, got nil")
	}
}

func TestGetAllAuthors(t *testing.T) {
	if store == nil {
		t.Fatal("❌ store is nil; TestMain did not initialize it")
	}

	authors, err := store.GetAllAuthors(context.Background())
	if err != nil {
		t.Fatalf("Failed to get all authors: %v", err)
	}
	if len(authors) == 0 {
		t.Log("No authors found, consider adding some for the test")
	}
}
