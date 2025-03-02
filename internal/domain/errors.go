package domain

import "errors"

// Define domain error types
var (
	// ErrNotFound represents an error when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput represents an error when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized represents an error when a user is not authorized to perform an action
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrDuplicateKey represents an error when trying to create a resource with a duplicate key
	ErrDuplicateKey = errors.New("duplicate key error")

	// ErrInternalServer represents an internal server error
	ErrInternalServer = errors.New("internal server error")
)
