package repository

import "core/model"

type BoardRepository interface {
	Create(board *model.Board) (*model.Board, error)
	GetById(id int64) (*model.Board, error)
	GetByUserId(userId int64) ([]*model.Board, error)
	CreateColumn(column *model.BoardColumn) (*model.BoardColumn, error)
	UpdateColumn(column *model.BoardColumn) (*model.BoardColumn, error)
	DeleteColumn(id int64, targetColumnId int64) error
}
