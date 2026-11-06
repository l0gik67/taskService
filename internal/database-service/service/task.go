package service

import (
	"database-service/model"
	"database-service/repository"
	"errors"
	"strings"
)

type TaskService interface {
	GetById(id int64) (*model.Task, error)
	List(filter model.TaskFilter) ([]*model.Task, error)
	Create(task *model.Task) (*model.Task, error)
	Update(task *model.Task) error
	Delete(taskId int64) error
}

type TaskServiceImpl struct {
	repo repository.TaskRepository
}

func NewTaskServiceImpl(repo repository.TaskRepository) *TaskServiceImpl {
	return &TaskServiceImpl{repo: repo}
}

func (service *TaskServiceImpl) GetById(id int64) (*model.Task, error) {
	return service.repo.GetById(id)
}

func (service *TaskServiceImpl) List(filter model.TaskFilter) ([]*model.Task, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return service.repo.List(filter)
}

func (service *TaskServiceImpl) Create(task *model.Task) (*model.Task, error) {
	if err := service.validate(task); err != nil {
		return nil, err
	}

	return service.repo.Create(task)
}

func (service *TaskServiceImpl) Update(task *model.Task) error {
	if err := service.validate(task); err != nil {
		return err
	}

	return service.repo.Update(task)
}

func (service *TaskServiceImpl) Delete(taskId int64) error {
	return service.repo.Delete(taskId)
}

func (service TaskServiceImpl) validate(task *model.Task) error {
	if task == nil {
		return errors.New("task can't be nil")
	}
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("task title is empty")
	}
	if task.Priority < 1 || task.Priority > 100 {
		return errors.New("task priority should be between 0 and 100")
	}
	if task.Duration < 0 {
		return errors.New("task duration can't be less than 0")
	}
	return nil
}
