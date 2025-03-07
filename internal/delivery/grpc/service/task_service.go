package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"task-management-system/api/proto"
	"task-management-system/internal/domain"
	"task-management-system/internal/logger"
	"task-management-system/internal/usecase"
)

// TaskService implements the gRPC TaskService
type TaskService struct {
	proto.UnimplementedTaskServiceServer
	taskUseCase *usecase.TaskUseCase
	authUseCase *usecase.AuthUseCase
}

// NewTaskService creates a new TaskService
func NewTaskService(taskUseCase *usecase.TaskUseCase, authUseCase *usecase.AuthUseCase) *TaskService {
	return &TaskService{
		taskUseCase: taskUseCase,
		authUseCase: authUseCase,
	}
}

// Register registers the service with a gRPC server
func (s *TaskService) Register(server *grpc.Server) {
	proto.RegisterTaskServiceServer(server, s)
}

// getUserIDFromContext extracts user ID from context metadata
func (s *TaskService) getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	token := values[0]
	userID, err := s.authUseCase.ValidateToken(token)
	if err != nil {
		logger.ErrorF("Token validation error: %v", err)
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userID, nil
}

// CreateTask implements the CreateTask RPC method
func (s *TaskService) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.TaskResponse, error) {
	// Validate request
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	if req.Priority < 1 || req.Priority > 5 {
		return nil, status.Error(codes.InvalidArgument, "priority must be between 1 and 5")
	}

	// Get due date
	var dueDate time.Time
	if req.DueDate != nil {
		dueDate = req.DueDate.AsTime()
	}

	// Create task
	task, err := s.taskUseCase.CreateTask(&usecase.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    int(req.Priority),
		DueDate:     dueDate,
		CreatedBy:   req.CreatedBy,
	})

	if err != nil {
		logger.ErrorF("Failed to create task: %v", err)
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	// Convert to response
	return s.domainTaskToProto(task), nil
}

// GetTask implements the GetTask RPC method
func (s *TaskService) GetTask(ctx context.Context, req *proto.GetTaskRequest) (*proto.TaskResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	// Get task
	task, err := s.taskUseCase.GetTaskByID(req.Id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		logger.ErrorF("Failed to get task: %v", err)
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	// Convert to response
	return s.domainTaskToProto(task), nil
}

// UpdateTask implements the UpdateTask RPC method
func (s *TaskService) UpdateTask(ctx context.Context, req *proto.UpdateTaskRequest) (*proto.TaskResponse, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	// Get due date
	var dueDate time.Time
	if req.DueDate != nil {
		dueDate = req.DueDate.AsTime()
	}

	// Map proto status to domain status
	var taskStatus domain.TaskStatus
	switch req.Status {
	case proto.TaskStatus_TASK_STATUS_PENDING:
		taskStatus = domain.TaskStatusPending
	case proto.TaskStatus_TASK_STATUS_IN_PROGRESS:
		taskStatus = domain.TaskStatusInProgress
	case proto.TaskStatus_TASK_STATUS_COMPLETED:
		taskStatus = domain.TaskStatusCompleted
	}

	// Update task
	task, err := s.taskUseCase.UpdateTask(&usecase.UpdateTaskInput{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Status:      taskStatus,
		Priority:    int(req.Priority),
		DueDate:     dueDate,
		UpdatedBy:   req.UpdatedBy,
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "unauthorized to update this task")
		}
		logger.ErrorF("Failed to update task: %v", err)
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	// Convert to response
	return s.domainTaskToProto(task), nil
}

// DeleteTask implements the DeleteTask RPC method
func (s *TaskService) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*emptypb.Empty, error) {
	// Validate request
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}

	// Delete task
	err := s.taskUseCase.DeleteTask(req.Id, req.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "unauthorized to delete this task")
		}
		logger.ErrorF("Failed to delete task: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &emptypb.Empty{}, nil
}

// ListTasks implements the ListTasks RPC method
func (s *TaskService) ListTasks(ctx context.Context, req *proto.ListTasksRequest) (*proto.ListTasksResponse, error) {
	// Map proto status to domain status
	var taskStatus domain.TaskStatus
	switch req.Status {
	case proto.TaskStatus_TASK_STATUS_PENDING:
		taskStatus = domain.TaskStatusPending
	case proto.TaskStatus_TASK_STATUS_IN_PROGRESS:
		taskStatus = domain.TaskStatusInProgress
	case proto.TaskStatus_TASK_STATUS_COMPLETED:
		taskStatus = domain.TaskStatusCompleted
	}

	// Get tasks
	var tasks []*domain.Task
	var err error
	if req.Status == proto.TaskStatus_TASK_STATUS_UNSPECIFIED {
		tasks, err = s.taskUseCase.ListTasks(nil)
	} else {
		tasks, err = s.taskUseCase.ListTasks(&usecase.ListTasksInput{
			Status: taskStatus,
		})
	}

	if err != nil {
		logger.ErrorF("Failed to list tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	// Convert to response
	resp := &proto.ListTasksResponse{
		Tasks: make([]*proto.TaskResponse, 0, len(tasks)),
	}

	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, s.domainTaskToProto(task))
	}

	return resp, nil
}

// AssignTask implements the AssignTask RPC method
func (s *TaskService) AssignTask(ctx context.Context, req *proto.AssignTaskRequest) (*proto.TaskResponse, error) {
	// Validate request
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task id is required")
	}
	if req.AssigneeId == "" {
		return nil, status.Error(codes.InvalidArgument, "assignee id is required")
	}

	// Assign task
	task, err := s.taskUseCase.AssignTask(&usecase.AssignTaskInput{
		TaskID:     req.TaskId,
		AssigneeID: req.AssigneeId,
		AssignedBy: req.AssignedBy,
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task or user not found")
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "unauthorized to assign this task")
		}
		logger.ErrorF("Failed to assign task: %v", err)
		return nil, status.Error(codes.Internal, "failed to assign task")
	}

	// Convert to response
	return s.domainTaskToProto(task), nil
}

// GetUserTasks implements the GetUserTasks RPC method
func (s *TaskService) GetUserTasks(ctx context.Context, req *proto.GetUserTasksRequest) (*proto.ListTasksResponse, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	// Get user tasks
	tasks, err := s.taskUseCase.GetUserTasks(req.UserId)
	if err != nil {
		logger.ErrorF("Failed to get user tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to get user tasks")
	}

	// Convert to response
	resp := &proto.ListTasksResponse{
		Tasks: make([]*proto.TaskResponse, 0, len(tasks)),
	}

	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, s.domainTaskToProto(task))
	}

	return resp, nil
}

// domainTaskToProto converts a domain task to proto task
func (s *TaskService) domainTaskToProto(task *domain.Task) *proto.TaskResponse {
	// Map domain status to proto status
	var status proto.TaskStatus
	switch task.Status {
	case domain.TaskStatusPending:
		status = proto.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusInProgress:
		status = proto.TaskStatus_TASK_STATUS_IN_PROGRESS
	case domain.TaskStatusCompleted:
		status = proto.TaskStatus_TASK_STATUS_COMPLETED
	default:
		status = proto.TaskStatus_TASK_STATUS_UNSPECIFIED
	}

	// Convert to proto
	protoTask := &proto.TaskResponse{
		Id:          task.ID.Hex(),
		Title:       task.Title,
		Description: task.Description,
		Status:      status,
		Priority:    int32(task.Priority),
		CreatedBy:   task.CreatedBy.Hex(),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}

	// Add due date if set
	if !task.DueDate.IsZero() {
		protoTask.DueDate = timestamppb.New(task.DueDate)
	}

	// Add assigned to if set
	if !task.AssignedTo.IsZero() {
		protoTask.AssignedTo = task.AssignedTo.Hex()
	}

	return protoTask
}
