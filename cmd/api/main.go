package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "task-management-system/api/swagger"

	"task-management-system/config"
	httpServer "task-management-system/internal/delivery/http"
	"task-management-system/internal/infrastructure/mongodb"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

// @title Task Management System API
// @version 0.1.0
// @description API for task management system built with Go and MongoDB.
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Initialize logger
	if os.Getenv("APP_ENV") == "development" {
		logger.SetDefaultLevel(logger.LevelDebug)
	} else {
		logger.SetDefaultLevel(logger.LevelInfo)
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

	// Add Swagger handler directly to the mux router
	if router, ok := server.GetRouter().(*mux.Router); ok {
		// Define Swagger UI route
		router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"), // URL to swagger JSON doc
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DomID("swagger-ui"),
		))
		logger.InfoF("Swagger UI initialized at /swagger/")
	} else {
		logger.WarnF("Could not initialize Swagger UI - router is not of type *mux.Router")
	}

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
