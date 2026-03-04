CREATE TABLE store (
    store_id SERIAL PRIMARY KEY,
    store_name VARCHAR(255) NOT NULL,
    account_id INTEGER REFERENCES account(account_id)
);
