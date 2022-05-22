CREATE TABLE IF NOT EXISTS order_shippings (
  id bigserial PRIMARY KEY,
  order_id bigint REFERENCES orders ON DELETE CASCADE,
  logistic_id bigint REFERENCES logistics ON DELETE CASCADE,
  subtotal bigint NOT NULL,
  voucher_id bigint REFERENCES vouchers ON DELETE CASCADE,
  total bigint NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_order_shippings_updated_at BEFORE UPDATE
    ON order_shippings FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
