CREATE TABLE IF NOT EXISTS carts (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  product_detail_id bigint NOT NULL REFERENCES product_details ON DELETE CASCADE,
  quantity integer NOT NULL DEFAULT 1,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_carts
ON carts(user_id, product_detail_id);

CREATE TRIGGER update_carts_updated_at BEFORE UPDATE
    ON carts FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
