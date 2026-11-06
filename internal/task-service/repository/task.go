package repository

import (
	"context"
	"core/model"
)

type TaskRepository interface {
	GetById(ctx context.Context, id int64) (*model.Task, error)
	GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error)
	Create(ctx context.Context, task *model.Task) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
}
