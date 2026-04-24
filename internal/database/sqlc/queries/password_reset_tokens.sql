-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (
    id,
    user_id,
    token_hash,
    expires_at
) VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token_hash, expires_at, used_at, created_at;

-- name: FindPasswordResetTokenByHash :one
SELECT id, user_id, token_hash, expires_at, used_at, created_at
FROM password_reset_tokens
WHERE token_hash = $1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: DeletePasswordResetTokensByUserID :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1;