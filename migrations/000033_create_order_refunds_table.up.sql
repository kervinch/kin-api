CREATE TABLE IF NOT EXISTS order_refunds (
  id bigserial PRIMARY KEY,
  order_detail_id bigint REFERENCES order_details ON DELETE CASCADE,
  brand_id bigint REFERENCES brands ON DELETE CASCADE,
  image_1 text NOT NULL,
  image_2 text NOT NULL,
  image_3 text NOT NULL,
  video text NOT NULL,
  explanation text NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_order_refunds_updated_at BEFORE UPDATE
    ON order_refunds FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
