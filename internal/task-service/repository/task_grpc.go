package repository

import (
	"context"
	"core/model"
	taskv1 "database-service/api/task/v1"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	minInt32Seconds = -1 << 31
	maxInt32Seconds = 1<<31 - 1
)

type GRPCTaskRepository struct {
	client taskv1.TaskServiceClient
}

var _ TaskRepository = (*GRPCTaskRepository)(nil)

func NewGRPCTaskRepository(client taskv1.TaskServiceClient) *GRPCTaskRepository {
	return &GRPCTaskRepository{client: client}
}

func (repository *GRPCTaskRepository) GetById(ctx context.Context, id int64) (*model.Task, error) {
	task, err := repository.client.GetTaskById(ctx, &taskv1.GetTaskByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}

	return taskFromProto(task)
}

func (repository *GRPCTaskRepository) GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	request := &taskv1.GetTasksByUserIdRequest{UserId: *filters.UserID}
	if filters.BoardID != nil {
		request.BoardId = *filters.BoardID
	}
	response, err := repository.client.GetTasksByUserId(ctx, request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil task list response")
	}

	tasks := make([]*model.Task, len(response.Tasks))
	for index, task := range response.Tasks {
		mappedTask, err := taskFromProto(task)
		if err != nil {
			return nil, fmt.Errorf("map task at index %d: %w", index, err)
		}
		tasks[index] = mappedTask
	}

	return tasks, nil
}

func (repository *GRPCTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	request, err := createTaskRequest(task)
	if err != nil {
		return nil, err
	}

	response, err := repository.client.CreateTask(ctx, request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil create task response")
	}

	return taskFromProto(response.Task)
}

func (repository *GRPCTaskRepository) Update(ctx context.Context, task *model.Task) error {
	request, err := updateTaskRequest(task)
	if err != nil {
		return err
	}

	_, err = repository.client.UpdateTask(ctx, request)
	return err
}

func (repository *GRPCTaskRepository) Delete(ctx context.Context, id int64) error {
	_, err := repository.client.DeleteTask(ctx, &taskv1.DeleteTaskRequest{Id: id})
	return err
}

func createTaskRequest(task *model.Task) (*taskv1.CreateTaskRequest, error) {
	if task == nil {
		return nil, errors.New("task is nil")
	}

	durationSeconds, err := durationToProtoSeconds(task.Duration)
	if err != nil {
		return nil, err
	}
	startTime, err := timeToProto(task.StartTime)
	if err != nil {
		return nil, err
	}

	return &taskv1.CreateTaskRequest{
		Title:           task.Title,
		Description:     task.Description,
		UserId:          task.UserId,
		ColumnId:        task.ColumnId,
		Position:        task.Position,
		Progress:        task.Progress,
		IsDone:          task.IsDone,
		DurationSeconds: durationSeconds,
		StartTime:       startTime,
		Priority:        task.Priority,
	}, nil
}

func updateTaskRequest(task *model.Task) (*taskv1.UpdateTaskRequest, error) {
	if task == nil {
		return nil, errors.New("task is nil")
	}

	durationSeconds, err := durationToProtoSeconds(task.Duration)
	if err != nil {
		return nil, err
	}
	startTime, err := timeToProto(task.StartTime)
	if err != nil {
		return nil, err
	}

	return &taskv1.UpdateTaskRequest{
		Id:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		UserId:      task.UserId,
		ColumnId:    task.ColumnId,
		Position:    task.Position,
		Progress:    task.Progress,
		IsDone:      task.IsDone,
		Duration:    durationSeconds,
		StartTime:   startTime,
		Priority:    task.Priority,
	}, nil
}

func taskFromProto(task *taskv1.Task) (*model.Task, error) {
	if task == nil {
		return nil, errors.New("database-service returned a nil task")
	}
	if task.StartTime == nil {
		return nil, errors.New("database-service returned a task without start_time")
	}
	if err := task.StartTime.CheckValid(); err != nil {
		return nil, fmt.Errorf("database-service returned an invalid task start_time: %w", err)
	}

	result := &model.Task{
		Id:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		UserId:      task.UserId,
		ColumnId:    task.ColumnId,
		Position:    task.Position,
		Progress:    task.Progress,
		IsDone:      task.IsDone,
		Duration:    time.Duration(task.DurationSeconds) * time.Second,
		StartTime:   task.StartTime.AsTime(),
		Priority:    task.Priority,
	}
	if task.CompletedAt != nil {
		if err := task.CompletedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("database-service returned an invalid task completed_at: %w", err)
		}
		completedAt := task.CompletedAt.AsTime()
		result.CompletedAt = &completedAt
	}
	return result, nil
}

func durationToProtoSeconds(duration time.Duration) (int32, error) {
	if duration%time.Second != 0 {
		return 0, errors.New("task duration must contain whole seconds")
	}

	seconds := int64(duration / time.Second)
	if seconds < minInt32Seconds || seconds > maxInt32Seconds {
		return 0, errors.New("task duration exceeds the protobuf int32 seconds range")
	}

	return int32(seconds), nil
}

func timeToProto(value time.Time) (*timestamppb.Timestamp, error) {
	timestamp := timestamppb.New(value)
	if err := timestamp.CheckValid(); err != nil {
		return nil, fmt.Errorf("task start_time is invalid: %w", err)
	}
	return timestamp, nil
}
