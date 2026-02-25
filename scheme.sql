CREATE TABLE account (
    account_id SERIAL PRIMARY KEY,
    licence_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE store (
    store_id SERIAL PRIMARY KEY,
    store_name VARCHAR(255) NOT NULL,
    account_id INTEGER REFERENCES account(account_id)
);

CREATE TABLE "user" (
    user_id SERIAL PRIMARY KEY,
    store_id INTEGER REFERENCES store(store_id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    access_code VARCHAR(50),
    access_level INTEGER
);

CREATE TABLE payementmethod (
    payement_method_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- 5. Table des ventes
CREATE TABLE vente (
    vente_id SERIAL PRIMARY KEY,
    element JSONB,
    price DECIMAL(10, 2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    payement_method_id INTEGER REFERENCES payementmethod(payement_method_id)
);

CREATE TABLE categorie (
    categorie_id SERIAL PRIMARY KEY,
    type VARCHAR(100)
);

CREATE TABLE image (
    image_id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    url TEXT
);

CREATE TABLE catalogue (
    catalogue_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2),
    tva SMALLINT,
    categorie_id INTEGER REFERENCES categorie(categorie_id),
    image_id INTEGER REFERENCES image(image_id)
);