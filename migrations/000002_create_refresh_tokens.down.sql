DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP TABLE IF EXISTS refresh_tokens;