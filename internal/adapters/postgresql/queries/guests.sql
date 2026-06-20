-- name: CreateGuest :one
INSERT INTO guests (
    event_id, name, email, phone, 
    rsvp_status, rsvp_token, ticket_code,
    is_checked_in, checked_in_at
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: FindGuestByID :one
SELECT 
    g.id AS guest_id,
    g.name AS guest_name,
    g.email AS guest_email,
    g.phone AS guest_phone,
    g.rsvp_status,
    g.ticket_code,
    g.is_checked_in,
    e.id AS event_id,
    e.user_id as host_id,
    e.title AS event_title,
    e.slug AS event_slug,
    e.description AS event_description,
    e.location AS event_title_location,
    e.start_time AS event_start_time
FROM guests g
INNER JOIN events e ON g.event_id = e.id
WHERE g.id = $1;

-- name: ListGuests :many
SELECT * FROM guests;

-- name: UpdateGuest :one
UPDATE guests SET
    event_id = COALESCE($2, event_id),
    name = COALESCE($3, name),
    email = COALESCE($4, email),
    phone = COALESCE($5, phone),
    rsvp_status = COALESCE($6, rsvp_status),
    rsvp_token = COALESCE($7, rsvp_token),
    ticket_code = COALESCE($8, ticket_code),
    is_checked_in = COALESCE($9, is_checked_in),
    checked_in_at = COALESCE($10, checked_in_at)
WHERE id = $1 RETURNING *;

-- name: DeleteGuest :one
DELETE FROM guests WHERE id = $1 RETURNING *;

-- name: GetGuestsByEventID :many
SELECT id, event_id, name, email, phone, rsvp_status, ticket_code, is_checked_in 
FROM guests 
WHERE event_id = $1;

-- name: GetRsvpGuestByToken :one
SELECT * FROM guests 
WHERE id = $1 AND rsvp_token = $2 AND rsvp_status = 'pending'
LIMIT 1
FOR UPDATE;

-- name: CheckGuestExists :one
SELECT EXISTS (
    SELECT 1 FROM guests 
    WHERE event_id = $1 AND LOWER(TRIM(name)) = LOWER(TRIM($2))
);