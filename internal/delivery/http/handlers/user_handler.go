package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"task-management-system/internal/domain"
	"task-management-system/internal/usecase"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// GetUser handles GET /users/{id} requests
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get user
	user, err := h.userUseCase.GetUserByID(userID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create a response struct to avoid sending password
	type UserResponse struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	resp := UserResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
	}

	// Return user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Password  string `json:"password,omitempty"`
}

// UpdateUser handles PUT /users/{id} requests
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get authenticated user ID from context
	authenticatedUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the authenticated user is updating their own profile
	if authenticatedUserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update user
	user, err := h.userUseCase.UpdateUser(&usecase.UpdateUserInput{
		ID:        userID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	})

	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case domain.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrDuplicateKey:
			http.Error(w, "Email already in use", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create a response struct to avoid sending password
	type UserResponse struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	resp := UserResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
	}

	// Return updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetProfile handles GET /me requests
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.userUseCase.GetUserByID(userID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create a response struct to avoid sending password
	type UserResponse struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	resp := UserResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
	}

	// Return user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
