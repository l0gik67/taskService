CREATE TABLE tasks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    user_id BIGINT,
    duration INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    priority int NOT NULL
);
