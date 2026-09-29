-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION archive_deleted_row() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_archive (id, name, username, password, created_at, updated_at)
    VALUES (OLD.id, OLD.name, OLD.username, OLD.password, OLD.created_at, OLD.updated_at);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION archive_deleted_row() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_archive SELECT (OLD).*;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
