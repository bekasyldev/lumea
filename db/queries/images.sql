-- name: CreateProductImage :one
-- new image goes to the end of the gallery (max position + 1 or 0)
INSERT INTO product_images (
    product_id,
    variant_id,
    s3_key,
    content_type,
    position
)
SELECT
    sqlc.arg(product_id)::bigint,
    sqlc.narg(variant_id)::bigint,
    sqlc.arg(s3_key)::text,
    sqlc.arg(content_type)::text,
    COALESCE(MAX(position) + 1, 0)
FROM product_images
WHERE product_id = sqlc.arg(product_id)::bigint
RETURNING
    id,
    product_id,
    variant_id,
    s3_key,
    content_type,
    position,
    created_at;

-- name: GetProductImageByID :one
SELECT
    id,
    product_id,
    variant_id,
    s3_key,
    content_type,
    position,
    created_at
FROM product_images
WHERE id = $1;

-- name: ListProductImages :many
SELECT
    id,
    product_id,
    variant_id,
    s3_key,
    content_type,
    position,
    created_at
FROM product_images
WHERE product_id = $1
ORDER BY position, id;

-- name: ListVariantImages :many
-- images of the variant + shared images of its product
SELECT
    id,
    product_id,
    variant_id,
    s3_key,
    content_type,
    position,
    created_at
FROM product_images
WHERE product_id = sqlc.arg(product_id)::bigint
  AND (variant_id = sqlc.arg(variant_id)::bigint OR variant_id IS NULL)
ORDER BY position, id;

-- name: GetMainProductImage :one
SELECT
    id,
    product_id,
    variant_id,
    s3_key,
    content_type,
    position,
    created_at
FROM product_images
WHERE product_id = $1
ORDER BY position, id
LIMIT 1;

-- name: UpdateProductImagePosition :execrows
UPDATE product_images
SET position = $2
WHERE id = $1;

-- name: DeleteProductImage :one
-- returns the key so the service can delete the file from S3
DELETE FROM product_images
WHERE id = $1
RETURNING s3_key;

-- name: ListProductImageKeys :many
SELECT s3_key
FROM product_images
WHERE product_id = $1;
