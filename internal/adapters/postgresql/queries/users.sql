-- name: CreateUser :one
INSERT INTO users (name, email, password, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    role = COALESCE($4, role),
    is_active = COALESCE($5, is_active),
    is_verified = COALESCE($6, is_verified),
    is_deleted = COALESCE($7, is_deleted)
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET 
    password = $2
WHERE id = $1
RETURNING id, email;