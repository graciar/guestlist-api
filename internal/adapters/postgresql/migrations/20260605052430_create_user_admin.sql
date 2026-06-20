-- +goose Up
-- 1. Enable the crypto extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 2. Insert the admin user with a live bcrypt hash
INSERT INTO users (
    name, 
    email, 
    password, 
    role
) 
VALUES (
    'admin', 
    'admin@admin.com', 
    crypt('admin', gen_salt('bf', 10)), -- Hashes 'admin123' using Bcrypt
    'admin'
);

-- +goose Down
DELETE FROM users WHERE email = 'admin@admin.com';