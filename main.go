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
	"project.com/myproject/auth"

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
	// Initialize JWT Manager and Middleware
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)

	// Connect to Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Store
	store := s.NewPostgresStore(db)

	// Initialize Handlers
	authHandler := h.NewAuthHandler(jwtManager)
	handler := h.NewHandler(store)

	// Create Router
	r := mux.NewRouter()

	// Register authentication routes
	authHandler.RegisterRoutes(r)

	// Protected Routes (Require Authentication)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Middleware)

	protected.HandleFunc("/books/{id}", handler.HandleBook).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/books", handler.HandleBooks).Methods("GET", "POST")

	protected.HandleFunc("/authors/{id}", handler.HandleAuthor).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/authors", handler.HandleAuthors).Methods("GET", "POST")

	protected.HandleFunc("/customers/{id}", handler.HandleCustomer).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/customers", handler.HandleCustomers).Methods("GET", "POST")

	protected.HandleFunc("/orders/{id}", handler.HandleOrder).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/orders", handler.HandleOrders).Methods("GET", "POST")

	protected.HandleFunc("/reports", handler.HandleReports).Methods("GET")

	// Ensure "reports" directory exists
	if _, err := os.Stat("reports"); os.IsNotExist(err) {
		if err := os.Mkdir("reports", 0755); err != nil {
			log.Fatalf("Failed to create reports directory: %v", err)
		}
	}

	log.Println("Server running on http://localhost:8080")

	// Graceful Shutdown Handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start Daily Report Generator
	go startDailyReportGenerator(store)

	// Start HTTP Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	// Graceful Shutdown of HTTP Server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server gracefully stopped.")
}

// Background Job for Daily Report Generation
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
