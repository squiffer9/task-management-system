package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
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
	log.Printf("DEBUG: success read mongoDB : %s", db)
	log.Printf("Connected to MongoDB: %s", cfg.Database.MongoDB.Name)

	// TODO: Initialize repositories
	// TODO: Initialize usecases
	// TODO: Initialize gRPC services

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Register gRPC services

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %d", cfg.Server.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
}
