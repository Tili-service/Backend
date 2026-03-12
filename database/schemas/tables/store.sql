CREATE TABLE store (
    store_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    buyer_id INTEGER NOT NULL REFERENCES account(account_id),
    licence_id UUID NOT NULL REFERENCES licence(licence_id),
    date_creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    numero_tva VARCHAR(50),
    siret VARCHAR(14)
);
