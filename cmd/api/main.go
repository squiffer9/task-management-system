package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// TODO: Initialize usecases
	// TODO: Initialize HTTP handlers and routes

	// Create HTTP server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
		// TODO: Add handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.InfoF("Starting HTTP server on port %d", cfg.Server.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalF("Failed to start HTTP server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.InfoF("Shutting down HTTP server... (Signal: %v)", sig)

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.ErrorF("Server shutdown error: %v", err)
	}

	logger.InfoF("Server gracefully stopped")
}
