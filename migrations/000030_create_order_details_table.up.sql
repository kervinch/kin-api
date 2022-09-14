CREATE TYPE order_details_status_enum AS ENUM ('awaiting_payment', 'expired', 'paid', 'pending', 'processing', 'delivery', 'completed', 'refund_requested', 'refund_rejected', 'refund_completed');

CREATE TABLE IF NOT EXISTS order_details (
  id bigserial PRIMARY KEY,
  order_id bigint REFERENCES orders ON DELETE CASCADE,
  brand_id bigint REFERENCES brands ON DELETE CASCADE,
  invoice_number text NOT NULL,
  subtotal bigint,
  voucher_id bigint REFERENCES vouchers ON DELETE CASCADE,
  total bigint,
  status order_details_status_enum NOT NULL DEFAULT 'pending',
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_order_details_updated_at BEFORE UPDATE
    ON orders FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
