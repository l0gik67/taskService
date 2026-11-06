package model

import (
	"errors"
	"strings"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Task struct {
	Id        int64         `json:"id"`
	Title     string        `json:"title"`
	UserId    int64         `json:"user_id"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	Priority  int32         `json:"priority"` // 1 - important; 100 - zero important
}

func (task *Task) Validate() error {
	if task == nil {
		return errors.New("task can't be nil")
	}
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("task title is empty")
	}
	if task.Priority < 1 || task.Priority > 100 {
		return errors.New("task priority should be between 0 and 100")
	}
	if task.Duration < 0 {
		return errors.New("task duration can't be less than 0")
	}
	return nil
}
