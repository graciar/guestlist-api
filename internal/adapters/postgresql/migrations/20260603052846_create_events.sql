-- +goose Up
CREATE TYPE event_type AS ENUM (
    'public',
    'private', 
    'reserved'
);

CREATE TYPE event_status AS ENUM (
    'open',
    'closed',
    'cancelled'
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    type event_type NOT NULL DEFAULT 'private',
    status event_status NOT NULL DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_type;
DROP TYPE IF EXISTS event_status;