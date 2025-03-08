package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpUtils "task-management-system/internal/delivery/http/utils"
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

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string    `json:"title" example:"Implement API documentation"`
	Description string    `json:"description" example:"Create comprehensive Swagger documentation for the REST API"`
	Priority    int       `json:"priority" example:"3" minimum:"1" maximum:"5"`
	DueDate     time.Time `json:"due_date" example:"2025-03-15T15:00:00Z"`
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param task body CreateTaskRequest true "Task information"
// @Success 201 {object} httpUtils.ResponseWrapper{data=domain.Task} "Task created successfully"
// @Failure 400 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Invalid input"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
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
			httpUtils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return created task
	httpUtils.RespondWithJSON(w, http.StatusCreated, task)
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "Task ID" example:"60f1a7c9e113d70001abcdef"
// @Success 200 {object} httpUtils.ResponseWrapper{data=domain.Task} "Task retrieved successfully"
// @Failure 404 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Task not found"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks/{id} [get]
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
			httpUtils.RespondWithError(w, http.StatusNotFound, "Task not found")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return task
	httpUtils.RespondWithJSON(w, http.StatusOK, task)
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       string            `json:"title,omitempty" example:"Updated task title"`
	Description string            `json:"description,omitempty" example:"Updated task description"`
	Status      domain.TaskStatus `json:"status,omitempty" example:"in_progress" enums:"pending,in_progress,completed"`
	Priority    int               `json:"priority,omitempty" example:"4" minimum:"1" maximum:"5"`
	DueDate     time.Time         `json:"due_date,omitempty" example:"2025-04-01T15:00:00Z"`
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "Task ID" example:"60f1a7c9e113d70001abcdef"
// @Param task body UpdateTaskRequest true "Updated task information"
// @Success 200 {object} httpUtils.ResponseWrapper{data=domain.Task} "Task updated successfully"
// @Failure 400 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Invalid input"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 403 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Forbidden"
// @Failure 404 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Task not found"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
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
			httpUtils.RespondWithError(w, http.StatusNotFound, "Task not found")
		case domain.ErrUnauthorized:
			httpUtils.RespondWithError(w, http.StatusForbidden, "You are not authorized to update this task")
		case domain.ErrInvalidInput:
			httpUtils.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return updated task
	httpUtils.RespondWithJSON(w, http.StatusOK, task)
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "Task ID" example:"60f1a7c9e113d70001abcdef"
// @Success 204 "No Content"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 403 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Forbidden"
// @Failure 404 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Task not found"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete task
	err := h.taskUseCase.DeleteTask(taskID, userID)
	if err != nil {
		// Handle different error types
		switch err {
		case domain.ErrNotFound:
			httpUtils.RespondWithError(w, http.StatusNotFound, "Task not found")
		case domain.ErrUnauthorized:
			httpUtils.RespondWithError(w, http.StatusForbidden, "You are not authorized to delete this task")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return success - no content
	w.WriteHeader(http.StatusNoContent)
}

// AssignTaskRequest represents the request body for assigning a task
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" example:"60f1a7c9e113d7000fedcba9"`
}

// AssignTask godoc
// @Summary Assign a task to a user
// @Description Assign a task to another user
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "Task ID" example:"60f1a7c9e113d70001abcdef"
// @Param assignment body AssignTaskRequest true "Assignment information"
// @Success 200 {object} httpUtils.ResponseWrapper{data=domain.Task} "Task assigned successfully"
// @Failure 400 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Invalid input"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 403 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Forbidden"
// @Failure 404 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Task or user not found"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks/{id}/assign [post]
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL
	vars := mux.Vars(r)
	taskID := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
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
			httpUtils.RespondWithError(w, http.StatusNotFound, "Task or user not found")
		case domain.ErrUnauthorized:
			httpUtils.RespondWithError(w, http.StatusForbidden, "You are not authorized to assign this task")
		default:
			httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return updated task
	httpUtils.RespondWithJSON(w, http.StatusOK, task)
}

// ListTasks godoc
// @Summary List tasks
// @Description Get a list of tasks with optional status filter
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param status query string false "Filter tasks by status" Enums(pending, in_progress, completed)
// @Success 200 {object} httpUtils.ResponseWrapper{data=[]domain.Task} "Tasks retrieved successfully"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /tasks [get]
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
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Return tasks
	httpUtils.RespondWithJSON(w, http.StatusOK, tasks)
}

// GetUserTasks godoc
// @Summary Get user's tasks
// @Description Get tasks created by or assigned to a user
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path string true "User ID" example:"60f1a7c9e113d70001234567"
// @Success 200 {object} httpUtils.ResponseWrapper{data=[]domain.Task} "Tasks retrieved successfully"
// @Failure 401 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Unauthorized"
// @Failure 404 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "User not found"
// @Failure 500 {object} httpUtils.ResponseWrapper{error=httpUtils.RespondErrorInfo} "Internal server error"
// @Router /users/{id}/tasks [get]
func (h *TaskHandler) GetUserTasks(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get tasks
	tasks, err := h.taskUseCase.GetUserTasks(userID)
	if err != nil {
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Return tasks
	httpUtils.RespondWithJSON(w, http.StatusOK, tasks)
}
