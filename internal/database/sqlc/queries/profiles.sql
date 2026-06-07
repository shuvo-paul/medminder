-- Profile queries

-- name: CreateProfile :one
INSERT INTO profiles (
    owner_user_id,
    name,
    date_of_birth,
    timezone
) VALUES ($1, $2, $3, $4)
RETURNING id, owner_user_id, name, date_of_birth, timezone, created_at, updated_at;

-- name: GetProfileByID :one
SELECT id, owner_user_id, name, date_of_birth, timezone, created_at, updated_at
FROM profiles
WHERE id = $1;

-- name: ListProfilesByOwner :many
SELECT id, owner_user_id, name, date_of_birth, timezone, created_at, updated_at
FROM profiles
WHERE owner_user_id = $1
ORDER BY created_at DESC;

-- name: UpdateProfile :one
UPDATE profiles
SET name = $2, date_of_birth = $3, timezone = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, owner_user_id, name, date_of_birth, timezone, created_at, updated_at;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE id = $1;

-- Dose Schedule queries

-- name: CreateDoseSchedule :one
INSERT INTO dose_schedules (
    profile_id,
    name,
    time
) VALUES ($1, $2, $3)
RETURNING id, profile_id, name, time, created_at, updated_at;

-- name: GetDoseScheduleByID :one
SELECT id, profile_id, name, time, created_at, updated_at
FROM dose_schedules
WHERE id = $1;

-- name: ListDoseSchedulesByProfile :many
SELECT id, profile_id, name, time, created_at, updated_at
FROM dose_schedules
WHERE profile_id = $1
ORDER BY time ASC;

-- name: UpdateDoseSchedule :one
UPDATE dose_schedules
SET name = $2, time = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, profile_id, name, time, created_at, updated_at;

-- name: DeleteDoseSchedule :exec
DELETE FROM dose_schedules WHERE id = $1;

-- name: DeleteDoseSchedulesByProfile :exec
DELETE FROM dose_schedules WHERE profile_id = $1;