package dto

import "time"

type TaskCreateDto struct {
	UserId      int64         `json:"user_id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	ColumnId    int64         `json:"column_id"`
	Position    int32         `json:"position"`
	Progress    int32         `json:"progress"`
	IsDone      bool          `json:"is_done"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	Priority    int32         `json:"priority"`
}

type TaskUpdateDto struct {
	Id          int64         `json:"id"`
	UserId      int64         `json:"user_id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	ColumnId    int64         `json:"column_id"`
	Position    int32         `json:"position"`
	Progress    int32         `json:"progress"`
	IsDone      bool          `json:"is_done"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	Priority    int32         `json:"priority"`
}
