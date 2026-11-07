package repository_test

import (
	"context"
	"core/model"
	boardv1 "database-service/api/board/v1"
	"errors"
	"task-service/repository"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type boardServiceClientStub struct {
	boardv1.BoardServiceClient
	createBoardFunc  func(context.Context, *boardv1.CreateBoardRequest, ...grpc.CallOption) (*boardv1.CreateBoardResponse, error)
	deleteColumnFunc func(context.Context, *boardv1.DeleteColumnRequest, ...grpc.CallOption) (*emptypb.Empty, error)
}

func (stub *boardServiceClientStub) CreateBoard(ctx context.Context, request *boardv1.CreateBoardRequest, options ...grpc.CallOption) (*boardv1.CreateBoardResponse, error) {
	return stub.createBoardFunc(ctx, request, options...)
}

func (stub *boardServiceClientStub) DeleteColumn(ctx context.Context, request *boardv1.DeleteColumnRequest, options ...grpc.CallOption) (*emptypb.Empty, error) {
	return stub.deleteColumnFunc(ctx, request, options...)
}

func TestGRPCBoardRepositoryCreateMapsRequestAndResponse(t *testing.T) {
	ctx := context.Background()
	client := &boardServiceClientStub{
		createBoardFunc: func(gotContext context.Context, request *boardv1.CreateBoardRequest, _ ...grpc.CallOption) (*boardv1.CreateBoardResponse, error) {
			if gotContext != ctx {
				t.Fatal("CreateBoard did not preserve context")
			}
			if request.OwnerUserId != 42 || request.Title != "Development" {
				t.Fatalf("request = %+v", request)
			}
			return &boardv1.CreateBoardResponse{Board: &boardv1.Board{
				Id:           7,
				OwnerUserId:  42,
				Title:        "Development",
				DoneColumnId: 11,
				Columns: []*boardv1.BoardColumn{
					{Id: 8, BoardId: 7, Title: "Backlog", Position: 0},
					{Id: 11, BoardId: 7, Title: "Готово", Position: 3},
				},
			}}, nil
		},
	}

	board, err := repository.NewGRPCBoardRepository(client).Create(ctx, &model.Board{OwnerUserId: 42, Title: "Development"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if board.Id != 7 || board.DoneColumnId != 11 || len(board.Columns) != 2 {
		t.Fatalf("board = %+v", board)
	}
}

func TestGRPCBoardRepositoryDeleteColumnPreservesError(t *testing.T) {
	wantErr := status.Error(codes.FailedPrecondition, "done column cannot be deleted")
	client := &boardServiceClientStub{
		deleteColumnFunc: func(_ context.Context, request *boardv1.DeleteColumnRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			if request.Id != 11 || request.TargetColumnId != 8 {
				t.Fatalf("request = %+v", request)
			}
			return nil, wantErr
		},
	}

	err := repository.NewGRPCBoardRepository(client).DeleteColumn(context.Background(), 11, 8)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
