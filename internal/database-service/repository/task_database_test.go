package repository

import (
	"core/model"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMapTaskLookupError(t *testing.T) {
	if err := mapTaskLookupError(pgx.ErrNoRows); !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("error = %v, want %v", err, model.ErrTaskNotFound)
	}

	wantErr := errors.New("database unavailable")
	if err := mapTaskLookupError(wantErr); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
