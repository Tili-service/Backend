CREATE TABLE licence (
    licence_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id INTEGER NOT NULL REFERENCES account(account_id) ON DELETE CASCADE,
    expiration TIMESTAMP NOT NULL,
    transaction VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);
