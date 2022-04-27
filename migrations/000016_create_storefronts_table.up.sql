CREATE TABLE IF NOT EXISTS storefronts (
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL,
  image_url text NOT NULL,
  slug text UNIQUE NOT NULL,
  is_active bool NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_storefronts_updated_at BEFORE UPDATE
    ON storefronts FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
