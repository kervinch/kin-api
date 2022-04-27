CREATE TABLE IF NOT EXISTS product_details (
  id bigserial PRIMARY KEY,
  product_id bigint NOT NULL REFERENCES products ON DELETE CASCADE,
  color text,
  size text,
  price bigint NOT NULL,
  SKU text,
  stock integer NOT NULL,
  is_active bool NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_product_details_updated_at BEFORE UPDATE
    ON product_details FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
