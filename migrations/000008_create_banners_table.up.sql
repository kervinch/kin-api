CREATE TABLE IF NOT EXISTS banners (
    id bigserial PRIMARY KEY,
    image_url text NOT NULL,
    title text NOT NULL,
    deeplink text,
    outbound_url text,
    is_active bool NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_banners_updated_at BEFORE UPDATE
    ON banners FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();