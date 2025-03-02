package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

// Task represents a task entity
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title" validate:"required"`
	Description string             `bson:"description" json:"description"`
	Status      TaskStatus         `bson:"status" json:"status"`
	Priority    int                `bson:"priority" json:"priority" validate:"min=1,max=5"`
	DueDate     time.Time          `bson:"due_date" json:"due_date"`
	AssignedTo  primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to,omitempty"`
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	FindByID(id primitive.ObjectID) (*Task, error)
	FindAll(filter map[string]interface{}) ([]*Task, error)
	Create(task *Task) error
	Update(task *Task) error
	Delete(id primitive.ObjectID) error
	FindByUser(userID primitive.ObjectID) ([]*Task, error)
	FindByStatus(status TaskStatus) ([]*Task, error)
}
