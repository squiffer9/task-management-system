package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-management-system/config"
	"task-management-system/internal/infrastructure/mongodb"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create MongoDB client
	client, err := mongodb.NewClient(cfg.Database.MongoDB.URI, cfg.Database.MongoDB.Timeout)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.CloseClient(client, cfg.Database.MongoDB.Timeout)

	// Get MongoDB database
	db := mongodb.GetDatabase(client, cfg.Database.MongoDB.Name)
	log.Printf("DEBUG: succese read mongoDB: %s", db)
	log.Printf("Connected to MongoDB: %s", cfg.Database.MongoDB.Name)

	// TODO: Initialize repositories
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
		log.Printf("Starting HTTP server on port %d", cfg.Server.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HTTP server...")

	// TODO: Implement graceful shutdown for HTTP server
}
