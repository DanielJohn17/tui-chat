-- +goose Up
-- +goose StatementBegin
ALTER TABLE participants
ADD COLUMN last_read_message_id bigint NOT NULL DEFAULT 0,
ADD COLUMN last_read_at timestamptz NOT NULL DEFAULT NOW();

CREATE INDEX idx_messages_conv_unread ON messages (conv_id, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_conv_unread;

ALTER TABLE participants
DROP COLUMN IF EXISTS last_read_at,
DROP COLUMN IF EXISTS last_read_message_id;
-- +goose StatementEnd
