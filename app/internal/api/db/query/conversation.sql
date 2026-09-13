-- name: GetOrCreateDirectConversation :many
WITH
  existing AS (
    SELECT
      p1.conv_id
    FROM
      participants p1
      JOIN participants p2 ON p1.conv_id = p2.conv_id
    WHERE
      p1.user_id = $1
      AND p2.user_id = $2
    LIMIT
      1
  ),
  new_conv AS (
    INSERT INTO
      conversations (created_at)
    SELECT
      CURRENT_TIMESTAMP
    WHERE
      NOT EXISTS (
        SELECT
          1
        FROM
          existing
      )
    RETURNING
      id
  ),
  inserted_p1 AS (
    INSERT INTO
      participants (conv_id, user_id)
    SELECT
      id,
      $1
    FROM
      new_conv
  ),
  inserted_p2 AS (
    INSERT INTO
      participants (conv_id, user_id)
    SELECT
      id,
      $2
    FROM
      new_conv
  ),
  target_conv AS (
    SELECT
      conv_id AS id
    FROM
      existing
    UNION ALL
    SELECT
      id
    FROM
      new_conv
  )
SELECT
  tc.id AS conv_id,
  u.id AS user_id,
  u.name,
  u.username
FROM
  target_conv tc
  JOIN participants p ON p.conv_id = tc.id
  JOIN users u ON u.id = p.user_id;

-- name: GetConversationsByUserId :many
WITH
  target_conv AS (
    SELECT
      conv_id
    FROM
      participants
    WHERE
      user_id = $1
  )
SELECT
  tc.conv_id,
  u.id AS user_id,
  u.name,
  u.username
FROM
  target_conv tc
  JOIN participants p ON p.conv_id = tc.conv_id and p.user_id <> $1
  JOIN users u ON u.id = p.user_id;

-- name: GetConvChats :many
SELECT
  id,
  sender_id,
  content,
  created_at,
  updated_at
FROM
  messages
WHERE
  conv_id = $1
ORDER BY
  created_at DESC,
  id DESC
LIMIT
  $2;

-- name: GetConvChatsPaginated :many
SELECT
  id,
  sender_id,
  content,
  created_at,
  updated_at
FROM
  messages
WHERE
  conv_id = $1
  AND (
    created_at < $2
    OR (
      created_at = $2
      AND id < $3
    )
  )
ORDER BY
  created_at DESC,
  id DESC
LIMIT
  $4;

-- name: CreateMessageAndGetRecipient :one
WITH
  verified_sender AS (
    SELECT
      conv_id
    FROM
      participants
    WHERE
      user_id = sqlc.arg(sender_id)
      AND conv_id = sqlc.arg(conv_id)
  ),
  recipient AS (
    SELECT
      user_id AS recipient_id
    FROM
      participants
    WHERE
      user_id <> sqlc.arg(sender_id)
      AND conv_id = sqlc.arg(conv_id)
    LIMIT
      1
  ),
  inserted_msg AS (
    INSERT INTO
      messages (conv_id, sender_id, content)
    SELECT
      vs.conv_id AS conv_id,
      sqlc.arg(sender_id) AS sender_id,
      sqlc.arg(content) AS content
    FROM
      verified_sender vs
    RETURNING
      id,
      conv_id,
      sender_id,
      content,
      created_at,
      updated_at
  )
SELECT
  im.id,
  im.conv_id,
  im.sender_id,
  r.recipient_id,
  im.content,
  im.created_at,
  im.updated_at
FROM
  inserted_msg im
  CROSS JOIN recipient r;
