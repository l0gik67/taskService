package service

import (
	"context"
	"core/model"
	"errors"
	"strings"
	"task-service/repository"
)

type UserService interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetById(ctx context.Context, id int64) (*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}

type UserServiceImpl struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repository: repository}
}

func (service *UserServiceImpl) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	user.Name = strings.TrimSpace(user.Name)
	return service.repository.Create(ctx, user)
}

func (service *UserServiceImpl) GetById(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetById(ctx, id)
}

func (service *UserServiceImpl) GetAll(ctx context.Context) ([]*model.User, error) {
	return service.repository.GetAll(ctx)
}

func (service *UserServiceImpl) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if user.Id <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	user.Name = strings.TrimSpace(user.Name)
	return service.repository.Update(ctx, user)
}

func (service *UserServiceImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("user id must be greater than zero")
	}
	return service.repository.Delete(ctx, id)
}
