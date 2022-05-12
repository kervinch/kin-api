CREATE TABLE IF NOT EXISTS inbox (
  id bigserial PRIMARY KEY,
  title text NOT NULL,
  content text NOT NULL,
  image_url text,
  deeplink text,
  slug text UNIQUE NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_inbox_updated_at BEFORE UPDATE
    ON inbox FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
