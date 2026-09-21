-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_messages_conv_id_created_at_id 
ON messages (conv_id, created_at DESC, id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_conv_id_created_at_id;
-- +goose StatementEnd
