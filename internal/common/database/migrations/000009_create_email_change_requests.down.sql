-- Drop email_change_requests table
DROP INDEX IF EXISTS idx_email_change_requests_user_id;
DROP INDEX IF EXISTS idx_email_change_requests_token_hash;
DROP INDEX IF EXISTS idx_email_change_requests_expires_at;
DROP TABLE IF EXISTS email_change_requests;
