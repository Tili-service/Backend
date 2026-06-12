CREATE TABLE catalog (
    catalog_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    store_id INTEGER NOT NULL REFERENCES store(store_id) ON DELETE CASCADE
);
