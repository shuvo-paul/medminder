-- Create dose_schedules table for medication dose schedules
CREATE TABLE IF NOT EXISTS dose_schedules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id  UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    time        TIME NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(profile_id, name)
);

CREATE INDEX idx_dose_schedules_profile_id ON dose_schedules(profile_id);