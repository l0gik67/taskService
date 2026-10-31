package model

import "time"

type Task struct {
	Id        int
	Title     string
	UserId    int
	Duration  time.Duration
	StartTime time.Time
	Priority  int // 1 - important; 100 - zero important
}
