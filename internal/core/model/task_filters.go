package model

import "errors"

var (
	ErrUserIDFilterRequired = errors.New("user_id filter is required")
	ErrUserIDFilterInvalid  = errors.New("user_id filter must be greater than zero")
	ErrBoardIDFilterInvalid = errors.New("board_id filter must be greater than zero")
)

type TaskFilters struct {
	UserID  *int64
	BoardID *int64
}

func (filters TaskFilters) Validate() error {
	if filters.UserID == nil {
		return ErrUserIDFilterRequired
	}
	if *filters.UserID <= 0 {
		return ErrUserIDFilterInvalid
	}
	if filters.BoardID != nil && *filters.BoardID <= 0 {
		return ErrBoardIDFilterInvalid
	}
	return nil
}
