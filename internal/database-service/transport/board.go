package transport

import (
	"context"
	"core/model"
	boardv1 "database-service/api/board/v1"
	"database-service/service"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BoardHandler struct {
	boardv1.UnimplementedBoardServiceServer
	boardService service.BoardService
}

func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{boardService: boardService}
}

func (handler *BoardHandler) CreateBoard(_ context.Context, request *boardv1.CreateBoardRequest) (*boardv1.CreateBoardResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create board request is nil")
	}
	boardModel := &model.Board{OwnerUserId: request.OwnerUserId, Title: request.Title}
	if err := boardModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	board, err := handler.boardService.Create(boardModel)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return &boardv1.CreateBoardResponse{Board: boardToProto(board)}, nil
}

func (handler *BoardHandler) GetBoardById(_ context.Context, request *boardv1.GetBoardByIdRequest) (*boardv1.Board, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get board request is nil")
	}
	if request.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board id must be greater than zero")
	}
	board, err := handler.boardService.GetById(request.Id)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return boardToProto(board), nil
}

func (handler *BoardHandler) GetBoardsByUserId(_ context.Context, request *boardv1.GetBoardsByUserIdRequest) (*boardv1.GetBoardsByUserIdResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "get boards by user request is nil")
	}
	if request.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user id must be greater than zero")
	}
	boards, err := handler.boardService.GetByUserId(request.UserId)
	if err != nil {
		return nil, boardStatusError(err)
	}
	response := &boardv1.GetBoardsByUserIdResponse{Boards: make([]*boardv1.Board, len(boards))}
	for index, board := range boards {
		response.Boards[index] = boardToProto(board)
	}
	return response, nil
}

func (handler *BoardHandler) CreateColumn(_ context.Context, request *boardv1.CreateColumnRequest) (*boardv1.BoardColumn, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create column request is nil")
	}
	columnModel := &model.BoardColumn{
		BoardId:  request.BoardId,
		Title:    request.Title,
		Position: request.Position,
	}
	if err := columnModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	column, err := handler.boardService.CreateColumn(columnModel)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return boardColumnToProto(column), nil
}

func (handler *BoardHandler) UpdateColumn(_ context.Context, request *boardv1.UpdateColumnRequest) (*boardv1.BoardColumn, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "update column request is nil")
	}
	columnModel := &model.BoardColumn{
		Id:       request.Id,
		BoardId:  request.BoardId,
		Title:    request.Title,
		Position: request.Position,
	}
	if columnModel.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board column id must be greater than zero")
	}
	if err := columnModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	column, err := handler.boardService.UpdateColumn(columnModel)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return boardColumnToProto(column), nil
}

func (handler *BoardHandler) DeleteColumn(_ context.Context, request *boardv1.DeleteColumnRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "delete column request is nil")
	}
	if request.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board column id must be greater than zero")
	}
	if err := handler.boardService.DeleteColumn(request.Id, request.TargetColumnId); err != nil {
		return nil, boardStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

func boardToProto(board *model.Board) *boardv1.Board {
	result := &boardv1.Board{
		Id:           board.Id,
		OwnerUserId:  board.OwnerUserId,
		Title:        board.Title,
		DoneColumnId: board.DoneColumnId,
		Columns:      make([]*boardv1.BoardColumn, len(board.Columns)),
	}
	for index, column := range board.Columns {
		result.Columns[index] = boardColumnToProto(column)
	}
	return result
}

func boardColumnToProto(column *model.BoardColumn) *boardv1.BoardColumn {
	return &boardv1.BoardColumn{
		Id:       column.Id,
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	}
}

func boardStatusError(err error) error {
	switch {
	case errors.Is(err, model.ErrBoardNotFound), errors.Is(err, model.ErrBoardColumnNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrDoneColumnCannotBeDeleted), errors.Is(err, model.ErrTargetColumnRequired), errors.Is(err, model.ErrTargetColumnBelongsToBoard):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
