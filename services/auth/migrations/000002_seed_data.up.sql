-- Insert predefined roles
INSERT INTO public.roles (name, description) VALUES
  ('streamer', 'Users who stream content'),
  ('user', 'Regular platform users'),
  ('admin', 'System administrators')
ON CONFLICT (name) DO NOTHING;

-- Insert predefined permissions
INSERT INTO public.permissions (name, resource, action, description) VALUES
  ('tournament:create', 'tournament', 'create', 'Create tournaments'),
  ('tournament:read', 'tournament', 'read', 'View tournaments'),
  ('tournament:update', 'tournament', 'update', 'Update tournaments'),
  ('tournament:delete', 'tournament', 'delete', 'Delete tournaments'),
  ('stream:manage', 'stream', 'manage', 'Manage stream settings'),
  ('user:read', 'user', 'read', 'View user profiles'),
  ('user:update', 'user', 'update', 'Update user profiles')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles
-- Streamer role permissions
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r, public.permissions p
WHERE r.name = 'streamer'
  AND p.name IN ('tournament:create', 'tournament:read', 'tournament:update', 'stream:manage', 'user:read')
ON CONFLICT DO NOTHING;

-- User role permissions
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r, public.permissions p
WHERE r.name = 'user'
  AND p.name IN ('tournament:read', 'user:read', 'user:update')
ON CONFLICT DO NOTHING;

-- Admin role permissions (all permissions)
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.roles r, public.permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

