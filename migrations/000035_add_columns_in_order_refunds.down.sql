DROP TYPE order_refunds_status_enum;
ALTER TABLE order_refunds DROP COLUMN IF EXISTS status;
ALTER TABLE order_refunds DROP COLUMN IF EXISTS receipt_number;
ALTER TABLE order_refunds DROP COLUMN IF EXISTS refund_value;