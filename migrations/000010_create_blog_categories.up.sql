CREATE TABLE IF NOT EXISTS blog_categories (
  id bigserial PRIMARY KEY,
  image text NOT NULL,
  name text NOT NULL,
  slug text UNIQUE NOT NULL,
  type text NOT NULL,
  status varchar(255) NOT NULL,
  order_number integer,
  deleted_at timestamp(0) NULL DEFAULT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_blog_categories_updated_at BEFORE UPDATE
    ON blog_categories FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
