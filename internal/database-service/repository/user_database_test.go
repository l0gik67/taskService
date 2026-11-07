package repository

import (
	"core/model"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapUserLookupError(t *testing.T) {
	if err := mapUserLookupError(pgx.ErrNoRows); !errors.Is(err, model.ErrUserNotFound) {
		t.Fatalf("error = %v, want %v", err, model.ErrUserNotFound)
	}

	wantErr := errors.New("database unavailable")
	if err := mapUserLookupError(wantErr); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestMapUserDeleteError(t *testing.T) {
	for _, code := range []string{"23001", "23503"} {
		foreignKeyError := &pgconn.PgError{Code: code}
		if err := mapUserDeleteError(foreignKeyError); !errors.Is(err, model.ErrUserInUse) {
			t.Fatalf("code = %s, error = %v, want %v", code, err, model.ErrUserInUse)
		}
	}

	wantErr := errors.New("database unavailable")
	if err := mapUserDeleteError(wantErr); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
