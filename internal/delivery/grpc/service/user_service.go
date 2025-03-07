package service

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"task-management-system/api/proto"
	"task-management-system/internal/domain"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

// UserService implements the gRPC UserService
type UserService struct {
	proto.UnimplementedUserServiceServer
	userUseCase *usecase.UserUseCase
	authUseCase *usecase.AuthUseCase
}

// NewUserService creates a new UserService
func NewUserService(userUseCase *usecase.UserUseCase, authUseCase *usecase.AuthUseCase) *UserService {
	return &UserService{
		userUseCase: userUseCase,
		authUseCase: authUseCase,
	}
}

// Register registers the service with a gRPC server
func (s *UserService) Register(server *grpc.Server) {
	proto.RegisterUserServiceServer(server, s)
}

// GetUser implements the GetUser RPC method
func (s *UserService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.UserResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	// Get user
	user, err := s.userUseCase.GetUserByID(req.Id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logger.ErrorF("Failed to get user: %v", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	// Convert to response
	return &proto.UserResponse{
		Id:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// ValidateToken implements the ValidateToken RPC method
func (s *UserService) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// Validate request
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Validate token
	userID, err := s.authUseCase.ValidateToken(req.Token)
	if err != nil {
		// Return a response with valid=false instead of an error
		return &proto.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// Get username
	user, err := s.userUseCase.GetUserByID(userID)
	if err != nil {
		logger.ErrorF("Failed to get user: %v", err)
		return &proto.ValidateTokenResponse{
			UserId: userID,
			Valid:  true,
		}, nil
	}

	// Return response
	return &proto.ValidateTokenResponse{
		UserId:   userID,
		Username: user.Username,
		Valid:    true,
	}, nil
}
