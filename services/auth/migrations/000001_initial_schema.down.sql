-- Drop indexes
DROP INDEX IF EXISTS public.idx_oauth_provider_links_provider;
DROP INDEX IF EXISTS public.idx_oauth_provider_links_user_id;
DROP INDEX IF EXISTS public.idx_role_permissions_role_id;
DROP INDEX IF EXISTS public.idx_user_roles_role_id;
DROP INDEX IF EXISTS public.idx_user_roles_user_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS public.oauth_provider_links;
DROP TABLE IF EXISTS public.role_permissions;
DROP TABLE IF EXISTS public.permissions;
DROP TABLE IF EXISTS public.user_roles;
DROP TABLE IF EXISTS public.roles;

