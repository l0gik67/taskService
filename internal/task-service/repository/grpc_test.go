package repository_test

import (
	"context"
	"core/model"
	taskv1 "database-service/api/task/v1"
	"errors"
	"strings"
	"task-service/repository"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type taskServiceClientStub struct {
	createTaskFunc       func(context.Context, *taskv1.CreateTaskRequest, ...grpc.CallOption) (*taskv1.CreateTaskResponse, error)
	updateTaskFunc       func(context.Context, *taskv1.UpdateTaskRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	deleteTaskFunc       func(context.Context, *taskv1.DeleteTaskRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	getTaskByIDFunc      func(context.Context, *taskv1.GetTaskByIdRequest, ...grpc.CallOption) (*taskv1.Task, error)
	getTasksByUserIDFunc func(context.Context, *taskv1.GetTasksByUserIdRequest, ...grpc.CallOption) (*taskv1.GetTasksByUserIdResponse, error)
}

var _ taskv1.TaskServiceClient = (*taskServiceClientStub)(nil)

func (stub *taskServiceClientStub) CreateTask(ctx context.Context, request *taskv1.CreateTaskRequest, options ...grpc.CallOption) (*taskv1.CreateTaskResponse, error) {
	return stub.createTaskFunc(ctx, request, options...)
}

func (stub *taskServiceClientStub) UpdateTask(ctx context.Context, request *taskv1.UpdateTaskRequest, options ...grpc.CallOption) (*emptypb.Empty, error) {
	return stub.updateTaskFunc(ctx, request, options...)
}

func (stub *taskServiceClientStub) DeleteTask(ctx context.Context, request *taskv1.DeleteTaskRequest, options ...grpc.CallOption) (*emptypb.Empty, error) {
	return stub.deleteTaskFunc(ctx, request, options...)
}

func (stub *taskServiceClientStub) GetTaskById(ctx context.Context, request *taskv1.GetTaskByIdRequest, options ...grpc.CallOption) (*taskv1.Task, error) {
	return stub.getTaskByIDFunc(ctx, request, options...)
}

func (stub *taskServiceClientStub) GetTasksByUserId(ctx context.Context, request *taskv1.GetTasksByUserIdRequest, options ...grpc.CallOption) (*taskv1.GetTasksByUserIdResponse, error) {
	return stub.getTasksByUserIDFunc(ctx, request, options...)
}

func TestGRPCTaskRepository_GetByIdMapsRequestAndResponse(t *testing.T) {
	ctx := context.Background()
	responseTask := testProtoTask(17, 42, "Task")
	calls := 0
	client := &taskServiceClientStub{
		getTaskByIDFunc: func(gotCtx context.Context, request *taskv1.GetTaskByIdRequest, _ ...grpc.CallOption) (*taskv1.Task, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("GetTaskById did not receive the original context")
			}
			if request.Id != 17 {
				t.Fatalf("request id = %d, want 17", request.Id)
			}
			return responseTask, nil
		},
	}

	task, err := repository.NewGRPCTaskRepository(client).GetById(ctx, 17)

	if err != nil {
		t.Fatalf("GetById returned error: %v", err)
	}
	assertTaskMatchesProto(t, task, responseTask)
	if calls != 1 {
		t.Fatalf("GetTaskById calls = %d, want 1", calls)
	}
}

func TestGRPCTaskRepository_GetByFiltersMapsRequestAndResponse(t *testing.T) {
	ctx := context.Background()
	userID := int64(42)
	filters := model.TaskFilters{UserID: &userID}
	responseTasks := []*taskv1.Task{
		testProtoTask(17, userID, "First task"),
		testProtoTask(18, userID, "Second task"),
	}
	calls := 0
	client := &taskServiceClientStub{
		getTasksByUserIDFunc: func(gotCtx context.Context, request *taskv1.GetTasksByUserIdRequest, _ ...grpc.CallOption) (*taskv1.GetTasksByUserIdResponse, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("GetTasksByUserId did not receive the original context")
			}
			if request.UserId != userID {
				t.Fatalf("request user id = %d, want %d", request.UserId, userID)
			}
			return &taskv1.GetTasksByUserIdResponse{Tasks: responseTasks}, nil
		},
	}

	tasks, err := repository.NewGRPCTaskRepository(client).GetByFilters(ctx, filters)

	if err != nil {
		t.Fatalf("GetByFilters returned error: %v", err)
	}
	if len(tasks) != len(responseTasks) {
		t.Fatalf("task count = %d, want %d", len(tasks), len(responseTasks))
	}
	for index := range responseTasks {
		assertTaskMatchesProto(t, tasks[index], responseTasks[index])
	}
	if calls != 1 {
		t.Fatalf("GetTasksByUserId calls = %d, want 1", calls)
	}
}

func TestGRPCTaskRepository_GetByFiltersRejectsEmptyFiltersBeforeRPC(t *testing.T) {
	calls := 0
	client := &taskServiceClientStub{
		getTasksByUserIDFunc: func(context.Context, *taskv1.GetTasksByUserIdRequest, ...grpc.CallOption) (*taskv1.GetTasksByUserIdResponse, error) {
			calls++
			return nil, nil
		},
	}

	tasks, err := repository.NewGRPCTaskRepository(client).GetByFilters(context.Background(), model.TaskFilters{})

	if !errors.Is(err, model.ErrUserIDFilterRequired) {
		t.Fatalf("error = %v, want %v", err, model.ErrUserIDFilterRequired)
	}
	if tasks != nil {
		t.Fatalf("tasks = %+v, want nil", tasks)
	}
	if calls != 0 {
		t.Fatalf("GetTasksByUserId calls = %d, want 0", calls)
	}
}

func TestGRPCTaskRepository_CreateMapsRequestAndResponseWithoutMutation(t *testing.T) {
	ctx := context.Background()
	startTime := testStartTime()
	task := &model.Task{
		Title:     "Task",
		UserId:    42,
		Duration:  90 * time.Second,
		StartTime: startTime,
		Priority:  3,
	}
	originalTask := *task
	responseTask := testProtoTask(17, 42, "Task")
	responseTask.DurationSeconds = 90
	responseTask.Priority = 3
	calls := 0
	client := &taskServiceClientStub{
		createTaskFunc: func(gotCtx context.Context, request *taskv1.CreateTaskRequest, _ ...grpc.CallOption) (*taskv1.CreateTaskResponse, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("CreateTask did not receive the original context")
			}
			if request.Title != task.Title || request.UserId != task.UserId || request.DurationSeconds != 90 || request.Priority != task.Priority {
				t.Fatalf("CreateTask request = %+v", request)
			}
			if request.StartTime == nil || !request.StartTime.AsTime().Equal(startTime) {
				t.Fatalf("request start time = %v, want %v", request.StartTime, startTime)
			}
			return &taskv1.CreateTaskResponse{Task: responseTask}, nil
		},
	}

	createdTask, err := repository.NewGRPCTaskRepository(client).Create(ctx, task)

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	assertTaskMatchesProto(t, createdTask, responseTask)
	if *task != originalTask {
		t.Fatalf("input task was mutated: got %+v, want %+v", *task, originalTask)
	}
	if calls != 1 {
		t.Fatalf("CreateTask calls = %d, want 1", calls)
	}
}

func TestGRPCTaskRepository_UpdateMapsRequest(t *testing.T) {
	ctx := context.Background()
	startTime := testStartTime()
	task := &model.Task{
		Id:        17,
		Title:     "Updated task",
		UserId:    42,
		Duration:  120 * time.Second,
		StartTime: startTime,
		Priority:  4,
	}
	originalTask := *task
	calls := 0
	client := &taskServiceClientStub{
		updateTaskFunc: func(gotCtx context.Context, request *taskv1.UpdateTaskRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("UpdateTask did not receive the original context")
			}
			if request.Id != task.Id || request.Title != task.Title || request.UserId != task.UserId || request.Duration != 120 || request.Priority != task.Priority {
				t.Fatalf("UpdateTask request = %+v", request)
			}
			if request.StartTime == nil || !request.StartTime.AsTime().Equal(startTime) {
				t.Fatalf("request start time = %v, want %v", request.StartTime, startTime)
			}
			return &emptypb.Empty{}, nil
		},
	}

	err := repository.NewGRPCTaskRepository(client).Update(ctx, task)

	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if *task != originalTask {
		t.Fatalf("input task was mutated: got %+v, want %+v", *task, originalTask)
	}
	if calls != 1 {
		t.Fatalf("UpdateTask calls = %d, want 1", calls)
	}
}

func TestGRPCTaskRepository_DeleteMapsRequest(t *testing.T) {
	ctx := context.Background()
	calls := 0
	client := &taskServiceClientStub{
		deleteTaskFunc: func(gotCtx context.Context, request *taskv1.DeleteTaskRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("DeleteTask did not receive the original context")
			}
			if request.Id != 17 {
				t.Fatalf("request id = %d, want 17", request.Id)
			}
			return &emptypb.Empty{}, nil
		},
	}

	err := repository.NewGRPCTaskRepository(client).Delete(ctx, 17)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("DeleteTask calls = %d, want 1", calls)
	}
}

func TestGRPCTaskRepository_PropagatesRPCErrors(t *testing.T) {
	rpcErr := status.Error(codes.Unavailable, "database-service unavailable")
	validTask := &model.Task{
		Title:     "Task",
		UserId:    42,
		Duration:  time.Second,
		StartTime: testStartTime(),
		Priority:  1,
	}
	userID := int64(42)

	for _, testCase := range []struct {
		name   string
		client *taskServiceClientStub
		call   func(*repository.GRPCTaskRepository) error
	}{
		{
			name: "get by id",
			client: &taskServiceClientStub{
				getTaskByIDFunc: func(context.Context, *taskv1.GetTaskByIdRequest, ...grpc.CallOption) (*taskv1.Task, error) {
					return nil, rpcErr
				},
			},
			call: func(repo *repository.GRPCTaskRepository) error {
				_, err := repo.GetById(context.Background(), 17)
				return err
			},
		},
		{
			name: "get by filters",
			client: &taskServiceClientStub{
				getTasksByUserIDFunc: func(context.Context, *taskv1.GetTasksByUserIdRequest, ...grpc.CallOption) (*taskv1.GetTasksByUserIdResponse, error) {
					return nil, rpcErr
				},
			},
			call: func(repo *repository.GRPCTaskRepository) error {
				_, err := repo.GetByFilters(context.Background(), model.TaskFilters{UserID: &userID})
				return err
			},
		},
		{
			name: "create",
			client: &taskServiceClientStub{
				createTaskFunc: func(context.Context, *taskv1.CreateTaskRequest, ...grpc.CallOption) (*taskv1.CreateTaskResponse, error) {
					return nil, rpcErr
				},
			},
			call: func(repo *repository.GRPCTaskRepository) error {
				_, err := repo.Create(context.Background(), validTask)
				return err
			},
		},
		{
			name: "update",
			client: &taskServiceClientStub{
				updateTaskFunc: func(context.Context, *taskv1.UpdateTaskRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, rpcErr
				},
			},
			call: func(repo *repository.GRPCTaskRepository) error {
				return repo.Update(context.Background(), validTask)
			},
		},
		{
			name: "delete",
			client: &taskServiceClientStub{
				deleteTaskFunc: func(context.Context, *taskv1.DeleteTaskRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, rpcErr
				},
			},
			call: func(repo *repository.GRPCTaskRepository) error {
				return repo.Delete(context.Background(), 17)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.call(repository.NewGRPCTaskRepository(testCase.client))

			if !errors.Is(err, rpcErr) {
				t.Fatalf("error = %v, want original RPC error %v", err, rpcErr)
			}
			if status.Code(err) != codes.Unavailable {
				t.Fatalf("status code = %v, want %v", status.Code(err), codes.Unavailable)
			}
		})
	}
}

func TestGRPCTaskRepository_GetByIdPreservesNotFoundStatus(t *testing.T) {
	rpcErr := status.Error(codes.NotFound, "task not found")
	client := &taskServiceClientStub{
		getTaskByIDFunc: func(context.Context, *taskv1.GetTaskByIdRequest, ...grpc.CallOption) (*taskv1.Task, error) {
			return nil, rpcErr
		},
	}

	_, err := repository.NewGRPCTaskRepository(client).GetById(context.Background(), 17)

	if !errors.Is(err, rpcErr) {
		t.Fatalf("error = %v, want original RPC error %v", err, rpcErr)
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestGRPCTaskRepository_RejectsLossyDurationBeforeRPC(t *testing.T) {
	calls := 0
	client := &taskServiceClientStub{
		createTaskFunc: func(context.Context, *taskv1.CreateTaskRequest, ...grpc.CallOption) (*taskv1.CreateTaskResponse, error) {
			calls++
			return nil, nil
		},
	}
	task := &model.Task{
		Title:     "Task",
		UserId:    42,
		Duration:  time.Second + time.Nanosecond,
		StartTime: testStartTime(),
		Priority:  1,
	}

	createdTask, err := repository.NewGRPCTaskRepository(client).Create(context.Background(), task)

	if err == nil || !strings.Contains(err.Error(), "whole seconds") {
		t.Fatalf("error = %v, want whole-seconds mapping error", err)
	}
	if createdTask != nil {
		t.Fatalf("created task = %+v, want nil", createdTask)
	}
	if calls != 0 {
		t.Fatalf("CreateTask calls = %d, want 0", calls)
	}
}

func testProtoTask(id int64, userID int64, title string) *taskv1.Task {
	return &taskv1.Task{
		Id:              id,
		Title:           title,
		UserId:          userID,
		DurationSeconds: 60,
		StartTime:       timestamppb.New(testStartTime()),
		Priority:        2,
	}
}

func testStartTime() time.Time {
	return time.Date(2026, time.November, 7, 12, 30, 0, 0, time.UTC)
}

func assertTaskMatchesProto(t *testing.T, task *model.Task, protoTask *taskv1.Task) {
	t.Helper()
	if task == nil {
		t.Fatal("mapped task is nil")
	}
	if task.Id != protoTask.Id || task.Title != protoTask.Title || task.UserId != protoTask.UserId || task.Duration != time.Duration(protoTask.DurationSeconds)*time.Second || task.Priority != protoTask.Priority {
		t.Fatalf("mapped task = %+v, protobuf task = %+v", task, protoTask)
	}
	if !task.StartTime.Equal(protoTask.StartTime.AsTime()) {
		t.Fatalf("mapped start time = %v, want %v", task.StartTime, protoTask.StartTime.AsTime())
	}
}
