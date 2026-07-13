DROP TABLE IF EXISTS user_password_credentials;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS products;

DROP TYPE IF EXISTS user_role;

DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_customer_created_at;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer_id;

DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS order_status;
