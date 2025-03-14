package main

import (
	"context"
	"net/http"
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
// @version 1.0.0
// @description RESTful API for managing tasks, with MongoDB backend and JWT authentication
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @basePath /api/v1
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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
		// Create a handler to serve the API specification file directly from the file system
		router.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "api/swagger/doc.json")
		})

		// Define Swagger UI route
		router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"), // URL to swagger JSON doc
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("list"),
			httpSwagger.DomID("swagger-ui"),
			httpSwagger.PersistAuthorization(true),
		))
		logger.InfoF("Swagger UI initialized at /swagger/, using spec from /swagger/doc.json")
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
