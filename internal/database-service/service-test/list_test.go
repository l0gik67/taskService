package service_test

import (
	"database-service/model"
	"database-service/service"
	"errors"
	"reflect"
	"testing"
)

type fakeListRepository struct {
	fakeTaskRepository
	called bool
	filter model.TaskFilter
	tasks  []*model.Task
	err    error
}

func (repo *fakeListRepository) List(filter model.TaskFilter) ([]*model.Task, error) {
	repo.called = true
	repo.filter = filter
	return repo.tasks, repo.err
}

func TestList_RequiresPositiveUserID(t *testing.T) {
	for _, tc := range []struct {
		name   string
		userID int64
	}{
		{name: "missing userId", userID: 0},
		{name: "negative userId", userID: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeListRepository{}
			svc := service.NewTaskServiceImpl(repo)
			_, err := svc.List(model.TaskFilter{UserID: tc.userID})
			if err == nil {
				t.Fatal("expected an error for a nonpositive userId")
			}
			if repo.called {
				t.Fatal("invalid filter must not reach the repository")
			}
		})
	}
}

func TestList_PassesFilterAndReturnsTasks(t *testing.T) {
	want := []*model.Task{{Id: 1, UserId: 42, Title: "First task"}, {Id: 2, UserId: 42, Title: "Second task"}}
	repo := &fakeListRepository{tasks: want}
	svc := service.NewTaskServiceImpl(repo)
	filter := model.TaskFilter{UserID: 42}
	tasks, err := svc.List(filter)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !repo.called || repo.filter != filter {
		t.Fatalf("repository received filter %+v, want %+v", repo.filter, filter)
	}
	if !reflect.DeepEqual(tasks, want) {
		t.Fatalf("tasks = %+v, want %+v", tasks, want)
	}
}

func TestList_EmptyResultIsNotAnError(t *testing.T) {
	svc := service.NewTaskServiceImpl(&fakeListRepository{})
	tasks, err := svc.List(model.TaskFilter{UserID: 42})
	if err != nil || len(tasks) != 0 {
		t.Fatalf("expected an empty result without an error, got %v, %v", tasks, err)
	}
}

func TestList_ReturnsRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	svc := service.NewTaskServiceImpl(&fakeListRepository{err: want})
	_, err := svc.List(model.TaskFilter{UserID: 42})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
