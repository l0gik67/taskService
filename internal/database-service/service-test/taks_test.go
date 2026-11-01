package service_test

import (
	"database-service/model"
	"database-service/service"
	"testing"
)

type fakeTaskRepository struct {
	createCalled bool
	updateCalled bool
}

func (r *fakeTaskRepository) Create(task *model.Task) (*model.Task, error) {
	r.createCalled = true
	return task, nil
}

func (r *fakeTaskRepository) GetById(id int64) (*model.Task, error) {
	panic("unexpected GetById call")
}

func (r *fakeTaskRepository) GetByUserId(userId int64) ([]*model.Task, error) {
	panic("unexpected GetByUserId call")
}

func (r *fakeTaskRepository) Update(task *model.Task) error {
	r.updateCalled = true
	return nil
}

func (r *fakeTaskRepository) Delete(id int64) error {
	panic("unexpected Delete call")
}

func TestCreate_NormalCase(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "First Task",
		UserId:   1,
		Priority: 1,
	}

	_, err := svc.Create(task)
	if err != nil {
		t.Fatal("unexpected error for empty title")
	}
	if !repo.createCalled {
		t.Fatal("expected Create to call repository")
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "   ",
		UserId:   1,
		Priority: 1,
	}

	_, err := svc.Create(task)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if repo.createCalled {
		t.Fatal("expected Create to not call repository")
	}
}

func TestCreate_PriorityLessThen1(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "title",
		UserId:   1,
		Priority: -1,
	}

	_, err := svc.Create(task)
	if err == nil {
		t.Fatal("expected error for priority less than 1")
	}
	if repo.createCalled {
		t.Fatal("expected Create to not call repository")
	}
}

func TestCreate_PriorityMoreThen100(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "title",
		UserId:   1,
		Priority: 101,
	}

	_, err := svc.Create(task)
	if err == nil {
		t.Fatal("expected error for priority greater than 100")
	}
	if repo.createCalled {
		t.Fatal("expected Create to not call repository")
	}
}

func TestUpdate_NormalCase(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "First Task",
		UserId:   1,
		Priority: 1,
	}

	err := svc.Update(task)
	if err != nil {
		t.Fatal("unexpected error for empty title")
	}
	if !repo.updateCalled {
		t.Fatal("expected Update to call repository")
	}
}

func TestUpdate_EmptyTitle(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "",
		UserId:   1,
		Priority: 1,
	}

	_, err := svc.Create(task)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if repo.createCalled {
		t.Fatal("expected Create to not call repository")
	}
}

func TestUpdate_PriorityLessThen1(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "title",
		UserId:   1,
		Priority: -1,
	}

	err := svc.Update(task)
	if err == nil {
		t.Fatal("expected error for priority less than 1")
	}
	if repo.updateCalled {
		t.Fatal("expected Update to not call repository")
	}
}

func TestUpdate_PriorityMoreThen100(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := service.NewTaskServiceImpl(repo)
	task := &model.Task{
		Title:    "title",
		UserId:   1,
		Priority: 101,
	}

	err := svc.Update(task)
	if err == nil {
		t.Fatal("expected error for priority greater than 100")
	}
	if repo.updateCalled {
		t.Fatal("expected Update to not call repository")
	}
}
