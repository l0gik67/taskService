package service

import (
	"core/model"
	"database-service/repository"
)

type TaskService interface {
	GetById(id int64) (*model.Task, error)
	GetByUserId(userId int64) ([]*model.Task, error)
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

func (service *TaskServiceImpl) GetByUserId(userId int64) ([]*model.Task, error) {
	return service.repo.GetByUserId(userId)
}

func (service *TaskServiceImpl) Create(task *model.Task) (*model.Task, error) {
	if err := task.Validate(); err != nil {
		return nil, err
	}

	return service.repo.Create(task)
}

func (service *TaskServiceImpl) Update(task *model.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	return service.repo.Update(task)
}

func (service *TaskServiceImpl) Delete(taskId int64) error {
	return service.repo.Delete(taskId)
}
