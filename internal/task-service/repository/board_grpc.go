package repository

import (
	"context"
	"core/model"
	boardv1 "database-service/api/board/v1"
	"errors"
)

type GRPCBoardRepository struct {
	client boardv1.BoardServiceClient
}

var _ BoardRepository = (*GRPCBoardRepository)(nil)

func NewGRPCBoardRepository(client boardv1.BoardServiceClient) *GRPCBoardRepository {
	return &GRPCBoardRepository{client: client}
}

func (repository *GRPCBoardRepository) Create(ctx context.Context, board *model.Board) (*model.Board, error) {
	response, err := repository.client.CreateBoard(ctx, &boardv1.CreateBoardRequest{
		OwnerUserId: board.OwnerUserId,
		Title:       board.Title,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil create board response")
	}
	return boardFromProto(response.Board)
}

func (repository *GRPCBoardRepository) GetById(ctx context.Context, id int64) (*model.Board, error) {
	board, err := repository.client.GetBoardById(ctx, &boardv1.GetBoardByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return boardFromProto(board)
}

func (repository *GRPCBoardRepository) GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error) {
	response, err := repository.client.GetBoardsByUserId(ctx, &boardv1.GetBoardsByUserIdRequest{UserId: userId})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("database-service returned a nil board list response")
	}
	boards := make([]*model.Board, len(response.Boards))
	for index, board := range response.Boards {
		boards[index], err = boardFromProto(board)
		if err != nil {
			return nil, err
		}
	}
	return boards, nil
}

func (repository *GRPCBoardRepository) CreateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	createdColumn, err := repository.client.CreateColumn(ctx, &boardv1.CreateColumnRequest{
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	})
	if err != nil {
		return nil, err
	}
	return boardColumnFromProto(createdColumn)
}

func (repository *GRPCBoardRepository) UpdateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	updatedColumn, err := repository.client.UpdateColumn(ctx, &boardv1.UpdateColumnRequest{
		Id:       column.Id,
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	})
	if err != nil {
		return nil, err
	}
	return boardColumnFromProto(updatedColumn)
}

func (repository *GRPCBoardRepository) DeleteColumn(ctx context.Context, id int64, targetColumnId int64) error {
	_, err := repository.client.DeleteColumn(ctx, &boardv1.DeleteColumnRequest{Id: id, TargetColumnId: targetColumnId})
	return err
}

func boardFromProto(board *boardv1.Board) (*model.Board, error) {
	if board == nil {
		return nil, errors.New("database-service returned a nil board")
	}
	result := &model.Board{
		Id:           board.Id,
		OwnerUserId:  board.OwnerUserId,
		Title:        board.Title,
		DoneColumnId: board.DoneColumnId,
		Columns:      make([]*model.BoardColumn, len(board.Columns)),
	}
	var err error
	for index, column := range board.Columns {
		result.Columns[index], err = boardColumnFromProto(column)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func boardColumnFromProto(column *boardv1.BoardColumn) (*model.BoardColumn, error) {
	if column == nil {
		return nil, errors.New("database-service returned a nil board column")
	}
	return &model.BoardColumn{
		Id:       column.Id,
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	}, nil
}
