package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"task-management-system/config"
	"task-management-system/internal/delivery/http/routes"
	"task-management-system/internal/infrastructure/logger"
	"task-management-system/internal/usecase"
)

// Server represents HTTP server
type Server struct {
	server *http.Server
	cfg    *config.Config
}

// NewServer creates a new HTTP server
func NewServer(
	cfg *config.Config,
	taskUseCase *usecase.TaskUseCase,
	userUseCase *usecase.UserUseCase,
	authUseCase *usecase.AuthUseCase,
) *Server {
	// Create router
	router := routes.NewRouter(taskUseCase, userUseCase, authUseCase)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server: server,
		cfg:    cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.InfoF("Starting HTTP server on port %d", s.cfg.Server.HTTP.Port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	logger.InfoF("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}
