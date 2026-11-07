package repository

import (
	"context"
	"core/model"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getCommonSql = `
SELECT
	id,
	title,
	description,
	user_id,
	column_id,
	position,
	progress,
	is_done,
	completed_at,
	duration,
	start_time,
	priority
FROM tasks
`
	getTaskSql          = getCommonSql + "WHERE id = $1;"
	getTasksByUserIdSql = getCommonSql + "WHERE user_id = $1 AND ($2 = 0 OR column_id IN (SELECT id FROM board_columns WHERE board_id = $2));"
	createTaskSql       = `
INSERT INTO tasks (title, description, user_id, column_id, position, progress, is_done, completed_at, duration, start_time, priority)
SELECT $1,
       $2,
       $3,
       CASE WHEN $9 OR $4 = boards.done_column_id THEN boards.done_column_id ELSE $4 END,
       $5,
       $6,
       $9 OR $4 = boards.done_column_id,
       CASE WHEN $9 OR $4 = boards.done_column_id THEN NOW() ELSE NULL END,
       $7,
       $8,
       $10
FROM board_columns
JOIN boards ON boards.id = board_columns.board_id
WHERE board_columns.id = $4
RETURNING id, column_id, is_done, completed_at;
`
	updateTaskSql = `
UPDATE tasks
SET title = $2,
    description = $3,
    user_id = $4,
    column_id = CASE WHEN $10 OR $5 = boards.done_column_id THEN boards.done_column_id ELSE $5 END,
    position = $6,
    progress = $7,
    is_done = $10 OR $5 = boards.done_column_id,
    completed_at = CASE
        WHEN $10 OR $5 = boards.done_column_id THEN COALESCE(tasks.completed_at, NOW())
        ELSE NULL
    END,
    duration = $8,
    start_time = $9,
    priority = $11
FROM board_columns
JOIN boards ON boards.id = board_columns.board_id
WHERE tasks.id = $1
  AND board_columns.id = $5
RETURNING tasks.column_id, tasks.is_done, tasks.completed_at;
`
	deleteTaskSql = `
DELETE FROM tasks
WHERE id = $1;
`
)

type TaskRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &TaskRepositoryImpl{
		pool: pool,
	}
}

func (repo *TaskRepositoryImpl) GetById(id int64) (*model.Task, error) {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO: add logs
		return nil, err
	}
	defer connection.Release()

	result := connection.QueryRow(context.Background(), getTaskSql, id)
	task := model.Task{}
	var durationSeconds int32

	err = scanTask(result, &task, &durationSeconds)
	if err != nil {
		// TODO add logs
		return nil, mapTaskLookupError(err)
	}
	task.Duration = time.Duration(durationSeconds) * time.Second

	return &task, nil
}

func mapTaskLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrTaskNotFound
	}
	return err
}

func (repo *TaskRepositoryImpl) GetByUserId(userId int64, boardId int64) ([]*model.Task, error) {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO: add logs
		return nil, err
	}
	defer connection.Release()

	result, err := connection.Query(context.Background(), getTasksByUserIdSql, userId, boardId)
	if err != nil {
		// TODO: add logs
		return nil, err
	}
	defer result.Close()

	tasks := make([]*model.Task, 0, 100)
	for result.Next() {
		task := model.Task{}
		var durationSeconds int32
		err := scanTask(result, &task, &durationSeconds)
		if err != nil {
			// TODO: add logs
			return nil, err
		}
		task.Duration = time.Duration(durationSeconds) * time.Second
		tasks = append(tasks, &task)
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (repo *TaskRepositoryImpl) Create(task *model.Task) (*model.Task, error) {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO add logs
		return nil, err
	}
	defer connection.Release()

	durationSeconds := int64(task.Duration / time.Second)
	result := connection.QueryRow(context.Background(),
		createTaskSql,
		task.Title,
		task.Description,
		task.UserId,
		task.ColumnId,
		task.Position,
		task.Progress,
		durationSeconds,
		task.StartTime,
		task.IsDone,
		task.Priority,
	)

	err = result.Scan(&task.Id, &task.ColumnId, &task.IsDone, &task.CompletedAt)
	if err != nil {
		// TODO add logs
		return nil, err
	}
	task.Duration = time.Duration(durationSeconds) * time.Second

	return task, nil
}

func (repo *TaskRepositoryImpl) Update(task *model.Task) error {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO add logs
		return err
	}
	defer connection.Release()

	err = connection.QueryRow(context.Background(),
		updateTaskSql,
		task.Id,
		task.Title,
		task.Description,
		task.UserId,
		task.ColumnId,
		task.Position,
		task.Progress,
		int64(task.Duration/time.Second),
		task.StartTime,
		task.IsDone,
		task.Priority,
	).Scan(&task.ColumnId, &task.IsDone, &task.CompletedAt)
	if err != nil {
		// TODO add logs
		return err
	}
	return nil
}

type taskRowScanner interface {
	Scan(destinations ...any) error
}

func scanTask(scanner taskRowScanner, task *model.Task, durationSeconds *int32) error {
	return scanner.Scan(
		&task.Id,
		&task.Title,
		&task.Description,
		&task.UserId,
		&task.ColumnId,
		&task.Position,
		&task.Progress,
		&task.IsDone,
		&task.CompletedAt,
		durationSeconds,
		&task.StartTime,
		&task.Priority,
	)
}

func (repo *TaskRepositoryImpl) Delete(taskId int64) error {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO add logs
		return err
	}
	defer connection.Release()

	_, err = connection.Exec(context.Background(), deleteTaskSql, taskId)

	if err != nil {
		// TODO add logs
		return err
	}
	return nil
}
