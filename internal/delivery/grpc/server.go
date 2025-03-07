package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"task-management-system/config"
	"task-management-system/internal/delivery/grpc/service"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

// Server represents gRPC server
type Server struct {
	server   *grpc.Server
	listener net.Listener
	cfg      *config.Config
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.Config,
	taskUseCase *usecase.TaskUseCase,
	userUseCase *usecase.UserUseCase,
	authUseCase *usecase.AuthUseCase,
) (*Server, error) {
	// Create listener
	port := fmt.Sprintf("%d", cfg.Server.GRPC.Port)
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", port))
	if err != nil {
		return nil, err
	}

	return NewServerWithListener(cfg, listener, taskUseCase, userUseCase, authUseCase)
}

// NewServerWithListener creates a new gRPC server with a provided listener (for testing)
func NewServerWithListener(
	cfg *config.Config,
	listener net.Listener,
	taskUseCase *usecase.TaskUseCase,
	userUseCase *usecase.UserUseCase,
	authUseCase *usecase.AuthUseCase,
) (*Server, error) {

	// Create gRPC server
	server := grpc.NewServer(
		grpc.ConnectionTimeout(5*time.Second),
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
	)

	// Create and register task service
	taskService := service.NewTaskService(taskUseCase, authUseCase)
	taskService.Register(server)

	// Create and register user service
	userService := service.NewUserService(userUseCase, authUseCase)
	userService.Register(server)

	// Register reflection service for gRPC tools
	reflection.Register(server)

	return &Server{
		server:   server,
		listener: listener,
		cfg:      cfg,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	logger.InfoF("Starting gRPC server on port %d", s.cfg.Server.GRPC.Port)
	return s.server.Serve(s.listener)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	logger.InfoF("Stopping gRPC server")
	s.server.GracefulStop()
}
