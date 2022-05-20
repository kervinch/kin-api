CREATE TABLE IF NOT EXISTS user_vouchers (
  id bigserial PRIMARY KEY,
  user_id bigint REFERENCES users ON DELETE CASCADE,
  voucher_id bigint REFERENCES vouchers ON DELETE CASCADE,
  quantity integer NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_user_vouchers_updated_at BEFORE UPDATE
    ON user_vouchers FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
