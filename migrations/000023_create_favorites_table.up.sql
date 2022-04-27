CREATE TABLE IF NOT EXISTS favorites (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  product_detail_id bigint NOT NULL REFERENCES product_details ON DELETE CASCADE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_favorites
ON favorites(user_id, product_detail_id);

CREATE TRIGGER update_favorites_updated_at BEFORE UPDATE
    ON favorites FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
