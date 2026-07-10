CREATE TABLE categorie (
    categorie_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100),
    catalog_id UUID NOT NULL REFERENCES catalog(catalog_id) ON DELETE CASCADE
);
