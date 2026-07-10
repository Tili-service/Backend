CREATE TABLE sale_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES sales(sale_id) ON DELETE CASCADE,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by_profile_id UUID REFERENCES profile(profile_id) ON DELETE SET NULL,
    lines JSONB,
    price DECIMAL(10, 2),
    payments JSONB,
    time_stamp TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    changes JSONB
);

CREATE INDEX idx_sale_history_sale_id ON sale_history(sale_id);
