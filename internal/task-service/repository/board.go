package repository

import (
	"context"
	"core/model"
)

type BoardRepository interface {
	Create(ctx context.Context, board *model.Board) (*model.Board, error)
	GetById(ctx context.Context, id int64) (*model.Board, error)
	GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error)
	CreateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	UpdateColumn(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	DeleteColumn(ctx context.Context, id int64, targetColumnId int64) error
}
