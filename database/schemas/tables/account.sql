CREATE TABLE account (
    account_id SERIAL PRIMARY KEY,
    licence_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);
