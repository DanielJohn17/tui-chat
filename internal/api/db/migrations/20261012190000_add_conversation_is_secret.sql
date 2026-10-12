-- +goose Up
ALTER TABLE conversations ADD COLUMN is_secret boolean NOT NULL DEFAULT FALSE;

DROP INDEX IF EXISTS uq_direct_conversations_active_pair;
CREATE UNIQUE INDEX uq_direct_conversations_active_pair
ON conversations (user_min_id, user_max_id, is_secret)
WHERE is_deleting = FALSE AND user_min_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_direct_conversations_active_pair;
CREATE UNIQUE INDEX uq_direct_conversations_active_pair
ON conversations (user_min_id, user_max_id)
WHERE is_deleting = FALSE AND user_min_id IS NOT NULL;

ALTER TABLE conversations DROP COLUMN IF EXISTS is_secret;
