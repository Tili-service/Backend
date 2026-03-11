-- access_level: 1 = superAdmin, 2 = admin, 3 = user

CREATE TABLE "user" (
    user_id SERIAL PRIMARY KEY,
    store_id INTEGER REFERENCES store(store_id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    access_code VARCHAR(50),
    access_level INTEGER
);
