CREATE TABLE IF NOT EXISTS inbox_users (
  id bigserial PRIMARY KEY,
  inbox_id bigint NOT NULL REFERENCES inbox ON DELETE CASCADE,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  is_read bool NOT NULL DEFAULT TRUE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_inbox_users
ON inbox_users(inbox_id, user_id);

CREATE TRIGGER update_inbox_users_updated_at BEFORE UPDATE
    ON inbox_users FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
