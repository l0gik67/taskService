package model

import (
	"errors"
	"strings"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserInUse    = errors.New("user is referenced by boards or tasks")
)

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (user *User) Validate() error {
	if user == nil {
		return errors.New("user can't be nil")
	}
	if strings.TrimSpace(user.Name) == "" {
		return errors.New("user name is empty")
	}
	return nil
}
