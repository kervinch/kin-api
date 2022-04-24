CREATE TABLE IF NOT EXISTS user_addresses (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  name text NOT NULL,
  receiver text NOT NULL,
  phone_number text NOT NULL,
  city text NOT NULL,
  postal_code text NOT NULL,
  address text NOT NULL,
  is_main bool NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_user_addresses_updated_at BEFORE UPDATE
    ON user_addresses FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
