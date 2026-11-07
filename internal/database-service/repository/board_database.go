package repository

import (
	"context"
	"core/model"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var defaultBoardColumns = []model.BoardColumn{
	{Title: "Backlog", Position: 0},
	{Title: "В работе", Position: 1},
	{Title: "Проверка", Position: 2},
	{Title: "Готово", Position: 3},
}

type BoardRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBoardRepository(pool *pgxpool.Pool) BoardRepository {
	return &BoardRepositoryImpl{pool: pool}
}

func (repository *BoardRepositoryImpl) Create(board *model.Board) (*model.Board, error) {
	ctx := context.Background()
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		"INSERT INTO boards (owner_user_id, title) VALUES ($1, $2) RETURNING id",
		board.OwnerUserId,
		board.Title,
	).Scan(&board.Id)
	if err != nil {
		return nil, err
	}

	board.Columns = make([]*model.BoardColumn, 0, len(defaultBoardColumns))
	for _, defaultColumn := range defaultBoardColumns {
		column := &model.BoardColumn{
			BoardId:  board.Id,
			Title:    defaultColumn.Title,
			Position: defaultColumn.Position,
		}
		err = tx.QueryRow(ctx,
			"INSERT INTO board_columns (board_id, title, position) VALUES ($1, $2, $3) RETURNING id",
			column.BoardId,
			column.Title,
			column.Position,
		).Scan(&column.Id)
		if err != nil {
			return nil, err
		}
		board.Columns = append(board.Columns, column)
		if column.Position == 3 {
			board.DoneColumnId = column.Id
		}
	}

	_, err = tx.Exec(ctx, "UPDATE boards SET done_column_id = $2 WHERE id = $1", board.Id, board.DoneColumnId)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return board, nil
}

func (repository *BoardRepositoryImpl) GetById(id int64) (*model.Board, error) {
	ctx := context.Background()
	connection, err := repository.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer connection.Release()

	board := &model.Board{}
	err = connection.QueryRow(ctx,
		"SELECT id, owner_user_id, title, done_column_id FROM boards WHERE id = $1",
		id,
	).Scan(&board.Id, &board.OwnerUserId, &board.Title, &board.DoneColumnId)
	if err != nil {
		return nil, mapBoardLookupError(err)
	}
	board.Columns, err = getBoardColumns(ctx, connection, board.Id)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (repository *BoardRepositoryImpl) GetByUserId(userId int64) ([]*model.Board, error) {
	ctx := context.Background()
	connection, err := repository.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer connection.Release()

	rows, err := connection.Query(ctx,
		"SELECT id, owner_user_id, title, done_column_id FROM boards WHERE owner_user_id = $1 ORDER BY id",
		userId,
	)
	if err != nil {
		return nil, err
	}
	boards := make([]*model.Board, 0)
	for rows.Next() {
		board := &model.Board{}
		if err := rows.Scan(&board.Id, &board.OwnerUserId, &board.Title, &board.DoneColumnId); err != nil {
			rows.Close()
			return nil, err
		}
		boards = append(boards, board)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, board := range boards {
		board.Columns, err = getBoardColumns(ctx, connection, board.Id)
		if err != nil {
			return nil, err
		}
	}
	return boards, nil
}

func (repository *BoardRepositoryImpl) CreateColumn(column *model.BoardColumn) (*model.BoardColumn, error) {
	ctx := context.Background()
	err := repository.pool.QueryRow(ctx,
		"INSERT INTO board_columns (board_id, title, position) VALUES ($1, $2, $3) RETURNING id",
		column.BoardId,
		column.Title,
		column.Position,
	).Scan(&column.Id)
	if err != nil {
		return nil, err
	}
	return column, nil
}

func (repository *BoardRepositoryImpl) UpdateColumn(column *model.BoardColumn) (*model.BoardColumn, error) {
	ctx := context.Background()
	err := repository.pool.QueryRow(ctx,
		"UPDATE board_columns SET title = $2, position = $3 WHERE id = $1 AND board_id = $4 RETURNING id",
		column.Id,
		column.Title,
		column.Position,
		column.BoardId,
	).Scan(&column.Id)
	if err != nil {
		return nil, mapBoardColumnLookupError(err)
	}
	return column, nil
}

func (repository *BoardRepositoryImpl) DeleteColumn(id int64, targetColumnId int64) error {
	ctx := context.Background()
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var boardId int64
	var doneColumnId int64
	err = tx.QueryRow(ctx,
		"SELECT board_columns.board_id, boards.done_column_id FROM board_columns JOIN boards ON boards.id = board_columns.board_id WHERE board_columns.id = $1",
		id,
	).Scan(&boardId, &doneColumnId)
	if err != nil {
		return mapBoardColumnLookupError(err)
	}
	if id == doneColumnId {
		return model.ErrDoneColumnCannotBeDeleted
	}

	var taskCount int64
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE column_id = $1", id).Scan(&taskCount); err != nil {
		return err
	}
	if taskCount > 0 {
		if targetColumnId <= 0 {
			return model.ErrTargetColumnRequired
		}
		var targetBoardId int64
		if err := tx.QueryRow(ctx, "SELECT board_id FROM board_columns WHERE id = $1", targetColumnId).Scan(&targetBoardId); err != nil {
			return mapBoardColumnLookupError(err)
		}
		if targetBoardId != boardId || targetColumnId == id {
			return model.ErrTargetColumnBelongsToBoard
		}
		if _, err := tx.Exec(ctx, "UPDATE tasks SET column_id = $2 WHERE column_id = $1", id, targetColumnId); err != nil {
			return err
		}
	}

	commandTag, err := tx.Exec(ctx, "DELETE FROM board_columns WHERE id = $1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrBoardColumnNotFound
	}
	return tx.Commit(ctx)
}

func getBoardColumns(ctx context.Context, connection *pgxpool.Conn, boardId int64) ([]*model.BoardColumn, error) {
	rows, err := connection.Query(ctx,
		"SELECT id, board_id, title, position FROM board_columns WHERE board_id = $1 ORDER BY position, id",
		boardId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*model.BoardColumn, 0)
	for rows.Next() {
		column := &model.BoardColumn{}
		if err := rows.Scan(&column.Id, &column.BoardId, &column.Title, &column.Position); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func mapBoardLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrBoardNotFound
	}
	return err
}

func mapBoardColumnLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrBoardColumnNotFound
	}
	return err
}
