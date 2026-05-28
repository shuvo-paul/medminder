ALTER TABLE oauth_authorization_codes
    DROP COLUMN IF EXISTS provider,
    DROP COLUMN IF EXISTS provider_user_id,
    DROP COLUMN IF EXISTS provider_email,
    DROP COLUMN IF EXISTS provider_name,
    DROP COLUMN IF EXISTS provider_email_verified;
