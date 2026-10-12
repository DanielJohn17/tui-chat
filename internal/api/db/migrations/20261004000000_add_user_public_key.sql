-- +goose Up
ALTER TABLE users ADD COLUMN public_key bytea;
ALTER TABLE users_archive ADD COLUMN public_key bytea;

-- +goose Down
ALTER TABLE users_archive DROP COLUMN IF EXISTS public_key;
ALTER TABLE users DROP COLUMN IF EXISTS public_key;
