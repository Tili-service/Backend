CREATE TABLE profile (
    profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES store(store_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    pin VARCHAR(10) NOT NULL,
    level_access INTEGER NOT NULL DEFAULT 4,
    is_active BOOLEAN DEFAULT TRUE
);
