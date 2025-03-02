package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"task-management-system/config"
	"task-management-system/internal/infrastructure/logger"
	"task-management-system/internal/infrastructure/mongodb"
)

func main() {
	// Initialize logger
	if os.Getenv("APP_ENV") == "development" {
		logger.SetDefaultLevel(logger.Debug)
	} else {
		logger.SetDefaultLevel(logger.Info)
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

	// TODO: Initialize usecases
	// TODO: Initialize gRPC services

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Register gRPC services

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		logger.FatalF("Failed to listen: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		logger.InfoF("Starting gRPC server on port %d", cfg.Server.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.FatalF("Failed to start gRPC server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.InfoF("Shutting down gRPC server... (Signal: %v)", sig)

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	logger.InfoF("Server gracefully stopped")
}
