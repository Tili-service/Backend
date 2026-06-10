CREATE TABLE categorie (
    categorie_id SERIAL PRIMARY KEY,
    type VARCHAR(100),
    catalog_id INTEGER NOT NULL REFERENCES catalog(catalog_id) ON DELETE CASCADE
);
