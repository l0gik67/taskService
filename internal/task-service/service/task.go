package service

import (
	"context"
	"core/model"
	"task-service/repository"
)

type TaskService interface {
	GetById(ctx context.Context, id int64) (*model.Task, error)
	GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error)
	Create(ctx context.Context, task *model.Task) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
}

type TaskServiceImpl struct {
	repository repository.TaskRepository
}

func NewTaskService(repository repository.TaskRepository) *TaskServiceImpl {
	return &TaskServiceImpl{repository: repository}
}

func (service *TaskServiceImpl) GetById(ctx context.Context, id int64) (*model.Task, error) {
	return service.repository.GetById(ctx, id)
}

func (service *TaskServiceImpl) GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	return service.repository.GetByFilters(ctx, filters)
}

func (service *TaskServiceImpl) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	if err := task.Validate(); err != nil {
		return nil, err
	}

	return service.repository.Create(ctx, task)
}

func (service *TaskServiceImpl) Update(ctx context.Context, task *model.Task) error {
	return service.repository.Update(ctx, task)
}

func (service *TaskServiceImpl) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}
