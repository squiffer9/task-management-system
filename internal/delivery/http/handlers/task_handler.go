package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"task-management-system/internal/domain"
	"task-management-system/internal/usecase"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	taskUseCase *usecase.TaskUseCase
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskUseCase *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{
		taskUseCase: taskUseCase,
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param task body CreateTaskRequest true "Task information"
// @Success 201 {object} domain.Task
// @Failure 400 {object} Error
// @Failure 401 {object} Error
// @Router /tasks [post]
// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	DueDate     time.Time `json:"due_date"`
}

// CreateTask handles POST /tasks requests
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create task
	task, err := h.taskUseCase.CreateTask(&usecase.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		CreatedBy:   userID,
	})

	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return created task
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTask handles GET /tasks/{id} requests
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get task
	task, err := h.taskUseCase.GetTaskByID(taskID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "Task not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return task
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Status      domain.TaskStatus `json:"status,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	DueDate     time.Time         `json:"due_date,omitempty"`
}

// UpdateTask handles PUT /tasks/{id} requests
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update task
	task, err := h.taskUseCase.UpdateTask(&usecase.UpdateTaskInput{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		UpdatedBy:   userID,
	})

	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "Task not found", http.StatusNotFound)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusForbidden)
		case domain.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return updated task
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE /tasks/{id} requests
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete task
	err := h.taskUseCase.DeleteTask(taskID, userID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "Task not found", http.StatusNotFound)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// AssignTaskRequest represents the request body for assigning a task
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id"`
}

// AssignTask handles POST /tasks/{id}/assign requests
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Assign task
	task, err := h.taskUseCase.AssignTask(&usecase.AssignTaskInput{
		TaskID:     taskID,
		AssigneeID: req.AssigneeID,
		AssignedBy: userID,
	})

	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			http.Error(w, "Task or user not found", http.StatusNotFound)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return updated task
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// ListTasks handles GET /tasks requests
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Get status from query parameter
	status := r.URL.Query().Get("status")

	var input *usecase.ListTasksInput
	if status != "" {
		input = &usecase.ListTasksInput{
			Status: domain.TaskStatus(status),
		}
	}

	// Get tasks
	tasks, err := h.taskUseCase.ListTasks(input)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return tasks
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetUserTasks handles GET /users/{id}/tasks requests
func (h *TaskHandler) GetUserTasks(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get tasks
	tasks, err := h.taskUseCase.GetUserTasks(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return tasks
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
