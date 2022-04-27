CREATE TABLE IF NOT EXISTS product_storefront_subscriptions (
  id bigserial PRIMARY KEY,
  product_id bigint NOT NULL REFERENCES products ON DELETE CASCADE,
  storefront_id bigint NOT NULL REFERENCES storefronts ON DELETE CASCADE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_product_storefront
ON product_storefront_subscriptions(product_id, storefront_id);

CREATE TRIGGER update_product_storefront_subscriptions_updated_at BEFORE UPDATE
    ON product_storefront_subscriptions FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();
