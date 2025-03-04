-- Drop indexs
DROP INDEX IF EXITS idx_order_items_product_id;
DROP INDEX IF EXITS idx_order_items_order_id;
DROP INDEX IF EXITS idx_orders_status;
DROP INDEX IF EXITS idx_orders_cusstomer_id;

-- Drop tables
DROP TABLE IF EXITS order_items;
DROP TABLE IF EXITS orders;