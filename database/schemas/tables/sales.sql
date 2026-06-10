CREATE TABLE sales (
    sale_id SERIAL PRIMARY KEY,
    element JSONB,
    price DECIMAL(10, 2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
<<<<<<< HEAD
    payments JSONB
=======
    payement_method_id INTEGER REFERENCES payementmethod(payement_method_id),
    is_deleted BOOLEAN DEFAULT FALSE
>>>>>>> 62293eedd62065f24cfb34b845cca3edb78a2772
);
