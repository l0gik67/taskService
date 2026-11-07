ALTER TABLE boards
    DROP CONSTRAINT boards_owner_user_id_fk;

ALTER TABLE tasks
    DROP CONSTRAINT tasks_user_id_fk,
    ALTER COLUMN user_id DROP NOT NULL;

DROP TABLE users;
