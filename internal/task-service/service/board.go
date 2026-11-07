package service

import (
	"context"
	"core/model"
	"errors"
	"strings"
	"task-service/repository"
)

type BoardService interface {
	Create(ctx context.Context, board *model.Board) (*model.Board, error)
	GetById(ctx context.Context, id int64) (*model.Board, error)
	GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error)
	CreateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	UpdateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	DeleteColumn(ctx context.Context, id int64, targetColumnId int64) error
}

type BoardServiceImpl struct {
	repository repository.BoardRepository
}

func NewBoardService(repository repository.BoardRepository) *BoardServiceImpl {
	return &BoardServiceImpl{repository: repository}
}

func (service *BoardServiceImpl) Create(ctx context.Context, board *model.Board) (*model.Board, error) {
	if err := board.Validate(); err != nil {
		return nil, err
	}
	board.Title = strings.TrimSpace(board.Title)
	return service.repository.Create(ctx, board)
}

func (service *BoardServiceImpl) GetById(ctx context.Context, id int64) (*model.Board, error) {
	if id <= 0 {
		return nil, errors.New("board id must be greater than zero")
	}
	return service.repository.GetById(ctx, id)
}

func (service *BoardServiceImpl) GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error) {
	if userId <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetByUserId(ctx, userId)
}

func (service *BoardServiceImpl) CreateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.CreateColumn(ctx, column)
}

func (service *BoardServiceImpl) UpdateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	if column.Id <= 0 {
		return nil, errors.New("board column id must be greater than zero")
	}
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.UpdateColumn(ctx, column)
}

func (service *BoardServiceImpl) DeleteColumn(ctx context.Context, id int64, targetColumnId int64) error {
	if id <= 0 {
		return errors.New("board column id must be greater than zero")
	}
	return service.repository.DeleteColumn(ctx, id, targetColumnId)
}
