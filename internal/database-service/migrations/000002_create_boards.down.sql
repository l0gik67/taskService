ALTER TABLE tasks
    DROP COLUMN completed_at,
    DROP COLUMN is_done,
    DROP COLUMN progress,
    DROP COLUMN position,
    DROP COLUMN column_id,
    DROP COLUMN description;

ALTER TABLE boards
    DROP CONSTRAINT boards_done_column_fk;

DROP TABLE board_columns;
DROP TABLE boards;
