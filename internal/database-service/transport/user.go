package transport

import (
	"context"
	"core/model"
	userv1 "database-service/api/user/v1"
	"database-service/service"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (handler *UserHandler) CreateUser(_ context.Context, request *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create user request is nil")
	}
	userModel := &model.User{Name: request.Name}
	if err := userModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := handler.userService.Create(userModel)
	if err != nil {
		return nil, userStatusError(err)
	}
	return &userv1.CreateUserResponse{User: userToProto(user)}, nil
}

func (handler *UserHandler) GetUserById(_ context.Context, request *userv1.GetUserByIdRequest) (*userv1.User, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get user request is nil")
	}
	if request.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user id must be greater than zero")
	}
	user, err := handler.userService.GetById(request.Id)
	if err != nil {
		return nil, userStatusError(err)
	}
	return userToProto(user), nil
}

func (handler *UserHandler) GetUsers(_ context.Context, request *userv1.GetUsersRequest) (*userv1.GetUsersResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get users request is nil")
	}
	users, err := handler.userService.GetAll()
	if err != nil {
		return nil, userStatusError(err)
	}
	response := &userv1.GetUsersResponse{Users: make([]*userv1.User, len(users))}
	for index, user := range users {
		response.Users[index] = userToProto(user)
	}
	return response, nil
}

func (handler *UserHandler) UpdateUser(_ context.Context, request *userv1.UpdateUserRequest) (*userv1.User, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "update user request is nil")
	}
	userModel := &model.User{Id: request.Id, Name: request.Name}
	if userModel.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user id must be greater than zero")
	}
	if err := userModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := handler.userService.Update(userModel)
	if err != nil {
		return nil, userStatusError(err)
	}
	return userToProto(user), nil
}

func (handler *UserHandler) DeleteUser(_ context.Context, request *userv1.DeleteUserRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "delete user request is nil")
	}
	if request.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user id must be greater than zero")
	}
	if err := handler.userService.Delete(request.Id); err != nil {
		return nil, userStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

func userToProto(user *model.User) *userv1.User {
	return &userv1.User{Id: user.Id, Name: user.Name}
}

func userStatusError(err error) error {
	switch {
	case errors.Is(err, model.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrUserInUse):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
