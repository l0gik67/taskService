package repository

import (
	"context"
	"core/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetById(ctx context.Context, id int64) (*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}
