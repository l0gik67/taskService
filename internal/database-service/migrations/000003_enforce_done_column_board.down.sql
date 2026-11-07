ALTER TABLE boards
    DROP CONSTRAINT boards_done_column_fk;

ALTER TABLE boards
    ADD CONSTRAINT boards_done_column_fk
    FOREIGN KEY (done_column_id)
    REFERENCES board_columns(id)
    DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE board_columns
    DROP CONSTRAINT board_columns_board_id_id_unique;
