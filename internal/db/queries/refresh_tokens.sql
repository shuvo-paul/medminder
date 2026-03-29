-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id,
    token_hash,
    expires_at
) VALUES ($1, $2, $3)
RETURNING id, user_id, expires_at, created_at;

-- name: GetRefreshTokenByHash :one
SELECT rt.id, rt.user_id, rt.token_hash, rt.expires_at, rt.created_at
FROM refresh_tokens rt
WHERE rt.token_hash = $1
  AND rt.expires_at > NOW();

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE id = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;
