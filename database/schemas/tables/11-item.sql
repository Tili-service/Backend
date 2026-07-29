CREATE TABLE item (
    item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    tax DECIMAL(5, 4) NOT NULL,
    tax_amount DECIMAL(10, 2) NOT NULL,
    categorie_id UUID REFERENCES categorie(categorie_id) NOT NULL
);
