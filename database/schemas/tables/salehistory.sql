CREATE TABLE sale_history (
    history_id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL REFERENCES sales(sale_id) ON DELETE CASCADE,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by_profile_id INTEGER REFERENCES profile(profile_id) ON DELETE SET NULL,
    previous_lines JSONB,
    previous_price DECIMAL(10, 2),
    previous_payement_method_id INTEGER,
    previous_time_stamp TIMESTAMP
);

CREATE INDEX idx_sale_history_sale_id ON sale_history(sale_id);
