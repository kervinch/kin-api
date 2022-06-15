CREATE TYPE vouchers_enum AS ENUM ('brand', 'ship', 'total');

CREATE TABLE IF NOT EXISTS vouchers (
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL,
  terms_and_condition text NOT NULL,
  type vouchers_enum NOT NULL DEFAULT 'total',
  image_url text,
  slug text UNIQUE NOT NULL,
  brand_id bigint REFERENCES brands ON DELETE CASCADE,
  logistic_id bigint REFERENCES logistics ON DELETE CASCADE,
  code text NOT NULL,
  is_percent bool NOT NULL DEFAULT TRUE,
  value integer NOT NULL,
  stock integer NOT NULL,
  is_active bool NOT NULL DEFAULT TRUE,
  effective_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  expired_at timestamp(0) with time zone NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  created_by bigint REFERENCES users ON DELETE CASCADE,
  updated_by bigint REFERENCES users ON DELETE CASCADE
);

CREATE TRIGGER update_vouchers_updated_at BEFORE UPDATE
    ON vouchers FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
