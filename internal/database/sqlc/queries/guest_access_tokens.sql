-- Guest access token queries

-- name: CreateGuestAccessToken :one
INSERT INTO guest_access_tokens (
    profile_id,
    token_hash,
    label,
    permissions,
    expires_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING id, profile_id, token_hash, label, permissions, expires_at, created_at, last_used_at;

-- name: GetGuestAccessTokenByHash :one
SELECT id, profile_id, token_hash, label, permissions, expires_at, created_at, last_used_at
FROM guest_access_tokens
WHERE token_hash = $1;

-- name: GetGuestAccessTokenByID :one
SELECT id, profile_id, token_hash, label, permissions, expires_at, created_at, last_used_at
FROM guest_access_tokens
WHERE id = $1;

-- name: ListGuestAccessTokensByProfile :many
SELECT id, profile_id, token_hash, label, permissions, expires_at, created_at, last_used_at
FROM guest_access_tokens
WHERE profile_id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: UpdateGuestAccessTokenLastUsedAt :exec
UPDATE guest_access_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: DeleteGuestAccessToken :exec
DELETE FROM guest_access_tokens WHERE id = $1;

-- name: DeleteGuestAccessTokensByProfile :exec
DELETE FROM guest_access_tokens WHERE profile_id = $1;
