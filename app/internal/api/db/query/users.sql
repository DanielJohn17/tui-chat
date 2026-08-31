-- name: CreateUser :one
INSERT INTO
  users (name, username, password)
VALUES
  ($1, $2, $3)
RETURNING
  id,
  name,
  username,
  created_at,
  updated_at;

-- name: GetUserByUsername :one
SELECT
  (id, name, username)
FROM
  users
WHERE
  username = $1;

