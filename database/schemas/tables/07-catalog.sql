CREATE TABLE catalog (
    catalog_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    store_id UUID NOT NULL REFERENCES store(store_id) ON DELETE CASCADE
);
