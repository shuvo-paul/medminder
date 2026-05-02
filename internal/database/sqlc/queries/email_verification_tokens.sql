-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (
    id,
    user_id,
    token_hash,
    expires_at
) VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token_hash, created_at, expires_at;

-- name: FindValidEmailVerificationToken :one
SELECT id, user_id, token_hash, created_at, expires_at
FROM email_verification_tokens
WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeleteEmailVerificationToken :exec
DELETE FROM email_verification_tokens
WHERE id = $1;

-- name: DeleteAllEmailVerificationTokensForUser :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;

-- name: CountEmailVerificationTokensCreatedToday :one
SELECT COUNT(*)
FROM email_verification_tokens
WHERE user_id = $1 AND created_at > (NOW() - INTERVAL '24 hours');