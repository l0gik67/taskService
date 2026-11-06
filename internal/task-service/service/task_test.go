package service_test

import (
	"context"
	"core/model"
	"errors"
	"task-service/repository"
	"task-service/service"
	"testing"
	"time"
)

type taskRepositoryStub struct {
	getByIDFunc      func(ctx context.Context, id int64) (*model.Task, error)
	getByFiltersFunc func(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error)
	createFunc       func(ctx context.Context, task *model.Task) (*model.Task, error)
	updateFunc       func(ctx context.Context, task *model.Task) error
	deleteFunc       func(ctx context.Context, id int64) error
}

var _ repository.TaskRepository = (*taskRepositoryStub)(nil)

func (stub *taskRepositoryStub) GetById(ctx context.Context, id int64) (*model.Task, error) {
	return stub.getByIDFunc(ctx, id)
}

func (stub *taskRepositoryStub) GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
	return stub.getByFiltersFunc(ctx, filters)
}

func (stub *taskRepositoryStub) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	return stub.createFunc(ctx, task)
}

func (stub *taskRepositoryStub) Update(ctx context.Context, task *model.Task) error {
	return stub.updateFunc(ctx, task)
}

func (stub *taskRepositoryStub) Delete(ctx context.Context, id int64) error {
	return stub.deleteFunc(ctx, id)
}

func TestCreate_RejectsNegativeDurationBeforeCallingRepository(t *testing.T) {
	createCalls := 0
	repo := &taskRepositoryStub{
		createFunc: func(ctx context.Context, task *model.Task) (*model.Task, error) {
			createCalls++
			return task, nil
		},
	}
	taskService := service.NewTaskService(repo)
	task := &model.Task{
		Title:    "Task with invalid duration",
		Priority: 1,
		Duration: -time.Second,
	}

	createdTask, err := taskService.Create(context.Background(), task)

	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "task duration can't be less than 0" {
		t.Fatalf("error = %q, want negative duration validation error", err)
	}
	if createdTask != nil {
		t.Fatalf("created task = %+v, want nil", createdTask)
	}
	if createCalls != 0 {
		t.Fatalf("repository Create calls = %d, want 0", createCalls)
	}
}

func TestCreate_ValidTaskIsPassedToRepositoryWithoutMutation(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "request context")
	task := &model.Task{
		Id:        7,
		Title:     "Valid task",
		UserId:    42,
		Duration:  15 * time.Minute,
		StartTime: time.Date(2026, time.November, 7, 10, 0, 0, 0, time.UTC),
		Priority:  2,
	}
	originalTask := *task
	createCalls := 0
	repo := &taskRepositoryStub{
		createFunc: func(gotCtx context.Context, gotTask *model.Task) (*model.Task, error) {
			createCalls++
			if gotCtx != ctx {
				t.Fatal("service did not preserve context")
			}
			if gotTask != task {
				t.Fatal("service did not pass the original task pointer")
			}
			return gotTask, nil
		},
	}
	taskService := service.NewTaskService(repo)

	createdTask, err := taskService.Create(ctx, task)

	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if createCalls != 1 {
		t.Fatalf("repository Create calls = %d, want 1", createCalls)
	}
	if createdTask != task {
		t.Fatal("Create did not return the repository result")
	}
	if *task != originalTask {
		t.Fatalf("task was mutated: got %+v, want %+v", *task, originalTask)
	}
}

func TestGetById_DelegatesToRepository(t *testing.T) {
	ctx := context.Background()
	wantTask := &model.Task{Id: 17}
	calls := 0
	repo := &taskRepositoryStub{
		getByIDFunc: func(gotCtx context.Context, id int64) (*model.Task, error) {
			calls++
			if gotCtx != ctx || id != 17 {
				t.Fatalf("GetById arguments = (%v, %d), want (%v, 17)", gotCtx, id, ctx)
			}
			return wantTask, nil
		},
	}

	gotTask, err := service.NewTaskService(repo).GetById(ctx, 17)

	if err != nil || gotTask != wantTask {
		t.Fatalf("GetById result = (%+v, %v), want (%+v, nil)", gotTask, err, wantTask)
	}
	if calls != 1 {
		t.Fatalf("repository GetById calls = %d, want 1", calls)
	}
}

func TestGetByFilters_RejectsInvalidFiltersBeforeCallingRepository(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		filters model.TaskFilters
		wantErr error
	}{
		{
			name:    "missing user ID",
			filters: model.TaskFilters{},
			wantErr: model.ErrUserIDFilterRequired,
		},
		{
			name: "zero user ID",
			filters: model.TaskFilters{
				UserID: userIDPointer(0),
			},
			wantErr: model.ErrUserIDFilterInvalid,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			getByFiltersCalls := 0
			repo := &taskRepositoryStub{
				getByFiltersFunc: func(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
					getByFiltersCalls++
					return nil, nil
				},
			}

			tasks, err := service.NewTaskService(repo).GetByFilters(context.Background(), testCase.filters)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("GetByFilters error = %v, want %v", err, testCase.wantErr)
			}
			if tasks != nil {
				t.Fatalf("GetByFilters tasks = %+v, want nil", tasks)
			}
			if getByFiltersCalls != 0 {
				t.Fatalf("repository GetByFilters calls = %d, want 0", getByFiltersCalls)
			}
		})
	}
}

func TestGetByFilters_DelegatesFiltersAndReturnsRepositoryResult(t *testing.T) {
	ctx := context.Background()
	filters := model.TaskFilters{UserID: userIDPointer(42)}
	wantTasks := []*model.Task{{Id: 18, UserId: 42}, {Id: 19, UserId: 42}}
	wantErr := errors.New("list failed")
	calls := 0
	repo := &taskRepositoryStub{
		getByFiltersFunc: func(gotCtx context.Context, gotFilters model.TaskFilters) ([]*model.Task, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("GetByFilters did not preserve context")
			}
			if gotFilters != filters {
				t.Fatalf("GetByFilters filters = %+v, want %+v", gotFilters, filters)
			}
			return wantTasks, wantErr
		},
	}

	gotTasks, err := service.NewTaskService(repo).GetByFilters(ctx, filters)

	if !errors.Is(err, wantErr) {
		t.Fatalf("GetByFilters error = %v, want %v", err, wantErr)
	}
	if len(gotTasks) != len(wantTasks) || gotTasks[0] != wantTasks[0] || gotTasks[1] != wantTasks[1] {
		t.Fatalf("GetByFilters tasks = %+v, want %+v", gotTasks, wantTasks)
	}
	if calls != 1 {
		t.Fatalf("repository GetByFilters calls = %d, want 1", calls)
	}
}

func userIDPointer(id int64) *int64 {
	return &id
}

func TestUpdate_DelegatesToRepository(t *testing.T) {
	ctx := context.Background()
	task := &model.Task{Id: 19}
	wantErr := errors.New("update failed")
	calls := 0
	repo := &taskRepositoryStub{
		updateFunc: func(gotCtx context.Context, gotTask *model.Task) error {
			calls++
			if gotCtx != ctx || gotTask != task {
				t.Fatalf("Update did not preserve its arguments")
			}
			return wantErr
		},
	}

	err := service.NewTaskService(repo).Update(ctx, task)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Update error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Fatalf("repository Update calls = %d, want 1", calls)
	}
}

func TestDelete_DelegatesToRepository(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("delete failed")
	calls := 0
	repo := &taskRepositoryStub{
		deleteFunc: func(gotCtx context.Context, id int64) error {
			calls++
			if gotCtx != ctx || id != 20 {
				t.Fatalf("Delete arguments = (%v, %d), want (%v, 20)", gotCtx, id, ctx)
			}
			return wantErr
		},
	}

	err := service.NewTaskService(repo).Delete(ctx, 20)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Delete error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Fatalf("repository Delete calls = %d, want 1", calls)
	}
}
