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

type userRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database, timeout time.Duration) domain.UserRepository {
	collection := db.Collection("users")

	// Create indexes
	indexModel := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
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

	return &userRepository{
		collection: collection,
		timeout:    timeout,
	}
}

// FindByID finds a user by its ID
func (r *userRepository) FindByID(id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByUsername finds a user by username
func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Create creates a new user
func (r *userRepository) Create(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Check if user with the same email or username already exists
	existingUser, err := r.FindByEmail(user.Email)
	if err == nil && existingUser != nil {
		return domain.ErrDuplicateKey
	}

	existingUser, err = r.FindByUsername(user.Username)
	if err == nil && existingUser != nil {
		return domain.ErrDuplicateKey
	}

	// Set created and updated times
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// If ID is not set, set it to a new ObjectID
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	_, err = r.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrDuplicateKey
	}
	return err
}

// Update updates an existing user
func (r *userRepository) Update(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Update the updated time
	user.UpdatedAt = time.Now()

	// Create an update document
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"updated_at": user.UpdatedAt,
		},
	}

	// Only update password if it's not empty
	if user.Password != "" {
		update["$set"].(bson.M)["password"] = user.Password
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		update,
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrDuplicateKey
		}
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a user by its ID
func (r *userRepository) Delete(id primitive.ObjectID) error {
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
