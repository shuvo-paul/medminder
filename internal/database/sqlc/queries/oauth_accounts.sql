-- name: CreateOAuthAccount :one
INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, connected_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOAuthAccountByProviderAndUserID :one
SELECT * FROM oauth_accounts
WHERE provider = $1 AND provider_user_id = $2;

-- name: GetOAuthAccountsByUserID :many
SELECT * FROM oauth_accounts
WHERE user_id = $1;

-- name: GetOAuthAccountByUserIDAndProvider :one
SELECT * FROM oauth_accounts
WHERE user_id = $1 AND provider = $2;

-- name: DeleteOAuthAccount :one
DELETE FROM oauth_accounts
WHERE id = $1
RETURNING *;

-- name: DeleteOAuthAccountByUserIDAndProvider :exec
DELETE FROM oauth_accounts
WHERE user_id = $1 AND provider = $2;