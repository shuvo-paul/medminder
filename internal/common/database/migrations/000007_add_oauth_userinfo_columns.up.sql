-- Add user info columns to oauth_authorization_codes.
-- The callback handler fetches user info from the OAuth provider and stores it here
-- so the token exchange endpoint can call GetOrCreateUserByOAuth without re-fetching.
ALTER TABLE oauth_authorization_codes
    ADD COLUMN provider                VARCHAR(50)  NOT NULL DEFAULT '',
    ADD COLUMN provider_user_id       VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN provider_email          VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN provider_name           VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN provider_email_verified BOOLEAN      NOT NULL DEFAULT FALSE;
