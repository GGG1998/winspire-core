-- name: GetOAuthProviderLink :one
SELECT id, user_id, provider, provider_user_id, provider_email, linked_at, last_used_at
FROM public.oauth_provider_links
WHERE user_id = $1 AND provider = $2;

-- name: CreateOAuthProviderLink :one
INSERT INTO public.oauth_provider_links (
  user_id, provider, provider_user_id, provider_email, linked_at
) VALUES (
  $1, $2, $3, $4, NOW()
)
RETURNING id, user_id, provider, provider_user_id, provider_email, linked_at, last_used_at;

-- name: UpdateOAuthProviderLinkLastUsed :exec
UPDATE public.oauth_provider_links
SET last_used_at = NOW()
WHERE user_id = $1 AND provider = $2;

