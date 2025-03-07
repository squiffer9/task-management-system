package integration

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	"task-management-system/api/proto"
	"task-management-system/config"
	grpcServer "task-management-system/internal/delivery/grpc"
	"task-management-system/internal/domain"
	"task-management-system/internal/infrastructure/mongodb"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

const bufSize = 1024 * 1024

var (
	listener *bufconn.Listener
	cfg      *config.Config
	client   *grpc.ClientConn
)

func TestMain(m *testing.M) {
	// Set up
	setup()

	// Run tests
	code := m.Run()

	// Tear down
	teardown()

	os.Exit(code)
}

func setup() {
	// Initialize logger
	logger.SetDefaultLevel(logger.LevelInfo)

	// Load configuration
	var err error
	cfg, err = config.LoadConfig("../../config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override database name for tests
	cfg.Database.MongoDB.Name = "task_management_test"

	// Create MongoDB client
	mongoClient, err := mongodb.NewClient(cfg.Database.MongoDB.URI, cfg.Database.MongoDB.Timeout)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Get MongoDB database
	db := mongodb.GetDatabase(mongoClient, cfg.Database.MongoDB.Name)

	// Drop database to ensure clean state
	if err := db.Drop(context.Background()); err != nil {
		log.Fatalf("Failed to drop test database: %v", err)
	}

	// Initialize repositories
	taskRepo := mongodb.NewTaskRepository(db, cfg.Database.MongoDB.Timeout)
	userRepo := mongodb.NewUserRepository(db, cfg.Database.MongoDB.Timeout)

	// Initialize usecases
	taskUseCase := usecase.NewTaskUseCase(taskRepo, userRepo)
	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, cfg.Auth.JWT.Secret, cfg.Auth.JWT.Expiry)

	// Create a buffer for gRPC
	listener = bufconn.Listen(bufSize)

	// Create and start gRPC server with the buffer listener instead of a real TCP listener
	server, err := grpcServer.NewServerWithListener(cfg, listener, taskUseCase, userUseCase, authUseCase)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start test gRPC server: %v", err)
		}
	}()

	// Create a client connection
	client, err = grpc.Dial(
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	// Create a test user
	createTestUser(userRepo)
}

func teardown() {
	if client != nil {
		client.Close()
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return listener.Dial()
}

func createTestUser(userRepo domain.UserRepository) {
	// Hash password manually instead of using the usecase function
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:        testUserID(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FirstName: "Test",
		LastName:  "User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := userRepo.Create(user); err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}
}

// Helper functions

func testUserID() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("60f1a7c9e113d70001234567")
	return id
}

// Test cases

func TestTaskService_CreateTask(t *testing.T) {
	taskClient := proto.NewTaskServiceClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create task
	req := &proto.CreateTaskRequest{
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    3,
		DueDate:     timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedBy:   testUserID().Hex(),
	}

	resp, err := taskClient.CreateTask(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Id)
	assert.Equal(t, req.Title, resp.Title)
	assert.Equal(t, req.Description, resp.Description)
	assert.Equal(t, proto.TaskStatus_TASK_STATUS_PENDING, resp.Status)
	assert.Equal(t, int32(3), resp.Priority)
	assert.Equal(t, testUserID().Hex(), resp.CreatedBy)
}

func TestTaskService_GetTask(t *testing.T) {
	taskClient := proto.NewTaskServiceClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First create a task
	createReq := &proto.CreateTaskRequest{
		Title:       "Task to Get",
		Description: "This is a task to get",
		Priority:    2,
		DueDate:     timestamppb.New(time.Now().Add(48 * time.Hour)),
		CreatedBy:   testUserID().Hex(),
	}

	createResp, err := taskClient.CreateTask(ctx, createReq)
	require.NoError(t, err)
	taskID := createResp.Id

	// Get the task
	getResp, err := taskClient.GetTask(ctx, &proto.GetTaskRequest{Id: taskID})
	require.NoError(t, err)
	assert.Equal(t, taskID, getResp.Id)
	assert.Equal(t, createReq.Title, getResp.Title)
	assert.Equal(t, createReq.Description, getResp.Description)
}

func TestTaskService_UpdateTask(t *testing.T) {
	taskClient := proto.NewTaskServiceClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First create a task
	createReq := &proto.CreateTaskRequest{
		Title:       "Task to Update",
		Description: "This is a task to update",
		Priority:    1,
		CreatedBy:   testUserID().Hex(),
	}

	createResp, err := taskClient.CreateTask(ctx, createReq)
	require.NoError(t, err)
	taskID := createResp.Id

	// Update the task
	updateReq := &proto.UpdateTaskRequest{
		Id:          taskID,
		Title:       "Updated Task",
		Description: "This task has been updated",
		Status:      proto.TaskStatus_TASK_STATUS_IN_PROGRESS,
		Priority:    4,
		UpdatedBy:   testUserID().Hex(),
	}

	updateResp, err := taskClient.UpdateTask(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, taskID, updateResp.Id)
	assert.Equal(t, updateReq.Title, updateResp.Title)
	assert.Equal(t, updateReq.Description, updateResp.Description)
	assert.Equal(t, updateReq.Status, updateResp.Status)
	assert.Equal(t, updateReq.Priority, updateResp.Priority)
}

func TestTaskService_ListTasks(t *testing.T) {
	taskClient := proto.NewTaskServiceClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create multiple tasks
	for i := 0; i < 3; i++ {
		req := &proto.CreateTaskRequest{
			Title:       fmt.Sprintf("List Task %d", i+1),
			Description: fmt.Sprintf("Task for listing test %d", i+1),
			Priority:    int32(i + 1),
			CreatedBy:   testUserID().Hex(),
		}
		_, err := taskClient.CreateTask(ctx, req)
		require.NoError(t, err)
	}

	// List all tasks
	listResp, err := taskClient.ListTasks(ctx, &proto.ListTasksRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listResp.Tasks), 3)

	// List only pending tasks
	pendingResp, err := taskClient.ListTasks(ctx, &proto.ListTasksRequest{
		Status: proto.TaskStatus_TASK_STATUS_PENDING,
	})
	require.NoError(t, err)
	for _, task := range pendingResp.Tasks {
		assert.Equal(t, proto.TaskStatus_TASK_STATUS_PENDING, task.Status)
	}
}

func TestUserService_GetUser(t *testing.T) {
	userClient := proto.NewUserServiceClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the test user
	resp, err := userClient.GetUser(ctx, &proto.GetUserRequest{Id: testUserID().Hex()})
	require.NoError(t, err)
	assert.Equal(t, testUserID().Hex(), resp.Id)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test", resp.FirstName)
	assert.Equal(t, "User", resp.LastName)
}
