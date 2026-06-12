-- Profile permission queries

-- name: CreateProfilePermission :one
INSERT INTO profile_permissions (
    profile_id,
    shared_with_user_id,
    granted_by_user_id,
    permissions,
    status,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at;

-- name: GetProfilePermission :one
SELECT id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at
FROM profile_permissions
WHERE profile_id = $1 AND shared_with_user_id = $2;

-- name: ListProfilePermissionsByProfile :many
SELECT id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at
FROM profile_permissions
WHERE profile_id = $1;

-- name: ListProfilePermissionsByUser :many
SELECT id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at
FROM profile_permissions
WHERE shared_with_user_id = $1;

-- name: DeleteProfilePermission :exec
DELETE FROM profile_permissions WHERE id = $1;

-- name: UpdateProfilePermissionPermissions :one
UPDATE profile_permissions
SET permissions = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at;

-- name: GetProfilePermissionByID :one
SELECT id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at
FROM profile_permissions
WHERE id = $1;

-- name: AcceptProfilePermission :one
UPDATE profile_permissions
SET status = 'accepted', expires_at = NULL, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at;

-- name: UpdateProfilePermissionStatus :one
UPDATE profile_permissions
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at;

-- name: UpdateProfilePermissionExpiresAt :one
UPDATE profile_permissions
SET expires_at = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, shared_with_user_id, granted_by_user_id, permissions, status, expires_at, created_at, updated_at;
