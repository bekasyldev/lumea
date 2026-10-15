-- name: ListCategories :many
SELECT
    id,
    name,
    slug,
    created_at,
    updated_at
FROM categories
ORDER BY created_at, id;

-- name: GetCategoryBySlug :one
SELECT
    id,
    name,
    slug,
    created_at,
    updated_at
FROM categories
WHERE slug = $1
ORDER BY created_at;

-- name: GetCategoryByID :one
SELECT
    id,
    name,
    slug,
    created_at,
    updated_at
FROM categories
WHERE id = $1
ORDER BY created_at;

-- name: CreateCategory :one
INSERT INTO categories (
    name,
    slug
) VALUES($1, $2)
RETURNING *;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = $1;

-- name: UpdateCategory :execrows
UPDATE categories
SET name = $2, slug = $3, updated_at = now()
WHERE id = $1;
