-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_modified_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION archive_deleted_row() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_archive SELECT (OLD).*;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE users (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name varchar(150) NOT NULL,
    username varchar(100) NOT NULL UNIQUE,
    password varchar(150) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE users_archive (
    id bigint PRIMARY KEY,
    name varchar(150) NOT NULL,
    username varchar(100) NOT NULL UNIQUE,
    password varchar(150) NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE conversations (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    is_deleting boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE participants (
    conv_id bigint NOT NULL,
    user_id bigint NOT NULL,
    PRIMARY KEY (conv_id, user_id),
    CONSTRAINT fk_part_conv FOREIGN KEY (conv_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_part_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE messages (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sender_id bigint NOT NULL,
    conv_id bigint NOT NULL,
    content text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_mess_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_mess_conv FOREIGN KEY (conv_id) REFERENCES conversations (id) ON DELETE CASCADE
);

-- +goose StatementBegin
CREATE TRIGGER update_users_modtime BEFORE
UPDATE ON users FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_messages_modtime BEFORE
UPDATE ON messages FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER archive_deleted_user BEFORE
DELETE ON users FOR EACH ROW
EXECUTE FUNCTION archive_deleted_row();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_deleted_user ON users;
DROP TRIGGER IF EXISTS update_messages_modtime ON messages;
DROP TRIGGER IF EXISTS update_users_modtime ON users;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS users_archive;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS archive_deleted_row();
DROP FUNCTION IF EXISTS update_modified_column();
-- +goose StatementEnd
