package transport_test

import (
	"context"
	"core/model"
	userv1 "database-service/api/user/v1"
	"database-service/service"
	"database-service/transport"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userServiceStub struct {
	service.UserService
	createFunc func(*model.User) (*model.User, error)
	deleteFunc func(int64) error
}

func (stub *userServiceStub) Create(user *model.User) (*model.User, error) {
	return stub.createFunc(user)
}

func (stub *userServiceStub) Delete(id int64) error {
	return stub.deleteFunc(id)
}

func TestUserHandlerCreate(t *testing.T) {
	handler := transport.NewUserHandler(&userServiceStub{
		createFunc: func(user *model.User) (*model.User, error) {
			if user.Name != "Vadim" {
				t.Fatalf("user = %+v", user)
			}
			user.Id = 7
			return user, nil
		},
	})

	response, err := handler.CreateUser(context.Background(), &userv1.CreateUserRequest{Name: "Vadim"})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if response.User.Id != 7 || response.User.Name != "Vadim" {
		t.Fatalf("response = %+v", response)
	}
}

func TestUserHandlerDeleteMapsInUseError(t *testing.T) {
	handler := transport.NewUserHandler(&userServiceStub{
		deleteFunc: func(int64) error { return model.ErrUserInUse },
	})

	_, err := handler.DeleteUser(context.Background(), &userv1.DeleteUserRequest{Id: 7})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %v, want %v", status.Code(err), codes.FailedPrecondition)
	}
}
