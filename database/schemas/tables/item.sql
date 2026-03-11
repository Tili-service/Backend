CREATE TABLE item (
    item_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    tax DECIMAL(5, 4) NOT NULL,
    tax_amount DECIMAL(10, 2) NOT NULL,
    categorie_id INTEGER REFERENCES categorie(categorie_id) NOT NULL,
)
