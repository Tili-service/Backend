CREATE TABLE sales (
    sale_id SERIAL PRIMARY KEY,
    element JSONB,
    price DECIMAL(10, 2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    payments JSONB
);
