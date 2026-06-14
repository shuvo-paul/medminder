-- name: CreateUser :one
INSERT INTO users (
    email,
    display_name,
    password_hash,
    email_verified
) VALUES ($1, $2, $3, $4)
RETURNING id, email, display_name, email_verified, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, display_name, password_hash, email_verified, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, display_name, password_hash, email_verified, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users SET email_verified = true, updated_at = NOW() WHERE id = $1;

-- name: UpdateUserEmail :exec
UPDATE users SET email = $2, email_verified = true, updated_at = NOW() WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
