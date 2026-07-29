CREATE TABLE store (
    store_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    buyer_id UUID NOT NULL REFERENCES account(account_id),
    licence_id UUID UNIQUE NOT NULL REFERENCES licence(licence_id),
    date_creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    numero_tva VARCHAR(50),
    siret VARCHAR(14)
);
