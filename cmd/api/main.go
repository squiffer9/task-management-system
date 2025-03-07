package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-management-system/config"
	httpServer "task-management-system/internal/delivery/http"
	"task-management-system/internal/infrastructure/mongodb"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

func main() {
	// Initialize logger
	if os.Getenv("APP_ENV") == "development" {
		logger.SetDefaultLevel(logger.Debug)
	} else {
		logger.SetDefaultLevel(logger.Info)
	}

	logger.InfoF("Starting task management API server")

	// Load configuration
	cfg, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		logger.FatalF("Failed to load configuration: %v", err)
	}

	logger.InfoF("Configuration loaded successfully")
	logger.DebugF("Database URI: %s, Database name: %s", cfg.Database.MongoDB.URI, cfg.Database.MongoDB.Name)

	// Create MongoDB client
	client, err := mongodb.NewClient(cfg.Database.MongoDB.URI, cfg.Database.MongoDB.Timeout)
	if err != nil {
		logger.FatalF("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongodb.CloseClient(client, cfg.Database.MongoDB.Timeout); err != nil {
			logger.ErrorF("Error closing MongoDB connection: %v", err)
		}
	}()

	// Get MongoDB database
	db := mongodb.GetDatabase(client, cfg.Database.MongoDB.Name)
	logger.InfoF("Connected to MongoDB: %s", cfg.Database.MongoDB.Name)

	// Initialize repositories
	taskRepo := mongodb.NewTaskRepository(db, cfg.Database.MongoDB.Timeout)
	userRepo := mongodb.NewUserRepository(db, cfg.Database.MongoDB.Timeout)

	logger.InfoF("Repositories initialized successfully")

	// Initialize usecases
	taskUseCase := usecase.NewTaskUseCase(taskRepo, userRepo)
	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, cfg.Auth.JWT.Secret, cfg.Auth.JWT.Expiry)

	logger.InfoF("Use cases initialized successfully")

	// Create HTTP server
	server := httpServer.NewServer(cfg, taskUseCase, userUseCase, authUseCase)

	// Start HTTP server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.FatalF("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.InfoF("Shutting down server... (Signal: %v)", sig)

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Stop(ctx); err != nil {
		logger.ErrorF("Server shutdown error: %v", err)
	}

	logger.InfoF("Server gracefully stopped")
}
