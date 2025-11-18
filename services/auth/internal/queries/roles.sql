-- name: GetUserRoles :many
SELECT r.id, r.name, r.description
FROM public.roles r
JOIN public.user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetUserPermissions :many
SELECT DISTINCT p.id, p.name, p.resource, p.action
FROM public.permissions p
JOIN public.role_permissions rp ON p.id = rp.permission_id
JOIN public.user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1;

-- name: CheckUserPermission :one
SELECT EXISTS(
  SELECT 1
  FROM public.permissions p
  JOIN public.role_permissions rp ON p.id = rp.permission_id
  JOIN public.user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = $1 AND p.name = $2
) as has_permission;

