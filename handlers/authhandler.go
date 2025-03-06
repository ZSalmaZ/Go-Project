package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"project.com/myproject/auth"

	"github.com/gorilla/mux"
)

// In-memory user store (Replace this with a database in production)
var userStore = make(map[string]string)
var mu sync.Mutex

type AuthHandler struct {
	JWTManager *auth.JWTManager
}

func NewAuthHandler(jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		JWTManager: jwtManager,
	}
}

// ✅ Register authentication routes
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.HandleLogin).Methods("POST")
	router.HandleFunc("/register", h.HandleRegister).Methods("POST") // ✅ Added register route
}

// ✅ Struct for Login & Register Requests
type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ✅ Struct for JSON Response
type authResponse struct {
	Token  string `json:"token,omitempty"`
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
}

// ✅ Handle User Registration
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	// Decode JSON Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Lock the user store (for thread safety)
	mu.Lock()
	defer mu.Unlock()

	// Check if user already exists
	if _, exists := userStore[req.Username]; exists {
		http.Error(w, `{"error": "User already exists"}`, http.StatusConflict)
		return
	}

	// Store the new user (⚠️ In real applications, **hash** the password before storing)
	userStore[req.Username] = req.Password

	// ✅ Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		Status: "User registered successfully",
	})
}

// ✅ Handle User Login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	// Decode JSON Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Check if the user exists
	mu.Lock()
	storedPassword, exists := userStore[req.Username]
	mu.Unlock()

	if !exists || storedPassword != req.Password {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.JWTManager.Generate(req.Username)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Return JSON Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		Token: token,
	})
}
