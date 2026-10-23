-- +goose Up
CREATE TABLE product_images (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id  BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id  BIGINT      REFERENCES product_variants(id) ON DELETE CASCADE, -- NULL = whole product
    s3_key      TEXT        NOT NULL UNIQUE,  
    content_type TEXT       NOT NULL,
    position    INTEGER     NOT NULL DEFAULT 0, -- order the product images shown
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id, position);
CREATE INDEX idx_product_images_variant_id ON product_images(variant_id);

-- +goose Down
DROP TABLE product_images;
