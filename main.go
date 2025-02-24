package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	r.HandleFunc("/books/{id}", handler.HandleBook).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/books", handler.HandleBooks).Methods("GET", "POST")

	r.HandleFunc("/authors/{id}", handler.HandleAuthor).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/authors", handler.HandleAuthors).Methods("GET", "POST")

	r.HandleFunc("/customers/{id}", handler.HandleCustomer).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/customers", handler.HandleCustomers).Methods("GET", "POST")

	r.HandleFunc("/orders/{id}", handler.HandleOrder).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/orders", handler.HandleOrders).Methods("GET", "POST")

	r.HandleFunc("/reports", handler.HandleReports).Methods("GET")

	if _, err := os.Stat("reports"); os.IsNotExist(err) {
		err = os.Mkdir("reports", 0755)
		if err != nil {
			log.Fatalf("Failed to create reports directory: %v", err)
		}
	}

	log.Println("Server running on http://localhost:8080")

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go startDailyReportGenerator(store)

	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Println("Error starting server:", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	db.Close()
}

func startDailyReportGenerator(store *s.PostgresStore) {

	ticker := time.NewTicker(24 * time.Hour)
	//ticker := time.NewTicker(1 * time.Minute)

	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()
		// Generate report for the previous day.
		yesterday := now.AddDate(0, 0, -1)
		start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		end := start.Add(24 * time.Hour)

		// Use a background context so it isn't tied to an HTTP request.
		ctx := context.Background()

		report, err := store.GetSalesReport(ctx, start, end)
		if err != nil {
			log.Printf("Error generating daily report: %v", err)
			continue
		}

		filename := fmt.Sprintf("reports/daily_report_%s.json", start.Format("20060102"))
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Printf("Error marshalling daily report: %v", err)
			continue
		}

		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			log.Printf("Error writing daily report to file: %v", err)
			continue
		}

		log.Printf("Daily report generated and saved: %s", filename)
	}
}
