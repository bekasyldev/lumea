-- name: ListOrders :many
SELECT
    id,
    customer_name,
    customer_phone,
    comment,
    total_price,
    status,
    created_at,
    updated_at
FROM orders
ORDER BY created_at DESC, id DESC;

-- name: GetOrderByID :one
SELECT
    id,
    customer_name,
    customer_phone,
    comment,
    total_price,
    status,
    created_at,
    updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderItemByID :one
SELECT
    id,
    order_id,
    product_variant_id,
    product_name,
    variant_name,
    quantity,
    unit_price
FROM order_items
WHERE id = $1;

-- name: ListOrderItemsByOrderID :many
SELECT
    id,
    order_id,
    product_variant_id,
    product_name,
    variant_name,
    quantity,
    unit_price
FROM order_items
WHERE order_id = $1
ORDER BY id;

-- name: CreateOrder :one
INSERT INTO orders (
    customer_name,
    customer_phone,
    comment,
    total_price
) VALUES ($1, $2, $3, $4)
RETURNING
    id,
    customer_name,
    customer_phone,
    comment,
    total_price,
    status,
    created_at,
    updated_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    product_variant_id,
    product_name,
    variant_name,
    quantity,
    unit_price
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    id,
    order_id,
    product_variant_id,
    product_name,
    variant_name,
    quantity,
    unit_price;

-- name: GetVariantsForOrder :many
SELECT
    v.id,
    v.name AS variant_name,
    p.name AS product_name,
    v.price,
    v.wholesale_price,
    v.wholesale_min_qty
FROM product_variants v
JOIN products p ON p.id = v.product_id
WHERE v.id = ANY(@ids::BIGINT[]) AND v.is_active AND p.is_active;

-- name: UpdateOrder :execrows
UPDATE orders
SET
    customer_name = $2,
    customer_phone = $3,
    comment = $4,
    updated_at = now()
WHERE id = $1;

-- name: UpdateOrderStatus :execrows
UPDATE orders
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: DecreaseVariantStock :execrows
UPDATE product_variants
SET stock = stock - $2
WHERE id = $1 AND stock >= $2;

-- name: ReturnVariantToStock :exec
UPDATE product_variants
SET stock = stock + $2
WHERE id = $1;

-- name: LockOrderForUpdate :one
SELECT id
FROM orders
WHERE id = $1
FOR UPDATE;

-- name: CountOrderItems :one
SELECT count(*)
FROM order_items
WHERE order_id = $1;

-- name: RecalculateOrderTotal :execrows
UPDATE orders
SET
    total_price = (
        SELECT COALESCE(SUM(quantity::BIGINT * unit_price), 0)
        FROM order_items
        WHERE order_id = $1
    ),
    updated_at = now()
WHERE id = $1;

-- name: DeleteOrderItem :execrows
DELETE FROM order_items
WHERE id = $1 AND order_id = $2;

-- name: DeleteOrder :execrows
DELETE FROM orders
WHERE id = $1;


