-- name: CreateProduct :one
INSERT INTO products (
    id, sku, name, description, category, price, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;