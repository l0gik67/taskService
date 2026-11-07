package service

import (
	"core/model"
	"database-service/repository"
	"errors"
	"strings"
)

type BoardService interface {
	Create(board *model.Board) (*model.Board, error)
	GetById(id int64) (*model.Board, error)
	GetByUserId(userId int64) ([]*model.Board, error)
	CreateColumn(column *model.BoardColumn) (*model.BoardColumn, error)
	UpdateColumn(column *model.BoardColumn) (*model.BoardColumn, error)
	DeleteColumn(id int64, targetColumnId int64) error
}

type BoardServiceImpl struct {
	repository repository.BoardRepository
}

func NewBoardService(repository repository.BoardRepository) *BoardServiceImpl {
	return &BoardServiceImpl{repository: repository}
}

func (service *BoardServiceImpl) Create(board *model.Board) (*model.Board, error) {
	if err := board.Validate(); err != nil {
		return nil, err
	}
	board.Title = strings.TrimSpace(board.Title)
	return service.repository.Create(board)
}

func (service *BoardServiceImpl) GetById(id int64) (*model.Board, error) {
	if id <= 0 {
		return nil, errors.New("board id must be greater than zero")
	}
	return service.repository.GetById(id)
}

func (service *BoardServiceImpl) GetByUserId(userId int64) ([]*model.Board, error) {
	if userId <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetByUserId(userId)
}

func (service *BoardServiceImpl) CreateColumn(column *model.BoardColumn) (*model.BoardColumn, error) {
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.CreateColumn(column)
}

func (service *BoardServiceImpl) UpdateColumn(column *model.BoardColumn) (*model.BoardColumn, error) {
	if column.Id <= 0 {
		return nil, errors.New("board column id must be greater than zero")
	}
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.UpdateColumn(column)
}

func (service *BoardServiceImpl) DeleteColumn(id int64, targetColumnId int64) error {
	if id <= 0 {
		return errors.New("board column id must be greater than zero")
	}
	return service.repository.DeleteColumn(id, targetColumnId)
}
