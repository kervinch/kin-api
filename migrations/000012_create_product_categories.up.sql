CREATE TABLE IF NOT EXISTS product_categories (
  id bigserial PRIMARY KEY,
  image_url text,
  name text NOT NULL,
  slug text UNIQUE NOT NULL,
  order_number integer,
  is_active bool NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_product_categories_updated_at BEFORE UPDATE
    ON product_categories FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
