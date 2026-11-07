package model

import "testing"

func TestBoardValidate(t *testing.T) {
	validBoard := &Board{OwnerUserId: 42, Title: "Development"}
	if err := validBoard.Validate(); err != nil {
		t.Fatalf("valid board returned error: %v", err)
	}

	for _, board := range []*Board{
		nil,
		{OwnerUserId: 0, Title: "Development"},
		{OwnerUserId: 42, Title: "   "},
	} {
		if err := board.Validate(); err == nil {
			t.Fatalf("board %+v must be invalid", board)
		}
	}
}

func TestBoardColumnValidate(t *testing.T) {
	validColumn := &BoardColumn{BoardId: 7, Title: "Review", Position: 2}
	if err := validColumn.Validate(); err != nil {
		t.Fatalf("valid column returned error: %v", err)
	}

	for _, column := range []*BoardColumn{
		nil,
		{BoardId: 0, Title: "Review"},
		{BoardId: 7, Title: "   "},
		{BoardId: 7, Title: "Review", Position: -1},
	} {
		if err := column.Validate(); err == nil {
			t.Fatalf("column %+v must be invalid", column)
		}
	}
}
