-- Note: OAuth provider links are managed by Supabase in auth.identities table
-- Use Supabase Go client API methods instead of direct SQL queries:
-- - client.Auth.GetUserIdentities() to get all linked providers
-- - client.Auth.LinkIdentity() to link a provider
-- - client.Auth.UnlinkIdentity() to unlink a provider
-- See: https://supabase.com/docs/guides/auth/auth-identity-linking
--
-- If you need to query auth.identities directly (requires service_role key):
-- Note: These queries are optional - prefer using Supabase Go client APIs

-- name: GetOAuthIdentities :many
SELECT id, user_id, provider, identity_data, created_at, updated_at
FROM auth.identities
WHERE user_id = $1;

-- name: GetOAuthIdentityByProvider :one
SELECT id, user_id, provider, identity_data, created_at, updated_at
FROM auth.identities
WHERE user_id = $1 AND provider = $2;
