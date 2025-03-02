package usecase

import (
	"errors"
	"fmt"
	"time"

	"task-management-system/internal/domain"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthUseCase handles authentication and authorization
type AuthUseCase struct {
	userRepo  domain.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(userRepo domain.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// LoginInput represents input data for user login
type LoginInput struct {
	Login    string // can be username or email
	Password string
}

// LoginOutput represents output data from user login
type LoginOutput struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
}

// Login authenticates a user and returns a JWT token
func (uc *AuthUseCase) Login(input *LoginInput) (*LoginOutput, error) {
	// Find the user by email or username
	var user *domain.User
	var err error

	if isValidEmail(input.Login) {
		user, err = uc.userRepo.FindByEmail(input.Login)
	} else {
		user, err = uc.userRepo.FindByUsername(input.Login)
	}

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("invalid login credentials")
		}
		return nil, err
	}

	// Verify password
	if !verifyPassword(user.Password, input.Password) {
		return nil, errors.New("invalid login credentials")
	}

	// Generate JWT token
	token, expiresAt, err := uc.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		UserID:      user.ID.Hex(),
		Username:    user.Username,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (uc *AuthUseCase) ValidateToken(tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token")
}

// GetUserFromToken retrieves a user by the user ID in the token
func (uc *AuthUseCase) GetUserFromToken(tokenString string) (*domain.User, error) {
	// Validate the token
	userID, err := uc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Convert ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Retrieve the user
	user, err := uc.userRepo.FindByID(userObjID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// RefreshToken refreshes a JWT token
func (uc *AuthUseCase) RefreshToken(tokenString string) (*LoginOutput, error) {
	// Validate the token
	userID, err := uc.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Convert ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Retrieve the user
	user, err := uc.userRepo.FindByID(userObjID)
	if err != nil {
		return nil, err
	}

	// Generate new JWT token
	token, expiresAt, err := uc.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		UserID:      user.ID.Hex(),
		Username:    user.Username,
	}, nil
}

// VerifyUserAccess verifies if a user has access to a resource
func (uc *AuthUseCase) VerifyUserAccess(userID string, resourceID string, resourceType string) error {
	// For now, implement a simple authorization model
	// In real-world applications, you would probably use a more sophisticated model
	// such as RBAC (Role-Based Access Control) or ABAC (Attribute-Based Access Control)

	switch resourceType {
	case "task":
		// Allow the creator of the task to access it
		// This is just a placeholder implementation
		// You should replace this with actual logic
		if userID == resourceID {
			return nil
		}
		return domain.ErrUnauthorized
	default:
		return errors.New("unknown resource type")
	}
}

// generateJWT generates a JWT token for a user
func (uc *AuthUseCase) generateJWT(user *domain.User) (string, time.Time, error) {
	// Set expiration time
	expiresAt := time.Now().Add(uc.jwtExpiry)

	// Create claims
	claims := &Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
