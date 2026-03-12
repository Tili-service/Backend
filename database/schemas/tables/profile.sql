CREATE TABLE profile (
    profile_id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES store(store_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    pin VARCHAR(10) NOT NULL,
    level_access INTEGER NOT NULL DEFAULT 4,
    is_active BOOLEAN DEFAULT TRUE
);
