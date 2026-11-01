-- +goose Up
ALTER TABLE product_variants 
ADD COLUMN stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0);
-- +goose Down
ALTER TABLE product_variants
DROP COLUMN stock;
