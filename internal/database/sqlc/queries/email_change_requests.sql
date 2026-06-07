-- name: CreateEmailChangeRequest :one
INSERT INTO email_change_requests (
    user_id,
    new_email,
    token_hash,
    expires_at
) VALUES ($1, $2, $3, $4)
RETURNING id, user_id, new_email, token_hash, expires_at, created_at;

-- name: FindValidEmailChangeRequestByTokenHash :one
SELECT id, user_id, new_email, token_hash, expires_at, created_at
FROM email_change_requests
WHERE token_hash = $1 AND expires_at > NOW();

-- name: GetPendingEmailChangeRequestByUserID :one
SELECT id, user_id, new_email, token_hash, expires_at, created_at
FROM email_change_requests
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteEmailChangeRequest :exec
DELETE FROM email_change_requests WHERE id = $1;

-- name: DeleteEmailChangeRequestsByUserID :exec
DELETE FROM email_change_requests WHERE user_id = $1;