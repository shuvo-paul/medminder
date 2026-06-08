CREATE TABLE IF NOT EXISTS profile_permissions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id          UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    shared_with_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    granted_by_user_id  UUID NOT NULL REFERENCES users(id),
    permissions         JSONB NOT NULL DEFAULT '[]',
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(profile_id, shared_with_user_id)
);

CREATE UNIQUE INDEX idx_single_owner_per_profile ON profile_permissions(profile_id)
    WHERE permissions @> '["profile:owner"]';

CREATE INDEX idx_profile_permissions_profile_id ON profile_permissions(profile_id);
CREATE INDEX idx_profile_permissions_user_id ON profile_permissions(shared_with_user_id);
CREATE INDEX idx_profile_permissions_status ON profile_permissions(status);
