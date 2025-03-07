package main

import (
	"os"
	"os/signal"
	"syscall"

	"task-management-system/config"
	grpcServer "task-management-system/internal/delivery/grpc"
	"task-management-system/internal/infrastructure/mongodb"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

func main() {
	// Initialize logger
	if os.Getenv("APP_ENV") == "development" {
		logger.SetDefaultLevel(logger.LevelDebug)
	} else {
		logger.SetDefaultLevel(logger.LevelInfo)
	}

	logger.InfoF("Starting task management gRPC server")

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

	// Create gRPC server
	server, err := grpcServer.NewServer(cfg, taskUseCase, userUseCase, authUseCase)
	if err != nil {
		logger.FatalF("Failed to create gRPC server: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.FatalF("Failed to start gRPC server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.InfoF("Shutting down gRPC server... (Signal: %v)", sig)

	// Gracefully stop the server
	server.Stop()
	logger.InfoF("Server gracefully stopped")
}
