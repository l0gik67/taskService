package transport_test

import (
	"context"
	"core/model"
	taskv1 "database-service/api/task/v1"
	"database-service/service"
	"database-service/transport"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type getByIDTaskServiceStub struct {
	service.TaskService
	err error
}

func (stub *getByIDTaskServiceStub) GetById(int64) (*model.Task, error) {
	return nil, stub.err
}

func TestGetTaskByIdMapsNotFoundError(t *testing.T) {
	handler := transport.NewTaskHandler(&getByIDTaskServiceStub{err: model.ErrTaskNotFound})

	_, err := handler.GetTaskById(context.Background(), &taskv1.GetTaskByIdRequest{Id: 17})

	if status.Code(err) != codes.NotFound {
		t.Fatalf("status code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestGetTaskByIdMapsUnexpectedErrorToInternal(t *testing.T) {
	handler := transport.NewTaskHandler(&getByIDTaskServiceStub{err: errors.New("database unavailable")})

	_, err := handler.GetTaskById(context.Background(), &taskv1.GetTaskByIdRequest{Id: 17})

	if status.Code(err) != codes.Internal {
		t.Fatalf("status code = %v, want %v", status.Code(err), codes.Internal)
	}
}
