CREATE TABLE catalogue (
    catalogue_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2),
    tax SMALLINT,
    categorie_id INTEGER REFERENCES categorie(categorie_id),
    image_id INTEGER REFERENCES image(image_id)
);
