package repository

import (
	"context"
	"core/model"
	userv1 "database-service/api/user/v1"
	"errors"
	"fmt"
)

type GRPCUserRepository struct {
	client userv1.UserServiceClient
}

var _ UserRepository = (*GRPCUserRepository)(nil)

func NewGRPCUserRepository(client userv1.UserServiceClient) *GRPCUserRepository {
	return &GRPCUserRepository{client: client}
}

func (repository *GRPCUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	response, err := repository.client.CreateUser(ctx, &userv1.CreateUserRequest{Name: user.Name})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil create user response")
	}
	return userFromProto(response.User)
}

func (repository *GRPCUserRepository) GetById(ctx context.Context, id int64) (*model.User, error) {
	user, err := repository.client.GetUserById(ctx, &userv1.GetUserByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return userFromProto(user)
}

func (repository *GRPCUserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	response, err := repository.client.GetUsers(ctx, &userv1.GetUsersRequest{})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil user list response")
	}
	users := make([]*model.User, len(response.Users))
	for index, user := range response.Users {
		mappedUser, err := userFromProto(user)
		if err != nil {
			return nil, fmt.Errorf("map user at index %d: %w", index, err)
		}
		users[index] = mappedUser
	}
	return users, nil
}

func (repository *GRPCUserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	updatedUser, err := repository.client.UpdateUser(ctx, &userv1.UpdateUserRequest{Id: user.Id, Name: user.Name})
	if err != nil {
		return nil, err
	}
	return userFromProto(updatedUser)
}

func (repository *GRPCUserRepository) Delete(ctx context.Context, id int64) error {
	_, err := repository.client.DeleteUser(ctx, &userv1.DeleteUserRequest{Id: id})
	return err
}

func userFromProto(user *userv1.User) (*model.User, error) {
	if user == nil {
		return nil, errors.New("database-service returned a nil user")
	}
	return &model.User{Id: user.Id, Name: user.Name}, nil
}
