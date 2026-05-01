CREATE TABLE sale_history (
    history_id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL REFERENCES sale(sale_id) ON DELETE CASCADE,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by_account_id INTEGER REFERENCES account(account_id) ON DELETE SET NULL,
    previous_lines JSONB,
    previous_price DECIMAL(10, 2),
    previous_payement_method_id INTEGER,
    previous_time_stamp TIMESTAMP
);

CREATE INDEX idx_sale_history_sale_id ON sale_history(sale_id);
