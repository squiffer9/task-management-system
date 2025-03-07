package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"task-management-system/api/proto"
	"task-management-system/internal/logger"
)

// Client represents a gRPC client
type Client struct {
	conn          *grpc.ClientConn
	taskClient    proto.TaskServiceClient
	userClient    proto.UserServiceClient
	authToken     string
	serverAddress string
}

// NewClient creates a new gRPC client
func NewClient(serverAddress string) (*Client, error) {
	// Set up a connection to the server with insecure transport (for internal network only)
	// In production, consider using TLS
	conn, err := grpc.Dial(serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Create clients
	taskClient := proto.NewTaskServiceClient(conn)
	userClient := proto.NewUserServiceClient(conn)

	return &Client{
		conn:          conn,
		taskClient:    taskClient,
		userClient:    userClient,
		serverAddress: serverAddress,
	}, nil
}

// SetAuthToken sets the authentication token for subsequent requests
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// createAuthContext creates a context with authorization metadata
func (c *Client) createAuthContext(ctx context.Context) context.Context {
	if c.authToken != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", c.authToken)
	}
	return ctx
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Task Service Methods

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, input *proto.CreateTaskRequest) (*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.taskClient.CreateTask(ctx, input)
}

// GetTask gets a task by ID
func (c *Client) GetTask(ctx context.Context, id string) (*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.taskClient.GetTask(ctx, &proto.GetTaskRequest{Id: id})
}

// UpdateTask updates a task
func (c *Client) UpdateTask(ctx context.Context, input *proto.UpdateTaskRequest) (*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.taskClient.UpdateTask(ctx, input)
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(ctx context.Context, id string, userID string) error {
	ctx = c.createAuthContext(ctx)
	_, err := c.taskClient.DeleteTask(ctx, &proto.DeleteTaskRequest{
		Id:     id,
		UserId: userID,
	})
	return err
}

// ListTasks lists tasks with optional status filter
func (c *Client) ListTasks(ctx context.Context, status proto.TaskStatus) ([]*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.taskClient.ListTasks(ctx, &proto.ListTasksRequest{
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

// AssignTask assigns a task to a user
func (c *Client) AssignTask(ctx context.Context, taskID, assigneeID, assignedBy string) (*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.taskClient.AssignTask(ctx, &proto.AssignTaskRequest{
		TaskId:     taskID,
		AssigneeId: assigneeID,
		AssignedBy: assignedBy,
	})
}

// GetUserTasks gets tasks for a user
func (c *Client) GetUserTasks(ctx context.Context, userID string) ([]*proto.TaskResponse, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.taskClient.GetUserTasks(ctx, &proto.GetUserTasksRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

// User Service Methods

// GetUser gets a user by ID
func (c *Client) GetUser(ctx context.Context, id string) (*proto.UserResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.userClient.GetUser(ctx, &proto.GetUserRequest{Id: id})
}

// ValidateToken validates a JWT token
func (c *Client) ValidateToken(ctx context.Context, token string) (*proto.ValidateTokenResponse, error) {
	return c.userClient.ValidateToken(ctx, &proto.ValidateTokenRequest{Token: token})
}
