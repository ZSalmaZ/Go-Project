package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	h "project.com/myproject/handlers"
	s "project.com/myproject/stores"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "2SOUsalma2003"
	dbName     = "mylibrary"
)

func main() {
	// PostgreSQL connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the store
	store := s.NewPostgresStore(db)

	// Initialize the handler with the store
	handler := h.NewHandler(store)

	// Setup router
	r := mux.NewRouter()

	// Define API routes using the unified handler
	r.HandleFunc("/books/{id}", handler.HandleBook).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/books", handler.HandleBooks).Methods("GET", "POST")

	r.HandleFunc("/authors/{id}", handler.HandleAuthor).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/authors", handler.HandleAuthors).Methods("GET", "POST")

	r.HandleFunc("/customers/{id}", handler.HandleCustomer).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/customers", handler.HandleCustomers).Methods("GET", "POST")

	r.HandleFunc("/orders/{id}", handler.HandleOrder).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/orders", handler.HandleOrders).Methods("GET", "POST")

	//r.HandleFunc("/reports", handler.HandleReports).Methods("GET")

	log.Println("Server running on http://localhost:8080")

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Println("Error starting server:", err)
		}
	}()

	// Wait for termination signal
	<-stop
	log.Println("Shutting down server...")

	// Close the database connection on shutdown
	db.Close()
}
