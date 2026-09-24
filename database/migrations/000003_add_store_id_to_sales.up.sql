-- sales.store_id was previously added by editing the already-applied
-- 000001 migration, which never reaches databases that already ran it.
-- This migration adds the column properly instead.
--
-- The column is added nullable: existing sales rows predate store scoping
-- and there is no reliable way to attribute them to a store, so they are
-- left with a NULL store_id rather than guessed or deleted. They will not
-- appear in any store-scoped query or KPI report until an operator
-- backfills them manually. New sales always set store_id (enforced in
-- application code), so this only affects rows created before this
-- migration runs.
ALTER TABLE sales
    ADD COLUMN store_id UUID REFERENCES store(store_id) ON DELETE CASCADE;

CREATE INDEX idx_sales_store_id ON sales(store_id, time_stamp);
