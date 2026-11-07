ALTER TABLE board_columns
    ADD CONSTRAINT board_columns_board_id_id_unique
    UNIQUE (board_id, id);

ALTER TABLE boards
    DROP CONSTRAINT boards_done_column_fk;

ALTER TABLE boards
    ADD CONSTRAINT boards_done_column_fk
    FOREIGN KEY (id, done_column_id)
    REFERENCES board_columns(board_id, id)
    DEFERRABLE INITIALLY DEFERRED;
