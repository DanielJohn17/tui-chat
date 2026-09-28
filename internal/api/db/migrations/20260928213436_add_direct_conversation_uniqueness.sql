-- +goose Up
-- +goose StatementBegin
ALTER TABLE conversations
ADD COLUMN user_min_id bigint REFERENCES users(id) ON DELETE CASCADE,
ADD COLUMN user_max_id bigint REFERENCES users(id) ON DELETE CASCADE,
ADD CONSTRAINT chk_user_order CHECK (user_min_id IS NULL OR user_min_id < user_max_id);

-- Backfill existing conversations from participants
UPDATE conversations c
SET
  user_min_id = p.min_id,
  user_max_id = p.max_id
FROM (
  SELECT conv_id, MIN(user_id) AS min_id, MAX(user_id) AS max_id
  FROM participants
  GROUP BY conv_id
  HAVING COUNT(*) = 2
) p
WHERE c.id = p.conv_id;

-- Create partial unique index on active direct conversations
CREATE UNIQUE INDEX uq_direct_conversations_active_pair
ON conversations (user_min_id, user_max_id)
WHERE is_deleting = FALSE AND user_min_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_direct_conversations_active_pair;

ALTER TABLE conversations
DROP CONSTRAINT IF EXISTS chk_user_order,
DROP COLUMN IF EXISTS user_min_id,
DROP COLUMN IF EXISTS user_max_id;
-- +goose StatementEnd
