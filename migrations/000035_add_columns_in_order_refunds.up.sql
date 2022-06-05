CREATE TYPE order_refunds_status_enum AS ENUM ('processing', 'reject_immediately', 'refund_immediately', 'return', 'refund');
ALTER TABLE order_refunds ADD COLUMN status order_refunds_status_enum NOT NULL DEFAULT 'processing';
ALTER TABLE order_refunds ADD COLUMN receipt_number text;
ALTER TABLE order_refunds ADD COLUMN refund_value bigint;