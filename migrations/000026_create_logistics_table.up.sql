CREATE TABLE IF NOT EXISTS logistics (
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  type text NOT NULL,
  is_active bool NOT NULL DEFAULT TRUE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_logistics_updated_at BEFORE UPDATE
    ON logistics FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
