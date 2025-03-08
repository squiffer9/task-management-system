package handlers

import (
	"encoding/json"
	"net/http"

	httpUtils "task-management-system/internal/delivery/http/utils"
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
	Username  string `json:"username" example:"johndoe" minLength:"3"`
	Email     string `json:"email" example:"john.doe@example.com" format:"email"`
	Password  string `json:"password" example:"securepassword123" minLength:"6"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
}

// RegisterResponse represents the response for user registration
type RegisterResponse struct {
	ID        string `json:"id" example:"60f1a7c9e113d70001234567"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john.doe@example.com"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param registration body RegisterRequest true "User registration information"
// @Success 201 {object} httpUtils.ResponseWrapper{data=RegisterResponse} "User registered successfully"
// @Failure 400 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Invalid input or duplicate username/email"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
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
		httpUtils.RespondWithError(w, http.StatusBadRequest, err.Error())
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
	httpUtils.RespondWithJSON(w, http.StatusCreated, resp)
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Login    string `json:"login" example:"johndoe" description:"Username or email"`
	Password string `json:"password" example:"securepassword123"`
}

// LoginResponse represents the response for user login
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt   string `json:"expires_at" example:"Sat, 08 Mar 2025 15:00:00 GMT"`
	UserID      string `json:"user_id" example:"60f1a7c9e113d70001234567"`
	Username    string `json:"username" example:"johndoe"`
}

// Login godoc
// @Summary Authenticate user
// @Description Authenticate a user and get a JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} httpUtils.ResponseWrapper{data=LoginResponse} "User authenticated successfully"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Invalid credentials"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Authenticate user
	result, err := h.authUseCase.Login(&usecase.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})

	if err != nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid login credentials")
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
	httpUtils.RespondWithJSON(w, http.StatusOK, resp)
}

// RefreshTokenRequest represents the request body for refreshing token
type RefreshTokenRequest struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Get a new JWT token using a valid token
// @Tags authentication
// @Accept json
// @Produce json
// @Param token body RefreshTokenRequest true "Current valid token"
// @Success 200 {object} httpUtils.ResponseWrapper{data=LoginResponse} "Token refreshed successfully"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Invalid or expired token"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Refresh token
	result, err := h.authUseCase.RefreshToken(req.Token)
	if err != nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
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
	httpUtils.RespondWithJSON(w, http.StatusOK, resp)
}
