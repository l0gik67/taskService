package model

import (
	"errors"
	"strings"
)

var (
	ErrBoardNotFound              = errors.New("board not found")
	ErrBoardColumnNotFound        = errors.New("board column not found")
	ErrDoneColumnCannotBeDeleted  = errors.New("done column cannot be deleted")
	ErrTargetColumnRequired       = errors.New("target column is required for a non-empty column")
	ErrTargetColumnBelongsToBoard = errors.New("target column must belong to the same board")
)

type Board struct {
	Id           int64          `json:"id"`
	OwnerUserId  int64          `json:"owner_user_id"`
	Title        string         `json:"title"`
	DoneColumnId int64          `json:"done_column_id"`
	Columns      []*BoardColumn `json:"columns"`
}

func (board *Board) Validate() error {
	if board == nil {
		return errors.New("board can't be nil")
	}
	if board.OwnerUserId <= 0 {
		return errors.New("board owner_user_id must be greater than zero")
	}
	if strings.TrimSpace(board.Title) == "" {
		return errors.New("board title is empty")
	}
	return nil
}

type BoardColumn struct {
	Id       int64  `json:"id"`
	BoardId  int64  `json:"board_id"`
	Title    string `json:"title"`
	Position int32  `json:"position"`
}

func (column *BoardColumn) Validate() error {
	if column == nil {
		return errors.New("board column can't be nil")
	}
	if column.BoardId <= 0 {
		return errors.New("board_id must be greater than zero")
	}
	if strings.TrimSpace(column.Title) == "" {
		return errors.New("board column title is empty")
	}
	if column.Position < 0 {
		return errors.New("board column position can't be less than zero")
	}
	return nil
}
