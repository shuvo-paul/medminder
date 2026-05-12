-- name: CreateAuthorizationCode :one
INSERT INTO oauth_authorization_codes (id, code_hash, user_id, nonce, purpose, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAuthorizationCodeByHash :one
SELECT * FROM oauth_authorization_codes
WHERE code_hash = $1;

-- name: MarkAuthorizationCodeAsUsed :one
UPDATE oauth_authorization_codes
SET used_at = NOW()
WHERE code_hash = $1 AND used_at IS NULL
RETURNING *;

-- name: DeleteExpiredAuthorizationCodes :many
DELETE FROM oauth_authorization_codes
WHERE expires_at < $1
RETURNING *;

-- name: DeleteAuthorizationCode :one
DELETE FROM oauth_authorization_codes
WHERE id = $1
RETURNING *;