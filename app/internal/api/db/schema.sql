-- Reusable Trigger Function
CREATE OR REPLACE FUNCTION update_modified_column () RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create tables
CREATE TABLE users (
  id bigint PRIMARY KEY,
  name varchar(150) NOT NULL,
  username varchar(100) NOT NULL UNIQUE,
  password varchar(150) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);


CREATE TABLE conversations (
  id bigint PRIMARY KEY,
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
  id bigint PRIMARY KEY,
  sender_id bigint NOT NULL,
  conv_id bigint NOT NULL,
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_mess_sender FOREIGN key (sender_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_mess_conv FOREIGN key (conv_id) REFERENCES conversations (id) ON DELETE CASCADE
);


-- Trigger bind
CREATE TRIGGER update_users_modtime BEFORE
UPDATE ON users FOR EACH ROW
EXECUTE FUNCTION update_modified_column ();


CREATE TRIGGER update_messages_modtime before
UPDATE ON messages FOR each ROW
EXECUTE function update_modified_column ();
