package repository

import "core/model"

type UserRepository interface {
	Create(user *model.User) (*model.User, error)
	GetById(id int64) (*model.User, error)
	GetAll() ([]*model.User, error)
	Update(user *model.User) (*model.User, error)
	Delete(id int64) error
}
