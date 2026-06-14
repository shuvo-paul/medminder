CREATE TABLE IF NOT EXISTS guest_access_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    token_hash      VARCHAR(64) NOT NULL UNIQUE,
    label           VARCHAR(100),
    permissions     JSONB NOT NULL DEFAULT '["medication:read", "reminder:read"]',
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_guest_tokens_profile ON guest_access_tokens(profile_id);
CREATE INDEX idx_guest_tokens_hash ON guest_access_tokens(token_hash);
