-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
    name TEXT NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE product_variants (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id         BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name               TEXT        NOT NULL,            -- e.g. "50 ml", "Shade 02"
    price              BIGINT      NOT NULL CHECK (price > 0),   -- KZT, normal price
    wholesale_price    BIGINT      CHECK (wholesale_price > 0),  -- price for big orders
    wholesale_min_qty  INTEGER     CHECK (wholesale_min_qty >= 2),
    is_active          BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- wholesale fields are either both set or both empty
    CONSTRAINT wholesale_pair CHECK (
        (wholesale_price IS NULL) = (wholesale_min_qty IS NULL)
    ),
    -- wholesale must actually be cheaper
    CONSTRAINT wholesale_cheaper CHECK (
        wholesale_price IS NULL OR wholesale_price < price
    )
);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

-- +goose Down
DROP TABLE product_variants;
DROP TABLE products;
