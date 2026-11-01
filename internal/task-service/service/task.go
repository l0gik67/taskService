package service

import "core/model"

type TaskService interface {
	GetById(id int64) (*model.Task, error)
	GetByUserId(userId int64) (*model.Task, error)
	Create(task *model.Task) (*model.Task, error)
	Update(task *model.Task) error
	Delete(id int64) error
}
