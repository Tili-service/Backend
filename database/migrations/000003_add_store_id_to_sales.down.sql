DROP INDEX IF EXISTS idx_sales_store_id;
ALTER TABLE sales DROP COLUMN IF EXISTS store_id;
