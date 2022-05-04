CREATE TYPE condition_enum AS ENUM ('new', 'used');

CREATE TABLE IF NOT EXISTS products (
  id bigserial PRIMARY KEY,
  product_category_id bigint NOT NULL REFERENCES product_categories ON DELETE CASCADE,
  brand_id bigint NOT NULL REFERENCES brands ON DELETE CASCADE,
  name text NOT NULL,
  description text NOT NULL,
  weight integer NOT NULL,
  minimum_order integer NOT NULL,
  preorder_days integer NOT NULL DEFAULT 0,
  condition condition_enum NOT NULL DEFAULT 'new',
  slug text UNIQUE NOT NULL,
  insurance_required bool NOT NULL DEFAULT TRUE,
  is_active bool NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_products_updated_at BEFORE UPDATE
    ON products FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
