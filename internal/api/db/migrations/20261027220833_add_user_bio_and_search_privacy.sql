-- +goose Up
ALTER TABLE users ADD COLUMN bio varchar(255) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN is_searchable boolean NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN invite_token varchar(64);

UPDATE users 
SET invite_token = substr(md5(random()::text || clock_timestamp()::text || id::text), 1, 32)
WHERE invite_token IS NULL;

ALTER TABLE users ALTER COLUMN invite_token SET DEFAULT substr(md5(random()::text || clock_timestamp()::text), 1, 32);
ALTER TABLE users ALTER COLUMN invite_token SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_invite_token_unique UNIQUE (invite_token);

ALTER TABLE users_archive ADD COLUMN bio varchar(255) NOT NULL DEFAULT '';
ALTER TABLE users_archive ADD COLUMN is_searchable boolean NOT NULL DEFAULT TRUE;
ALTER TABLE users_archive ADD COLUMN invite_token varchar(64);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION archive_deleted_row() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_archive (id, name, username, password, created_at, updated_at, public_key, bio, is_searchable, invite_token)
    VALUES (OLD.id, OLD.name, OLD.username, OLD.password, OLD.created_at, OLD.updated_at, OLD.public_key, OLD.bio, OLD.is_searchable, OLD.invite_token);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION archive_deleted_row() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_archive (id, name, username, password, created_at, updated_at)
    VALUES (OLD.id, OLD.name, OLD.username, OLD.password, OLD.created_at, OLD.updated_at);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

ALTER TABLE users_archive DROP COLUMN IF EXISTS invite_token;
ALTER TABLE users_archive DROP COLUMN IF EXISTS is_searchable;
ALTER TABLE users_archive DROP COLUMN IF EXISTS bio;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_invite_token_unique;
ALTER TABLE users DROP COLUMN IF EXISTS invite_token;
ALTER TABLE users DROP COLUMN IF EXISTS is_searchable;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
