package repository

import "core/model"

type TaskRepository interface {
	GetById(id int64) (*model.Task, error)
	GetByUserId(userId int64, boardId int64) ([]*model.Task, error)
	Create(task *model.Task) (*model.Task, error)
	Update(task *model.Task) error
	Delete(taskId int64) error
}
