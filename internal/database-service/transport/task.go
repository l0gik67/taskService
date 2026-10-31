package transport

import (
	"context"
	taskv1 "database-service/api/task/v1"
	"database-service/model"
	"database-service/service"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (handler *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*taskv1.CreateTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "create task request is nil")
	}
	if err := req.StartTime.CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "create task request start time is invalid")
	}

	task := &model.Task{
		Title:     req.Title,
		UserId:    req.UserId,
		Duration:  time.Duration(req.DurationSeconds) * time.Second,
		StartTime: req.StartTime.AsTime(),
		Priority:  req.Priority,
	}

	createdTask, err := handler.taskService.Create(task)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &taskv1.CreateTaskResponse{
		Task: &taskv1.Task{
			Id:              createdTask.Id,
			Title:           createdTask.Title,
			UserId:          createdTask.UserId,
			DurationSeconds: int32(createdTask.Duration / time.Second),
			StartTime:       timestamppb.New(createdTask.StartTime),
			Priority:        createdTask.Priority,
		},
	}, nil
}
