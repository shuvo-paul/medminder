-- Create oauth_authorization_codes table
CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    nonce VARCHAR(255) NOT NULL,
    purpose VARCHAR(20) NOT NULL CHECK (purpose IN ('register', 'login', 'link')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index on code_hash for fast lookup during token exchange
CREATE INDEX IF NOT EXISTS idx_oauth_authorization_codes_code_hash
    ON oauth_authorization_codes(code_hash);

-- Index on expires_at for cleanup of expired codes
CREATE INDEX IF NOT EXISTS idx_oauth_authorization_codes_expires_at
    ON oauth_authorization_codes(expires_at);

-- Index on user_id for fast lookups (consistent with other token tables)
CREATE INDEX IF NOT EXISTS idx_oauth_authorization_codes_user_id
    ON oauth_authorization_codes(user_id);