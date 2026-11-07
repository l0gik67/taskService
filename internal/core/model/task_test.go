package model

import "testing"

func TestTaskValidateProgress(t *testing.T) {
	for _, progress := range []int32{-1, 101} {
		task := &Task{Title: "Task", Priority: 1, Progress: progress}
		if err := task.Validate(); err == nil {
			t.Fatalf("progress %d must be invalid", progress)
		}
	}
}
