-- db/queries.sql
-- name: CreateOrder :exec
INSERT INTO orders (
    id, customer_id, status, total_price, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1;

-- name: UpdateOrder :exec
UPDATE orders
SET status = $1, total_price = $2, updated_at = $3
WHERE id = $4;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateOrderItem :exec
INSERT INTO order_items (
    id, order_id, product_id, quantity, price
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1;

-- name: DeleteOrderItems :exec
DELETE FROM order_items
WHERE order_id = $1;

-- name: DeleteOrderItem :exec
DELETE FROM order_items
WHERE id = $1 AND order_id = $2;