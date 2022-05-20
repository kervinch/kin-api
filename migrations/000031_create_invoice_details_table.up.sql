CREATE TABLE IF NOT EXISTS invoice_details (
  id bigserial PRIMARY KEY,
  order_detail_id bigint REFERENCES order_details ON DELETE CASCADE,
  product_detail_id bigint REFERENCES product_details ON DELETE CASCADE,
  product_name text NOT NULL,
  quantity integer NOT NULL,
  price integer NOT NULL,
  total integer NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_invoice_details_updated_at BEFORE UPDATE
    ON invoice_details FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
