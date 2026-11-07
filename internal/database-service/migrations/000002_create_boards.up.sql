CREATE TABLE boards (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    done_column_id BIGINT
);

CREATE TABLE board_columns (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    board_id BIGINT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    position INT NOT NULL CHECK (position >= 0)
);

ALTER TABLE boards
    ADD CONSTRAINT boards_done_column_fk
    FOREIGN KEY (done_column_id)
    REFERENCES board_columns(id)
    DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE tasks
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN column_id BIGINT REFERENCES board_columns(id),
    ADD COLUMN position INT NOT NULL DEFAULT 0 CHECK (position >= 0),
    ADD COLUMN progress INT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    ADD COLUMN is_done BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN completed_at TIMESTAMP;

INSERT INTO boards (owner_user_id, title)
SELECT DISTINCT COALESCE(user_id, 0), 'Основная доска'
FROM tasks;

INSERT INTO board_columns (board_id, title, position)
SELECT boards.id, defaults.title, defaults.position
FROM boards
CROSS JOIN (
    VALUES
        ('Backlog', 0),
        ('В работе', 1),
        ('Проверка', 2),
        ('Готово', 3)
) AS defaults(title, position);

UPDATE boards
SET done_column_id = board_columns.id
FROM board_columns
WHERE board_columns.board_id = boards.id
  AND board_columns.position = 3;

UPDATE tasks
SET column_id = board_columns.id,
    position = existing_tasks.position
FROM boards,
     board_columns,
     (
         SELECT id,
                ROW_NUMBER() OVER (PARTITION BY COALESCE(user_id, 0) ORDER BY id) - 1 AS position
         FROM tasks
     ) AS existing_tasks
WHERE existing_tasks.id = tasks.id
  AND boards.owner_user_id = COALESCE(tasks.user_id, 0)
  AND board_columns.board_id = boards.id
  AND board_columns.position = 0;

ALTER TABLE tasks
    ALTER COLUMN column_id SET NOT NULL;
