CREATE TABLE sales (
    sale_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    element JSONB,
    price DECIMAL(10, 2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    payments JSONB,
    is_deleted BOOLEAN DEFAULT FALSE
);
