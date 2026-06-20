-- name: CreateEvent :one
INSERT INTO events (
    user_id, title, slug, description, 
    location, start_time, type, status
    )
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
    ) RETURNING *;

-- name: FindEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM events;

-- name: UpdateEvent :one
UPDATE events SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    location = COALESCE($4, location),
    start_time = COALESCE($5, start_time),
    type = COALESCE($6, type),
    status = COALESCE($7, status)
WHERE id = $1 RETURNING *;

-- name: DeleteEvent :one
DELETE FROM events WHERE id = $1 RETURNING *;

-- name: GetUserEvents :many
SELECT * FROM events WHERE user_id = $1;

-- name: GetUserEventStats :one
SELECT 
    e.user_id,
    
    -- Event Metrics
    COUNT(DISTINCT e.id) AS total_events,
    COUNT(DISTINCT CASE WHEN e.status = 'open' THEN e.id END) AS open_events,
    COUNT(DISTINCT CASE WHEN e.status = 'closed' THEN e.id END) AS closed_events,
    COUNT(DISTINCT CASE WHEN e.status = 'cancelled' THEN e.id END) AS cancelled_events,
    
    -- Guest Metrics
    COUNT(g.id) AS total_guests_invited,
    COUNT(CASE WHEN g.rsvp_status = 'attending' THEN 1 END) AS total_guests_confirmed,
    COUNT(CASE WHEN g.is_checked_in = TRUE THEN 1 END) AS total_guests_checked_in,
    
    -- Attendance Success Rate
    CASE 
        WHEN COUNT(g.id) > 0 THEN 
            ((COUNT(CASE WHEN g.is_checked_in = TRUE THEN 1 END)::NUMERIC / COUNT(g.id)::NUMERIC) * 100)::INTEGER
        ELSE 0 
    END AS attendance_rate_percentage

FROM events e
LEFT JOIN guests g ON e.id = g.event_id
WHERE e.user_id = $1
GROUP BY e.user_id;

-- name: GetEventStats :one
SELECT 
    e.id AS event_id,
    e.title,
    e.slug,
    e.status AS event_status,
    e.start_time,
    
    -- Guest Metrics for this single event
    COUNT(g.id) AS total_guests_invited,
    COUNT(CASE WHEN g.rsvp_status = 'attending' THEN 1 END) AS total_guests_confirmed,
    COUNT(CASE WHEN g.is_checked_in = TRUE THEN 1 END) AS total_guests_checked_in,
    
    -- Attendance Success Rate
    CASE 
        WHEN COUNT(g.id) > 0 THEN 
            ROUND((COUNT(CASE WHEN g.is_checked_in = TRUE THEN 1 END)::NUMERIC / COUNT(g.id)::NUMERIC) * 100, 2)
        ELSE 0 
    END AS attendance_rate_percentage

FROM events e
LEFT JOIN guests g ON e.id = g.event_id
WHERE e.id = $1
GROUP BY e.id;

-- name: GetEventByHostIDAndTitle :one
SELECT * FROM events WHERE user_id = $1 AND title = $2;
