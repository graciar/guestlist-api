-- name: CreatePasswordReset :one
INSERT INTO password_resets (
    user_id,
    token_hash,
    expires_at
) VALUES (
    $1,
    $2,
    $3
) RETURNING *;

-- name: GetPasswordResetByTokenHash :one
SELECT * FROM password_resets WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now();

-- name: InvalidatePasswordReset :execrows
UPDATE password_resets SET used_at = now() WHERE id = $1;