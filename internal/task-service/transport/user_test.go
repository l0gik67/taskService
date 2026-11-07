package transport

import (
	"context"
	"core/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-service/service"
	"testing"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type httpUserServiceStub struct {
	service.UserService
	createFunc func(context.Context, *model.User) (*model.User, error)
	deleteFunc func(context.Context, int64) error
}

func (stub *httpUserServiceStub) Create(ctx context.Context, user *model.User) (*model.User, error) {
	return stub.createFunc(ctx, user)
}

func (stub *httpUserServiceStub) Delete(ctx context.Context, id int64) error {
	return stub.deleteFunc(ctx, id)
}

func TestUserHandlerCreate(t *testing.T) {
	stub := &httpUserServiceStub{
		createFunc: func(_ context.Context, user *model.User) (*model.User, error) {
			user.Id = 7
			return user, nil
		},
	}
	router := mux.NewRouter()
	RegisterUserRoutes(router, NewUserHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Vadim"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"id":7`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestUserHandlerDeleteReferencedUserReturnsConflict(t *testing.T) {
	stub := &httpUserServiceStub{
		deleteFunc: func(context.Context, int64) error {
			return status.Error(codes.FailedPrecondition, "user is referenced")
		},
	}
	router := mux.NewRouter()
	RegisterUserRoutes(router, NewUserHandler(stub))
	request := httptest.NewRequest(http.MethodDelete, "/users/7", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
