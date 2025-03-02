package mongodb

import (
	"context"
	"errors"
	"time"

	"task-management-system/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type taskRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *mongo.Database, timeout time.Duration) domain.TaskRepository {
	collection := db.Collection("tasks")

	// Create indexes
	indexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "created_by", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "assigned_to", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "due_date", Value: 1}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		// Log error but continue - indexes are for performance, not functionality
		// In production, you might want to handle this differently
		// log.Printf("Error creating indexes: %v", err)
	}

	return &taskRepository{
		collection: collection,
		timeout:    timeout,
	}
}

// FindByID finds a task by its ID
func (r *taskRepository) FindByID(id primitive.ObjectID) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var task domain.Task
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &task, nil
}

// FindAll finds all tasks matching the filter
func (r *taskRepository) FindAll(filter map[string]interface{}) ([]*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	filterBson := bson.M{}
	if filter != nil {
		filterBson = bson.M(filter)
	}

	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}})
	cursor, err := r.collection.Find(ctx, filterBson, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Create creates a new task
func (r *taskRepository) Create(task *domain.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Set created and updated times
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// If ID is not set, set it to a new ObjectID
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}

	// Default status to pending if not set
	if task.Status == "" {
		task.Status = domain.TaskStatusPending
	}

	_, err := r.collection.InsertOne(ctx, task)
	return err
}

// Update updates an existing task
func (r *taskRepository) Update(task *domain.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Update the updated time
	task.UpdatedAt = time.Now()

	// Create an update document
	update := bson.M{
		"$set": bson.M{
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"priority":    task.Priority,
			"due_date":    task.DueDate,
			"assigned_to": task.AssignedTo,
			"updated_at":  task.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": task.ID},
		update,
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a task by its ID
func (r *taskRepository) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// FindByUser finds tasks by user ID (either created by or assigned to)
func (r *taskRepository) FindByUser(userID primitive.ObjectID) ([]*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"created_by": userID},
			{"assigned_to": userID},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// FindByStatus finds tasks by status
func (r *taskRepository) FindByStatus(status domain.TaskStatus) ([]*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	filter := bson.M{"status": status}

	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*domain.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}
