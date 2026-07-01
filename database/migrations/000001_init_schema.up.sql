CREATE TABLE account (
    account_id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE licence (
    licence_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id INTEGER NOT NULL REFERENCES account(account_id) ON DELETE CASCADE,
    expiration TIMESTAMP NOT NULL,
    next_payment TIMESTAMP NOT NULL,
    transaction VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);
CREATE TABLE store (
    store_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    buyer_id INTEGER NOT NULL REFERENCES account(account_id),
    licence_id UUID NOT NULL REFERENCES licence(licence_id),
    date_creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    numero_tva VARCHAR(50),
    siret VARCHAR(14)
);
CREATE TABLE catalog (
    catalog_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    store_id INTEGER NOT NULL REFERENCES store(store_id) ON DELETE CASCADE
);
CREATE TABLE categorie (
    categorie_id SERIAL PRIMARY KEY,
    type VARCHAR(100),
    catalog_id INTEGER NOT NULL REFERENCES catalog(catalog_id) ON DELETE CASCADE
);
CREATE TABLE item (
    item_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    tax DECIMAL(5, 4) NOT NULL,
    tax_amount DECIMAL(10, 2) NOT NULL,
    categorie_id INTEGER REFERENCES categorie(categorie_id) NOT NULL
);
CREATE TABLE profile (
    profile_id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES store(store_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    pin VARCHAR(10) NOT NULL,
    level_access INTEGER NOT NULL DEFAULT 4,
    is_active BOOLEAN DEFAULT TRUE
);
CREATE TABLE sales (
    sale_id SERIAL PRIMARY KEY,
    element JSONB,
    price DECIMAL(10, 2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    payments JSONB,
    is_deleted BOOLEAN DEFAULT FALSE
);
CREATE TABLE sale_history (
    history_id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL REFERENCES sales(sale_id) ON DELETE CASCADE,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by_profile_id INTEGER REFERENCES profile(profile_id) ON DELETE SET NULL,
    lines JSONB,
    price DECIMAL(10, 2),
    payments JSONB,
    time_stamp TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    changes JSONB
);

CREATE TABLE image (
    image_id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    url TEXT
);
CREATE TABLE payementmethod (
    payement_method_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_sale_history_sale_id ON sale_history(sale_id);

