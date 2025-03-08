package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	httpUtils "task-management-system/internal/delivery/http/utils"
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

// UserResponse represents the response for user data
type UserResponse struct {
	ID        string `json:"id" example:"60f1a7c9e113d70001234567"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john.doe@example.com"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
	CreatedAt string `json:"created_at" example:"Sat, 01 Mar 2025 12:00:00 GMT"`
	UpdatedAt string `json:"updated_at" example:"Sat, 08 Mar 2025 15:00:00 GMT"`
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "User ID" example:"60f1a7c9e113d70001234567"
// @Success 200 {object} ResponseWrapper{data=UserResponse} "User retrieved successfully"
// @Failure 401 {object} ResponseWrapper{error=ErrorInfo} "Unauthorized"
// @Failure 404 {object} ResponseWrapper{error=ErrorInfo} "User not found"
// @Failure 500 {object} ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /users/{id} [get]
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
			httpUtils.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Create a response struct to avoid sending password
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
	httpUtils.RespondWithJSON(w, http.StatusOK, resp)
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email     string `json:"email,omitempty" example:"new.email@example.com" format:"email"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
	Password  string `json:"password,omitempty" example:"newsecurepassword123" minLength:"6"`
}

// UpdateUser godoc
// @Summary Update user
// @Description Update a user's profile
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "User ID" example:"60f1a7c9e113d70001234567"
// @Param user body UpdateUserRequest true "Updated user information"
// @Success 200 {object} ResponseWrapper{data=UserResponse} "User updated successfully"
// @Failure 400 {object} ResponseWrapper{error=ErrorInfo} "Invalid input"
// @Failure 401 {object} ResponseWrapper{error=ErrorInfo} "Unauthorized"
// @Failure 403 {object} ResponseWrapper{error=ErrorInfo} "Forbidden - cannot update another user's profile"
// @Failure 404 {object} ResponseWrapper{error=ErrorInfo} "User not found"
// @Failure 409 {object} ResponseWrapper{error=ErrorInfo} "Email already in use"
// @Failure 500 {object} ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get authenticated user ID from context
	authenticatedUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if the authenticated user is updating their own profile
	if authenticatedUserID != userID {
		httpUtils.RespondWithError(w, http.StatusForbidden, "You can only update your own profile")
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
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
			httpUtils.RespondWithError(w, http.StatusNotFound, "User not found")
		case domain.ErrInvalidInput:
			httpUtils.RespondWithError(w, http.StatusBadRequest, err.Error())
		case domain.ErrDuplicateKey:
			httpUtils.RespondWithError(w, http.StatusConflict, "Email already in use")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Create a response struct to avoid sending password
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
	httpUtils.RespondWithJSON(w, http.StatusOK, resp)
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} ResponseWrapper{data=UserResponse} "User profile retrieved successfully"
// @Failure 401 {object} ResponseWrapper{error=ErrorInfo} "Unauthorized"
// @Failure 404 {object} ResponseWrapper{error=ErrorInfo} "User not found"
// @Failure 500 {object} ResponseWrapper{error=ErrorInfo} "Internal server error"
// @Router /me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user
	user, err := h.userUseCase.GetUserByID(userID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			httpUtils.RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Create a response struct to avoid sending password
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
	httpUtils.RespondWithJSON(w, http.StatusOK, resp)
}
