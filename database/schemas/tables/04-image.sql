CREATE TABLE image (
    image_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    url TEXT
);
