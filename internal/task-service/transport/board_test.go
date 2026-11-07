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

type boardServiceStub struct {
	service.BoardService
	createFunc       func(context.Context, *model.Board) (*model.Board, error)
	deleteColumnFunc func(context.Context, int64, int64) error
}

func (stub *boardServiceStub) Create(ctx context.Context, board *model.Board) (*model.Board, error) {
	return stub.createFunc(ctx, board)
}

func (stub *boardServiceStub) DeleteColumn(ctx context.Context, id int64, targetColumnId int64) error {
	return stub.deleteColumnFunc(ctx, id, targetColumnId)
}

func TestBoardHandlerCreate(t *testing.T) {
	stub := &boardServiceStub{
		createFunc: func(_ context.Context, board *model.Board) (*model.Board, error) {
			if board.OwnerUserId != 42 || board.Title != "Development" {
				t.Fatalf("board = %+v", board)
			}
			board.Id = 7
			board.DoneColumnId = 11
			return board, nil
		},
	}
	router := mux.NewRouter()
	RegisterBoardRoutes(router, NewBoardHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/boards", strings.NewReader(`{"owner_user_id":42,"title":"Development"}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"done_column_id":11`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestBoardHandlerDeleteDoneColumnReturnsConflict(t *testing.T) {
	stub := &boardServiceStub{
		deleteColumnFunc: func(context.Context, int64, int64) error {
			return status.Error(codes.FailedPrecondition, "done column cannot be deleted")
		},
	}
	router := mux.NewRouter()
	RegisterBoardRoutes(router, NewBoardHandler(stub))
	request := httptest.NewRequest(http.MethodDelete, "/boards/7/columns/11", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
