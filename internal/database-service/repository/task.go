package repository

import (
	"context"
	"time"

	"database-service/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The duration column stores whole seconds as INTEGER.
const (
	getCommonSql = `
SELECT
	id,
	title,
	user_id,
	duration,
	start_time,
	priority
FROM tasks
`
	getTaskSql          = getCommonSql + "WHERE id = $1;"
	getTasksByUserIdSql = getCommonSql + "WHERE user_id = $1;"
	createTaskSql       = `
INSERT INTO tasks (title, user_id, duration, start_time, priority)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
`
	updateTaskSql = `
UPDATE tasks
SET title = $2,
    user_id = $3,
    duration = $4,
    start_time = $5,
    priority = $6
WHERE id = $1;
`
	deleteTaskSql = `
DELETE FROM tasks
WHERE id = $1;
`
)

type TaskRepository interface {
	GetById(id int64) (*model.Task, error)
	GetByUserId(userId int64) ([]*model.Task, error)
	Create(task *model.Task) (*model.Task, error)
	Update(task *model.Task) error
	Delete(taskId int64) error
}

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

	err = result.Scan(&task.Id, &task.Title, &task.UserId, &durationSeconds, &task.StartTime, &task.Priority)
	if err != nil {
		// TODO add logs
		return nil, err
	}
	task.Duration = time.Duration(durationSeconds) * time.Second

	return &task, nil
}

func (repo *TaskRepositoryImpl) GetByUserId(userId int64) ([]*model.Task, error) {
	connection, err := repo.pool.Acquire(context.Background())
	if err != nil {
		// TODO: add logs
		return nil, err
	}
	defer connection.Release()

	result, err := connection.Query(context.Background(), getTasksByUserIdSql, userId)
	if err != nil {
		// TODO: add logs
		return nil, err
	}
	defer result.Close()

	tasks := make([]*model.Task, 0, 100)
	for result.Next() {
		task := model.Task{}
		var durationSeconds int32
		err := result.Scan(&task.Id, &task.Title, &task.UserId, &durationSeconds, &task.StartTime, &task.Priority)
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
		task.UserId,
		durationSeconds,
		task.StartTime,
		task.Priority,
	)

	err = result.Scan(&task.Id)
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

	_, err = connection.Exec(context.Background(),
		updateTaskSql,
		task.Id,
		task.Title,
		task.UserId,
		int64(task.Duration/time.Second),
		task.StartTime,
		task.Priority,
	)
	if err != nil {
		// TODO add logs
		return err
	}
	return nil
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
