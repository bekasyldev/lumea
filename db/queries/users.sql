-- name: CreateUser :one
INSERT INTO users (username, hashed_password)
VALUES ($1, $2)
RETURNING id, username, role, created_at;

-- name: GetUser :one
SELECT id, username, role, created_at FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, role, created_at FROM users
ORDER BY id;

-- name: GetUserByUsername :one
SELECT username, hashed_password, role FROM users
WHERE username = $1;