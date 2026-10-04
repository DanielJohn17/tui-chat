-- +goose Up
ALTER TABLE users ADD COLUMN public_key bytea;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS public_key;
