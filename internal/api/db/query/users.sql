-- name: CreateUser :one
INSERT INTO
users (name, username, password)
VALUES
($1, $2, $3)
RETURNING
    id,
    name,
    username,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at;

-- name: SearchUsers :many
SELECT id, name, username, bio
FROM users
WHERE id <> sqlc.arg(user_id)::bigint
  AND is_searchable = TRUE
  AND LOWER(username) LIKE LOWER(sqlc.arg(pattern)::text) ESCAPE '\'
ORDER BY (LOWER(username) = LOWER(sqlc.arg(query)::text)) DESC,
         LOWER(username) COLLATE "C", username COLLATE "C", id
LIMIT 20;

-- name: GetUserByUsername :one
SELECT
    id,
    name,
    username,
    password,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at
FROM
    users
WHERE
    username = $1;

-- name: GetUserById :one
SELECT
    id,
    name,
    username,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at
FROM
    users
WHERE
    id = $1;

-- name: GetUserByInviteToken :one
SELECT
    id,
    name,
    username,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at
FROM
    users
WHERE
    invite_token = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET
    name = $2,
    bio = $3,
    is_searchable = $4,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    name,
    username,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at;

-- name: RegenerateInviteToken :one
UPDATE users
SET
    invite_token = $2,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    name,
    username,
    bio,
    is_searchable,
    invite_token,
    created_at,
    updated_at;

-- name: DeleteUser :one
DELETE FROM users
WHERE
    id = $1
RETURNING
    id,
    name,
    username;

-- name: SetUserPublicKey :exec
UPDATE users SET public_key = $1 WHERE id = $2;

-- name: GetUserPublicKey :one
SELECT public_key FROM users WHERE id = $1;
