-- Create oauth_accounts table
CREATE TABLE IF NOT EXISTS oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint: one MedMinder user per provider account
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_accounts_provider_user_id
    ON oauth_accounts(provider, provider_user_id);

-- Unique constraint: one account per provider per user (enforces one Google per account)
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_accounts_user_provider
    ON oauth_accounts(user_id, provider);