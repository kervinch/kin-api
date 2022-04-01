CREATE TABLE IF NOT EXISTS brands (
    id bigserial PRIMARY KEY,
    image_url text NOT NULL,
    name text NOT NULL,
    order_number integer,
    is_active bool NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_brands_updated_at BEFORE UPDATE
    ON brands FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();