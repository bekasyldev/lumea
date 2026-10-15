-- +goose Up
CREATE TABLE IF NOT EXISTS categories (
    id  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    slug    TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE products
ADD category_id BIGINT;

ALTER TABLE products
ADD CONSTRAINT fk_products_category_id
FOREIGN KEY (category_id) REFERENCES categories(id)
ON DELETE SET NULL;

CREATE INDEX idx_products_category_id ON products(category_id);

-- +goose Down
-- dropping the column also drops its foreign key and index
ALTER TABLE products DROP COLUMN category_id;
DROP TABLE categories;
