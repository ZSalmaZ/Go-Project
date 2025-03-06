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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"project.com/myproject/auth"
	h "project.com/myproject/handlers"
	"project.com/myproject/internal/cache"
	s "project.com/myproject/stores"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "2SOUsalma2003"
	dbName     = "mylibrary"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize Redis Cache
	cache.InitCache()

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

	// Initialize JWT Manager and Middleware
	jwtManager := auth.NewJWTManager("your_secret_key", time.Hour)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)

	// Initialize Handlers
	authHandler := h.NewAuthHandler(jwtManager)
	handler := h.NewHandler(store)

	// Create Router
	r := mux.NewRouter()

	// Apply logging and metrics middleware
	r.Use(loggingMiddleware)
	r.Use(metricsMiddleware)

	// Public Routes (No Authentication Needed)
	r.HandleFunc("/login", authHandler.HandleLogin).Methods("POST")
	r.HandleFunc("/register", authHandler.HandleRegister).Methods("POST")

	// Debugging: Print Registered Routes
	log.Println("🔹 Registered Routes:")
	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err == nil {
			log.Println("➡", path)
		}
		return nil
	})

	// Protected Routes (Require Authentication)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Middleware)
	protected.Use(auth.NewRateLimiterMiddleware()) // Apply Rate Limiting

	// Books API
	protected.HandleFunc("/books/{id}", handler.HandleBook).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/books", handler.HandleBooks).Methods("GET", "POST")

	// Authors API
	protected.HandleFunc("/authors/{id}", handler.HandleAuthor).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/authors", handler.HandleAuthors).Methods("GET", "POST")

	// Customers API
	protected.HandleFunc("/customers/{id}", handler.HandleCustomer).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/customers", handler.HandleCustomers).Methods("GET", "POST")

	// Orders API
	protected.HandleFunc("/orders/{id}", handler.HandleOrder).Methods("GET", "DELETE", "PUT")
	protected.HandleFunc("/orders", handler.HandleOrders).Methods("GET", "POST")

	// Reports API
	protected.HandleFunc("/reports", handler.HandleReports).Methods("GET")

	// Metrics Endpoint (For Prometheus)
	r.Handle("/metrics", promhttp.Handler())

	// Ensure "reports" directory exists
	if _, err := os.Stat("reports"); os.IsNotExist(err) {
		if err := os.Mkdir("reports", 0755); err != nil {
			log.Fatalf("Failed to create reports directory: %v", err)
		}
	}

	log.Println("🚀 Server running on http://localhost:8080")

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
			log.Fatalf("❌ Could not start server: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 Shutting down server...")

	// Graceful Shutdown of HTTP Server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("✅ Server gracefully stopped.")
}

// Background Job for Daily Report Generation
func startDailyReportGenerator(store *s.PostgresStore) {
	ticker := time.NewTicker(24 * time.Hour)

	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		end := start.Add(24 * time.Hour)

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

		log.Printf("📊 Daily report generated: %s", filename)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("[%s] %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, duration)
	})
}
