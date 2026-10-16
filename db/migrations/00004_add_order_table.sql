-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_name  TEXT     NOT NULL,
    customer_phone TEXT     NOT NULL,
    comment        TEXT,
    total_price BIGINT      NOT NULL CHECK (total_price > 0),
    status      TEXT        NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- pending -> confirmed -> shipped -> delivered
    CONSTRAINT orders_status_check CHECK (
        status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled')
    )
);

-- order table should be a snapshot
CREATE TABLE IF NOT EXISTS order_items ( 
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id            BIGINT  NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_variant_id  BIGINT  NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    product_name        TEXT    NOT NULL,                          
    variant_name        TEXT    NOT NULL,                          
    quantity            INTEGER NOT NULL CHECK (quantity > 0),
    unit_price          BIGINT  NOT NULL CHECK (unit_price > 0),   -- KZT per unit, after wholesale rule

    CONSTRAINT order_items_order_variant_unique UNIQUE (order_id, product_variant_id)
);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
