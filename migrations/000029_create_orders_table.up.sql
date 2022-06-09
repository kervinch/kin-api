CREATE TYPE orders_status_enum AS ENUM ('awaiting_payment', 'expired', 'paid', 'pending', 'processing', 'delivery' 'completed', 'refund_requested', 'refund_rejected', 'refund_completed');

CREATE TABLE IF NOT EXISTS orders (
  id bigserial PRIMARY KEY,
  user_id bigint REFERENCES users ON DELETE CASCADE,
  receiver text NOT NULL,
  phone_number text NOT NULL,
  city text NOT NULL,
  postal_code text NOT NULL,
  address text NOT NULL,
  subtotal bigint,
  voucher_id bigint REFERENCES vouchers ON DELETE CASCADE,
  total bigint,
  status orders_status_enum NOT NULL DEFAULT 'awaiting_payment',
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE
    ON orders FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
