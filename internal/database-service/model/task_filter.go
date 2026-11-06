package model

import "errors"

type TaskFilter struct {
	UserID int64
}

func (filter TaskFilter) Validate() error {
	if filter.UserID <= 0 {
		return errors.New("userId must be greater than zero")
	}
	return nil
}
