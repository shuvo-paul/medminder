-- Ownership transfer queries

-- name: CreateOwnershipTransfer :one
INSERT INTO ownership_transfers (
    profile_id,
    from_user_id,
    to_user_id,
    status,
    expires_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING id, profile_id, from_user_id, to_user_id, status, expires_at, created_at, updated_at;

-- name: GetOwnershipTransferByID :one
SELECT id, profile_id, from_user_id, to_user_id, status, expires_at, created_at, updated_at
FROM ownership_transfers
WHERE id = $1;

-- name: ListPendingTransfersByUser :many
SELECT id, profile_id, from_user_id, to_user_id, status, expires_at, created_at, updated_at
FROM ownership_transfers
WHERE to_user_id = $1 AND status = 'pending' AND expires_at > NOW();

-- name: ListPendingTransfersWithDetailsByUser :many
SELECT
    ot.id,
    ot.profile_id,
    ot.from_user_id,
    ot.to_user_id,
    ot.status,
    ot.expires_at,
    ot.created_at,
    ot.updated_at,
    p.name as profile_name,
    fu.display_name as from_name,
    tu.display_name as to_name
FROM ownership_transfers ot
JOIN profiles p ON p.id = ot.profile_id
JOIN users fu ON fu.id = ot.from_user_id
JOIN users tu ON tu.id = ot.to_user_id
WHERE ot.to_user_id = $1 AND ot.status = 'pending' AND ot.expires_at > NOW();

-- name: GetPendingTransferByProfile :one
SELECT id, profile_id, from_user_id, to_user_id, status, expires_at, created_at, updated_at
FROM ownership_transfers
WHERE profile_id = $1 AND status = 'pending';

-- name: UpdateOwnershipTransferStatus :one
UPDATE ownership_transfers
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, from_user_id, to_user_id, status, expires_at, created_at, updated_at;
