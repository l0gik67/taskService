CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

INSERT INTO users (id, name) OVERRIDING SYSTEM VALUE
SELECT existing_users.id, 'Пользователь ' || existing_users.id
FROM (
    SELECT user_id AS id
    FROM tasks
    WHERE user_id > 0
    UNION
    SELECT owner_user_id AS id
    FROM boards
    WHERE owner_user_id > 0
) AS existing_users
ORDER BY existing_users.id;

DO $$
DECLARE
    legacy_user_id BIGINT;
BEGIN
    IF EXISTS (SELECT 1 FROM tasks WHERE user_id IS NULL OR user_id <= 0)
        OR EXISTS (SELECT 1 FROM boards WHERE owner_user_id <= 0) THEN
        INSERT INTO users (name)
        VALUES ('Импортированный пользователь')
        RETURNING id INTO legacy_user_id;

        UPDATE tasks
        SET user_id = legacy_user_id
        WHERE user_id IS NULL OR user_id <= 0;

        UPDATE boards
        SET owner_user_id = legacy_user_id
        WHERE owner_user_id <= 0;
    END IF;
END $$;

SELECT setval(
    pg_get_serial_sequence('users', 'id'),
    COALESCE(MAX(id), 1),
    MAX(id) IS NOT NULL
)
FROM users;

ALTER TABLE tasks
    ALTER COLUMN user_id SET NOT NULL,
    ADD CONSTRAINT tasks_user_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT;

ALTER TABLE boards
    ADD CONSTRAINT boards_owner_user_id_fk
        FOREIGN KEY (owner_user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT;
