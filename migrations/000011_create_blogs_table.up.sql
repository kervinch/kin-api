CREATE TABLE IF NOT EXISTS blogs (
  id bigserial PRIMARY KEY,
  blog_category_id bigint NOT NULL REFERENCES blog_categories ON DELETE CASCADE,
  thumbnail text NOT NULL,
  title text NOT NULL,
  description text NOT NULL,
  content text NOT NULL,
  slug text NOT NULL,
  type text NOT NULL,
  published_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  feature bool NOT NULL,
  status varchar(255) NOT NULL,
  tags text DEFAULT NULL,
  created_by integer NOT NULL,
  deleted_at timestamp(0) NULL DEFAULT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  created_by_text text DEFAULT NULL
);

CREATE TRIGGER update_blogs_updated_at BEFORE UPDATE
    ON blogs FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
