-- name: ListProducts :many
SELECT
    id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at,
    category_id
FROM products
ORDER BY created_at;

-- name: GetProductByID :one
SELECT
    id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at,
    category_id
FROM products
WHERE id = $1;


-- name: ListProductVariants :many
SELECT
    id,
    product_id,
    name,
    price,
    wholesale_price,
    wholesale_min_qty,
    is_active,
    created_at,
    updated_at
FROM product_variants
WHERE product_id = $1
ORDER BY created_at;

-- name: CreateProduct :exec
INSERT INTO products (
    category_id,
    name, 
    slug, 
    description, 
    is_active
) VALUES($1, $2, $3, $4, $5);


-- name: CreateProductVariant :exec
INSERT INTO product_variants (
    product_id, 
    name, 
    price, 
    wholesale_price, 
    wholesale_min_qty,
    is_active
) VALUES($1, $2, $3, $4, $5, $6);


