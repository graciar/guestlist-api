-- +goose Up
CREATE TYPE rsvp_status AS ENUM (
    'pending',
    'attending',
    'declined'
);

CREATE TABLE IF NOT EXISTS guests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(100) NULL,
    phone VARCHAR(100) NULL,
    rsvp_status rsvp_status NOT NULL DEFAULT 'pending',
    rsvp_token VARCHAR(100) NULL,
    ticket_code VARCHAR(100) UNIQUE NULL,
    is_checked_in BOOLEAN NOT NULL DEFAULT FALSE,
    checked_in_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Inline constraint right here makes the code cleaner!
    CONSTRAINT unique_event_guest UNIQUE (event_id, name)
);

-- DROP TYPE IF EXISTS rsvp_status CASCADE;

-- +goose Down
DROP TABLE IF EXISTS guests;
DROP TYPE IF EXISTS rsvp_status;