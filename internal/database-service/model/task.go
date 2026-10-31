package model

import "time"

type Task struct {
	Id        int64
	Title     string
	UserId    int64
	Duration  time.Duration
	StartTime time.Time
	Priority  int32 // 1 - important; 100 - zero important
}
