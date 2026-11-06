package transport

import (
	"context"
	"core/model"
	taskv1 "database-service/api/task/v1"
	"database-service/service"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (handler *TaskHandler) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "update task request is nil")
	}
	if err := req.StartTime.CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "update task request start time is invalid")
	}

	task := &model.Task{
		Id:        req.Id,
		Title:     req.Title,
		UserId:    req.UserId,
		Duration:  time.Duration(req.Duration) * time.Second,
		StartTime: req.StartTime.AsTime(),
		Priority:  req.Priority,
	}

	if err := handler.taskService.Update(task); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (handler *TaskHandler) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "delete task request is nil")
	}
	if err := handler.taskService.Delete(req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (handler *TaskHandler) GetTaskById(ctx context.Context, req *taskv1.GetTaskByIdRequest) (*taskv1.Task, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "get task request is nil")
	}
	task, err := handler.taskService.GetById(req.Id)

	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &taskv1.Task{
		Id:              task.Id,
		Title:           task.Title,
		UserId:          task.UserId,
		StartTime:       timestamppb.New(task.StartTime),
		DurationSeconds: int32(task.Duration / time.Second),
		Priority:        task.Priority,
	}, nil
}

func (handler *TaskHandler) GetTasksByUserId(ctx context.Context, req *taskv1.GetTasksByUserIdRequest) (*taskv1.GetTasksByUserIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "get tasks by user request is nil")
	}

	tasks, err := handler.taskService.GetByUserId(req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &taskv1.GetTasksByUserIdResponse{
		Tasks: make([]*taskv1.Task, len(tasks)),
	}
	for i, task := range tasks {
		response.Tasks[i] = &taskv1.Task{
			Id:              task.Id,
			Title:           task.Title,
			UserId:          task.UserId,
			StartTime:       timestamppb.New(task.StartTime),
			DurationSeconds: int32(task.Duration / time.Second),
			Priority:        task.Priority,
		}
	}
	return response, nil
}
