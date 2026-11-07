package repository_test

import (
	"context"
	"core/model"
	userv1 "database-service/api/user/v1"
	"task-service/repository"
	"testing"

	"google.golang.org/grpc"
)

type userServiceClientStub struct {
	userv1.UserServiceClient
	createUserFunc func(context.Context, *userv1.CreateUserRequest, ...grpc.CallOption) (*userv1.CreateUserResponse, error)
	getUsersFunc   func(context.Context, *userv1.GetUsersRequest, ...grpc.CallOption) (*userv1.GetUsersResponse, error)
}

func (stub *userServiceClientStub) CreateUser(ctx context.Context, request *userv1.CreateUserRequest, options ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	return stub.createUserFunc(ctx, request, options...)
}

func (stub *userServiceClientStub) GetUsers(ctx context.Context, request *userv1.GetUsersRequest, options ...grpc.CallOption) (*userv1.GetUsersResponse, error) {
	return stub.getUsersFunc(ctx, request, options...)
}

func TestGRPCUserRepositoryCreateMapsRequestAndResponse(t *testing.T) {
	ctx := context.Background()
	client := &userServiceClientStub{
		createUserFunc: func(gotContext context.Context, request *userv1.CreateUserRequest, _ ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			if gotContext != ctx {
				t.Fatal("CreateUser did not preserve context")
			}
			if request.Name != "Vadim" {
				t.Fatalf("request = %+v", request)
			}
			return &userv1.CreateUserResponse{User: &userv1.User{Id: 7, Name: "Vadim"}}, nil
		},
	}

	user, err := repository.NewGRPCUserRepository(client).Create(ctx, &model.User{Name: "Vadim"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.Id != 7 || user.Name != "Vadim" {
		t.Fatalf("user = %+v", user)
	}
}

func TestGRPCUserRepositoryGetAllMapsResponse(t *testing.T) {
	client := &userServiceClientStub{
		getUsersFunc: func(context.Context, *userv1.GetUsersRequest, ...grpc.CallOption) (*userv1.GetUsersResponse, error) {
			return &userv1.GetUsersResponse{Users: []*userv1.User{
				{Id: 1, Name: "First"},
				{Id: 2, Name: "Second"},
			}}, nil
		},
	}

	users, err := repository.NewGRPCUserRepository(client).GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(users) != 2 || users[1].Name != "Second" {
		t.Fatalf("users = %+v", users)
	}
}
