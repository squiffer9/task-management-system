package handlers

import (
	"encoding/json"
	"net/http"

	"task-management-system/internal/domain"
	"task-management-system/internal/usecase"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	userUseCase *usecase.UserUseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUseCase *usecase.AuthUseCase, userUseCase *usecase.UserUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		userUseCase: userUseCase,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// RegisterResponse represents the response for user registration
type RegisterResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Register handles POST /auth/register requests
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Register user
	user, err := h.userUseCase.RegisterUser(&usecase.RegisterUserInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		// Handle error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create response
	resp := RegisterResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// Return created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Login    string `json:"login"` // username or email
	Password string `json:"password"`
}

// LoginResponse represents the response for user login
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
}

// Login handles POST /auth/login requests
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	result, err := h.authUseCase.Login(&usecase.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})

	if err != nil {
		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	// Create response
	resp := LoginResponse{
		AccessToken: result.AccessToken,
		ExpiresAt:   result.ExpiresAt.Format(http.TimeFormat),
		UserID:      result.UserID,
		Username:    result.Username,
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RefreshTokenRequest represents the request body for refreshing token
type RefreshTokenRequest struct {
	Token string `json:"token"`
}

// RefreshToken handles POST /auth/refresh-token requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Refresh token
	result, err := h.authUseCase.RefreshToken(req.Token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Create response
	resp := LoginResponse{
		AccessToken: result.AccessToken,
		ExpiresAt:   result.ExpiresAt.Format(http.TimeFormat),
		UserID:      result.UserID,
		Username:    result.Username,
	}

	// Return new token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
