package usecase

import (
	"errors"
	"time"

	"task-management-system/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskUseCase handles business logic related to tasks
type TaskUseCase struct {
	taskRepo domain.TaskRepository
	userRepo domain.UserRepository
}

// NewTaskUseCase creates a new task use case
func NewTaskUseCase(taskRepo domain.TaskRepository, userRepo domain.UserRepository) *TaskUseCase {
	return &TaskUseCase{
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}

// CreateTaskInput represents input data for task creation
type CreateTaskInput struct {
	Title       string
	Description string
	Priority    int
	DueDate     time.Time
	CreatedBy   string // User ID as string
}

// CreateTask creates a new task
func (uc *TaskUseCase) CreateTask(input *CreateTaskInput) (*domain.Task, error) {
	// Validate input
	if input.Title == "" {
		return nil, domain.ErrInvalidInput
	}

	// Validate priority (1-5)
	if input.Priority < 1 || input.Priority > 5 {
		return nil, errors.New("priority must be between 1 and 5")
	}

	// Convert creator ID from string to ObjectID
	creatorID, err := primitive.ObjectIDFromHex(input.CreatedBy)
	if err != nil {
		return nil, errors.New("invalid creator ID format")
	}

	// Verify that creator exists
	_, err = uc.userRepo.FindByID(creatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("creator user not found")
		}
		return nil, err
	}

	// Create the task
	task := &domain.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TaskStatusPending,
		Priority:    input.Priority,
		DueDate:     input.DueDate,
		CreatedBy:   creatorID,
	}

	// Save to repository
	err = uc.taskRepo.Create(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTaskByID retrieves a task by its ID
func (uc *TaskUseCase) GetTaskByID(id string) (*domain.Task, error) {
	// Convert ID from string to ObjectID
	taskID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}

	// Retrieve the task
	task, err := uc.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// UpdateTaskInput represents input data for task update
type UpdateTaskInput struct {
	ID          string
	Title       string
	Description string
	Status      domain.TaskStatus
	Priority    int
	DueDate     time.Time
	UpdatedBy   string // User ID as string
}

// UpdateTask updates an existing task
func (uc *TaskUseCase) UpdateTask(input *UpdateTaskInput) (*domain.Task, error) {
	// Convert ID from string to ObjectID
	taskID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}

	// Retrieve the existing task
	task, err := uc.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// Validate priority if provided
	if input.Priority != 0 && (input.Priority < 1 || input.Priority > 5) {
		return nil, errors.New("priority must be between 1 and 5")
	}

	// Convert updater ID from string to ObjectID
	updaterID, err := primitive.ObjectIDFromHex(input.UpdatedBy)
	if err != nil {
		return nil, errors.New("invalid updater ID format")
	}

	// Verify that updater exists and is authorized
	// (either the creator or assigned to the task)
	if !task.CreatedBy.Equal(updaterID) && !task.AssignedTo.Equal(updaterID) {
		return nil, domain.ErrUnauthorized
	}

	// Update task fields if provided
	if input.Title != "" {
		task.Title = input.Title
	}

	if input.Description != "" {
		task.Description = input.Description
	}

	if input.Status != "" {
		// Validate status transition
		if !isValidStatusTransition(task.Status, input.Status) {
			return nil, errors.New("invalid status transition")
		}
		task.Status = input.Status
	}

	if input.Priority != 0 {
		task.Priority = input.Priority
	}

	// Only update due date if a non-zero time is provided
	if !input.DueDate.IsZero() {
		task.DueDate = input.DueDate
	}

	// Save to repository
	err = uc.taskRepo.Update(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask deletes a task by ID
func (uc *TaskUseCase) DeleteTask(id string, userID string) error {
	// Convert IDs from string to ObjectID
	taskID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid task ID format")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Retrieve the task to check authorization
	task, err := uc.taskRepo.FindByID(taskID)
	if err != nil {
		return err
	}

	// Only the creator can delete a task
	if !task.CreatedBy.Equal(userObjID) {
		return domain.ErrUnauthorized
	}

	// Delete from repository
	return uc.taskRepo.Delete(taskID)
}

// AssignTaskInput represents input data for task assignment
type AssignTaskInput struct {
	TaskID     string
	AssigneeID string
	AssignedBy string
}

// AssignTask assigns a task to a user
func (uc *TaskUseCase) AssignTask(input *AssignTaskInput) (*domain.Task, error) {
	// Convert IDs from string to ObjectID
	taskID, err := primitive.ObjectIDFromHex(input.TaskID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}

	assigneeID, err := primitive.ObjectIDFromHex(input.AssigneeID)
	if err != nil {
		return nil, errors.New("invalid assignee ID format")
	}

	assignerID, err := primitive.ObjectIDFromHex(input.AssignedBy)
	if err != nil {
		return nil, errors.New("invalid assigner ID format")
	}

	// Retrieve the task
	task, err := uc.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// Only the creator can assign a task
	if !task.CreatedBy.Equal(assignerID) {
		return nil, domain.ErrUnauthorized
	}

	// Verify that assignee exists
	_, err = uc.userRepo.FindByID(assigneeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("assignee user not found")
		}
		return nil, err
	}

	// Assign the task
	task.AssignedTo = assigneeID

	// If task is pending, move it to in progress
	if task.Status == domain.TaskStatusPending {
		task.Status = domain.TaskStatusInProgress
	}

	// Save to repository
	err = uc.taskRepo.Update(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetUserTasks retrieves all tasks for a specific user (created by or assigned to)
func (uc *TaskUseCase) GetUserTasks(userID string) ([]*domain.Task, error) {
	// Convert ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Retrieve the tasks
	tasks, err := uc.taskRepo.FindByUser(userObjID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// ListTasksInput represents filtering options for task listing
type ListTasksInput struct {
	Status domain.TaskStatus
}

// ListTasks lists tasks with optional filtering
func (uc *TaskUseCase) ListTasks(input *ListTasksInput) ([]*domain.Task, error) {
	// If status filter is provided, use it
	if input != nil && input.Status != "" {
		return uc.taskRepo.FindByStatus(input.Status)
	}

	// Otherwise return all tasks
	return uc.taskRepo.FindAll(nil)
}

// Helper function to validate status transitions
func isValidStatusTransition(current domain.TaskStatus, new domain.TaskStatus) bool {
	// Define valid transitions
	switch current {
	case domain.TaskStatusPending:
		// Pending can move to in progress or completed
		return new == domain.TaskStatusInProgress || new == domain.TaskStatusCompleted
	case domain.TaskStatusInProgress:
		// In progress can move to completed only
		return new == domain.TaskStatusCompleted
	case domain.TaskStatusCompleted:
		// Completed can move back to in progress (if revisions needed)
		return new == domain.TaskStatusInProgress
	default:
		return false
	}
}
