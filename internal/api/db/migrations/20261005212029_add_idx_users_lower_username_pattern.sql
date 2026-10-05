-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_users_lower_username_pattern 
ON users (LOWER(username) text_pattern_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_lower_username_pattern;
-- +goose StatementEnd
