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
