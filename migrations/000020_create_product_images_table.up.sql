CREATE TABLE IF NOT EXISTS product_images (
  id bigserial PRIMARY KEY,
  product_detail_id bigint NOT NULL REFERENCES product_details ON DELETE CASCADE,
  image_url text NOT NULL,
  is_main bool NOT NULL DEFAULT FALSE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_product_images_updated_at BEFORE UPDATE
    ON product_images FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
