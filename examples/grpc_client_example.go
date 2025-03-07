package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"task-management-system/api/proto"
	grpcClient "task-management-system/internal/client/grpc"
)

func main() {
	// Create a gRPC client
	client, err := grpcClient.NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	// Example: Create a new task
	createTaskExample(client)

	// Example: List tasks
	listTasksExample(client)

	// Example: Update task
	updateTaskExample(client)

	// Example: Validate token
	validateTokenExample(client)
}

func createTaskExample(client *grpcClient.Client) {
	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set auth token (in a real application, this would be obtained from authentication)
	client.SetAuthToken("sample_jwt_token")

	// Create task request
	req := &proto.CreateTaskRequest{
		Title:       "Example gRPC Task",
		Description: "This task was created using gRPC client",
		Priority:    3,
		DueDate:     timestamppb.New(time.Now().Add(7 * 24 * time.Hour)), // Due in 1 week
		CreatedBy:   "60f1a7c9e113d70001234567",                          // Example user ID
	}

	// Call the server
	resp, err := client.CreateTask(ctx, req)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
		return
	}

	fmt.Printf("Task created successfully: ID=%s, Title=%s\n", resp.Id, resp.Title)
}

func listTasksExample(client *grpcClient.Client) {
	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// List all tasks
	tasks, err := client.ListTasks(ctx, proto.TaskStatus_TASK_STATUS_UNSPECIFIED)
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		return
	}

	fmt.Printf("Retrieved %d tasks:\n", len(tasks))
	for i, task := range tasks {
		fmt.Printf("%d. %s (Status: %s, Priority: %d)\n",
			i+1, task.Title, task.Status.String(), task.Priority)
	}
}

func updateTaskExample(client *grpcClient.Client) {
	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This assumes a task with this ID exists
	taskID := "60f1a7c9e113d70001abcdef" // Example task ID

	// Update task request
	req := &proto.UpdateTaskRequest{
		Id:        taskID,
		Status:    proto.TaskStatus_TASK_STATUS_IN_PROGRESS,
		UpdatedBy: "60f1a7c9e113d70001234567", // Example user ID
	}

	// Call the server
	resp, err := client.UpdateTask(ctx, req)
	if err != nil {
		log.Printf("Failed to update task: %v", err)
		return
	}

	fmt.Printf("Task updated successfully: ID=%s, New Status=%s\n",
		resp.Id, resp.Status.String())
}

func validateTokenExample(client *grpcClient.Client) {
	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sample token (in a real application, this would be a real JWT token)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.token"

	// Validate token
	resp, err := client.ValidateToken(ctx, token)
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		return
	}

	if resp.Valid {
		fmt.Printf("Token is valid for user: %s (%s)\n", resp.Username, resp.UserId)
	} else {
		fmt.Println("Token is invalid")
	}
}
