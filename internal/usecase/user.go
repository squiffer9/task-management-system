package usecase

import (
	"errors"
	"regexp"
	"time"

	"task-management-system/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserUseCase handles business logic related to users
type UserUseCase struct {
	userRepo domain.UserRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo domain.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

// RegisterUserInput represents input data for user registration
type RegisterUserInput struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// RegisterUser registers a new user
func (uc *UserUseCase) RegisterUser(input *RegisterUserInput) (*domain.User, error) {
	// Validate input
	if err := validateUserInput(input); err != nil {
		return nil, err
	}

	// Check if user with the same email already exists
	existingUser, err := uc.userRepo.FindByEmail(input.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Check if user with the same username already exists
	existingUser, err = uc.userRepo.FindByUsername(input.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already taken")
	}

	// Hash the password
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create the user
	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}

	// Save to repository
	err = uc.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (uc *UserUseCase) GetUserByID(id string) (*domain.User, error) {
	// Convert ID from string to ObjectID
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Retrieve the user
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uc *UserUseCase) GetUserByEmail(email string) (*domain.User, error) {
	// Validate email
	if !isValidEmail(email) {
		return nil, errors.New("invalid email format")
	}

	// Retrieve the user
	user, err := uc.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (uc *UserUseCase) GetUserByUsername(username string) (*domain.User, error) {
	// Validate username
	if len(username) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}

	// Retrieve the user
	user, err := uc.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserInput represents input data for user update
type UpdateUserInput struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// UpdateUser updates user information
func (uc *UserUseCase) UpdateUser(input *UpdateUserInput) (*domain.User, error) {
	// Convert ID from string to ObjectID
	userID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Retrieve the existing user
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Validate and update email if provided
	if input.Email != "" && input.Email != user.Email {
		if !isValidEmail(input.Email) {
			return nil, errors.New("invalid email format")
		}

		// Check if the new email is already used by another user
		existingUser, err := uc.userRepo.FindByEmail(input.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("email already used by another user")
		}

		user.Email = input.Email
	}

	// Update first name if provided
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}

	// Update last name if provided
	if input.LastName != "" {
		user.LastName = input.LastName
	}

	// Update password if provided
	if input.Password != "" {
		if len(input.Password) < 6 {
			return nil, errors.New("password must be at least 6 characters long")
		}

		// Hash the new password
		hashedPassword, err := hashPassword(input.Password)
		if err != nil {
			return nil, err
		}

		user.Password = hashedPassword
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save to repository
	err = uc.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (uc *UserUseCase) DeleteUser(id string) error {
	// Convert ID from string to ObjectID
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Delete from repository
	return uc.userRepo.Delete(userID)
}

// ValidateCredentials validates user login credentials
func (uc *UserUseCase) ValidateCredentials(login string, password string) (*domain.User, error) {
	var user *domain.User
	var err error

	// Check if login is email or username
	if isValidEmail(login) {
		user, err = uc.userRepo.FindByEmail(login)
	} else {
		user, err = uc.userRepo.FindByUsername(login)
	}

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("invalid login credentials")
		}
		return nil, err
	}

	// Verify password
	if !verifyPassword(user.Password, password) {
		return nil, errors.New("invalid login credentials")
	}

	return user, nil
}

// Helper functions

// validateUserInput validates user registration input
func validateUserInput(input *RegisterUserInput) error {
	// Validate username
	if len(input.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	// Validate email
	if !isValidEmail(input.Email) {
		return errors.New("invalid email format")
	}

	// Validate password
	if len(input.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	// Simple regex for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against its hash
func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
