# Data Model: Supabase Authentication Integration

**Date**: 2025-21-18  
**Feature**: Supabase Authentication Integration

## Overview

This document defines the data model for the authentication system, including entities, relationships, validation rules, and state transitions. The model integrates with Supabase Auth for identity management and extends it with custom role-based access control (RBAC).

---

## Entities

### 1. User Account (Supabase Auth)

**Source**: Supabase `auth.users` table (managed by Supabase)  
**Description**: Core user identity managed by Supabase Auth

**Fields** (from Supabase):
- `id` (UUID, PRIMARY KEY) - User unique identifier
- `email` (TEXT, UNIQUE) - User email address
- `encrypted_password` (TEXT) - Hashed password (managed by Supabase)
- `email_confirmed_at` (TIMESTAMP) - Email verification timestamp
- `created_at` (TIMESTAMP) - Account creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp
- `user_metadata` (JSONB) - Custom user metadata
  - `role` (STRING) - Primary role (for JWT claims)
  - `user_type` (STRING) - 'streamer' or 'user' (determined by frontend)
  - `display_name` (STRING, optional) - User display name
- `app_metadata` (JSONB) - Application metadata (admin-managed)
- `phone` (TEXT, optional) - Phone number (if phone auth enabled)
- `phone_confirmed_at` (TIMESTAMP, optional) - Phone verification timestamp

**Validation Rules**:
- Email must be valid format (enforced by Supabase)
- Password must meet strength requirements (enforced by Supabase)
- Email must be unique across all users
- `user_type` must be either 'streamer' or 'user'

**State Transitions**:
1. **Unverified** → **Verified**: User clicks email verification link
2. **Active** → **Suspended**: Admin action (via app_metadata)
3. **Active** → **Deleted**: Soft delete (Supabase handles)

**Relationships**:
- One-to-many with `user_roles` (via `user_id`)
- One-to-many with `oauth_provider_links` (via `user_id`)
- One-to-many with `password_reset_tokens` (via `user_id`)

---

### 2. Role

**Source**: Custom table `public.roles`  
**Description**: Defines user roles in the system (Streamer, User, Admin, etc.)

**Fields**:
- `id` (UUID, PRIMARY KEY, DEFAULT gen_random_uuid())
- `name` (VARCHAR(50), UNIQUE, NOT NULL) - Role identifier (e.g., 'streamer', 'user', 'admin')
- `description` (TEXT, optional) - Human-readable role description
- `created_at` (TIMESTAMP, DEFAULT NOW())
- `updated_at` (TIMESTAMP, DEFAULT NOW())

**Validation Rules**:
- `name` must be lowercase, alphanumeric with underscores
- `name` must be unique
- Reserved role names: 'system', 'internal' (cannot be created)

**Predefined Roles**:
- `streamer` - Users who stream content
- `user` - Regular platform users
- `admin` - System administrators (future)

**Relationships**:
- Many-to-many with `users` (via `user_roles`)
- Many-to-many with `permissions` (via `role_permissions`)

---

### 3. User Role Assignment

**Source**: Custom table `public.user_roles`  
**Description**: Maps users to their assigned roles (many-to-many relationship)

**Fields**:
- `user_id` (UUID, NOT NULL, REFERENCES auth.users(id) ON DELETE CASCADE)
- `role_id` (UUID, NOT NULL, REFERENCES roles(id) ON DELETE CASCADE)
- `assigned_at` (TIMESTAMP, DEFAULT NOW())
- `assigned_by` (UUID, optional, REFERENCES auth.users(id)) - Admin who assigned the role
- PRIMARY KEY (`user_id`, `role_id`)

**Validation Rules**:
- User can have multiple roles
- Cannot assign duplicate role to same user
- `user_id` must reference existing user in `auth.users`
- `role_id` must reference existing role

**State Transitions**:
- **Created**: Role assigned to user
- **Deleted**: Role removed from user (triggers JWT refresh)

**Relationships**:
- Many-to-one with `users` (via `user_id`)
- Many-to-one with `roles` (via `role_id`)

---

### 4. Permission

**Source**: Custom table `public.permissions`  
**Description**: Defines granular permissions for resources and actions

**Fields**:
- `id` (UUID, PRIMARY KEY, DEFAULT gen_random_uuid())
- `name` (VARCHAR(100), UNIQUE, NOT NULL) - Permission identifier (e.g., 'tournament:create', 'stream:manage')
- `resource` (VARCHAR(50), NOT NULL) - Resource type (e.g., 'tournament', 'stream', 'user')
- `action` (VARCHAR(50), NOT NULL) - Action type (e.g., 'create', 'read', 'update', 'delete', 'manage')
- `description` (TEXT, optional) - Human-readable permission description
- `created_at` (TIMESTAMP, DEFAULT NOW())

**Validation Rules**:
- `name` format: `<resource>:<action>` (e.g., 'tournament:create')
- `name` must be unique
- `resource` and `action` must be non-empty

**Predefined Permissions**:
- `tournament:create` - Create tournaments
- `tournament:read` - View tournaments
- `tournament:update` - Update tournaments
- `tournament:delete` - Delete tournaments
- `stream:manage` - Manage stream settings
- `user:read` - View user profiles
- `user:update` - Update user profiles

**Relationships**:
- Many-to-many with `roles` (via `role_permissions`)

---

### 5. Role Permission Assignment

**Source**: Custom table `public.role_permissions`  
**Description**: Maps roles to their permissions (many-to-many relationship)

**Fields**:
- `role_id` (UUID, NOT NULL, REFERENCES roles(id) ON DELETE CASCADE)
- `permission_id` (UUID, NOT NULL, REFERENCES permissions(id) ON DELETE CASCADE)
- `granted_at` (TIMESTAMP, DEFAULT NOW())
- PRIMARY KEY (`role_id`, `permission_id`)

**Validation Rules**:
- Role can have multiple permissions
- Cannot assign duplicate permission to same role
- Both `role_id` and `permission_id` must reference existing records

**Relationships**:
- Many-to-one with `roles` (via `role_id`)
- Many-to-one with `permissions` (via `permission_id`)

---

### 6. OAuth Provider Link

**Source**: Custom table `public.oauth_provider_links`  
**Description**: Links user accounts to external OAuth providers (Discord, Twitch, Google, Facebook)

**Fields**:
- `id` (UUID, PRIMARY KEY, DEFAULT gen_random_uuid())
- `user_id` (UUID, NOT NULL, REFERENCES auth.users(id) ON DELETE CASCADE)
- `provider` (VARCHAR(20), NOT NULL) - OAuth provider name ('discord', 'twitch', 'google', 'facebook')
- `provider_user_id` (VARCHAR(255), NOT NULL) - User ID from OAuth provider
- `provider_email` (TEXT, optional) - Email from OAuth provider
- `access_token` (TEXT, optional, encrypted) - Encrypted OAuth access token
- `refresh_token` (TEXT, optional, encrypted) - Encrypted OAuth refresh token
- `linked_at` (TIMESTAMP, DEFAULT NOW())
- `last_used_at` (TIMESTAMP, optional) - Last time this provider was used for login
- UNIQUE (`provider`, `provider_user_id`)

**Validation Rules**:
- `provider` must be one of: 'discord', 'twitch', 'google', 'facebook'
- `provider_user_id` must be unique per provider
- User can link multiple providers
- Streamers can only link: 'discord', 'twitch'
- Users can only link: 'google', 'facebook'

**State Transitions**:
- **Created**: OAuth account linked to user
- **Deleted**: OAuth account unlinked (user can still use email/password)

**Relationships**:
- Many-to-one with `users` (via `user_id`)

---

### 7. Password Reset Token

**Source**: Custom table `public.password_reset_tokens` (or Supabase handles this)  
**Description**: Temporary tokens for password recovery

**Note**: Supabase may handle password reset tokens internally. If we need custom logic, we'll create this table.

**Fields** (if custom implementation needed):
- `id` (UUID, PRIMARY KEY, DEFAULT gen_random_uuid())
- `user_id` (UUID, NOT NULL, REFERENCES auth.users(id) ON DELETE CASCADE)
- `token` (VARCHAR(255), UNIQUE, NOT NULL) - Reset token (hashed)
- `expires_at` (TIMESTAMP, NOT NULL) - Token expiration (typically 1 hour)
- `used_at` (TIMESTAMP, optional) - When token was used
- `created_at` (TIMESTAMP, DEFAULT NOW())

**Validation Rules**:
- Token expires after 1 hour (configurable)
- Token can only be used once
- Token must be unique

**State Transitions**:
- **Created**: Password reset requested
- **Used**: Password reset completed
- **Expired**: Token expired (can be cleaned up)

**Relationships**:
- Many-to-one with `users` (via `user_id`)

---

### 8. Authentication Session

**Source**: Supabase manages sessions (JWT-based)  
**Description**: Active user sessions (stateless, stored in JWT)

**Note**: Supabase uses stateless JWT sessions. We don't need a sessions table unless we need session revocation.

**JWT Claims** (in Supabase JWT):
- `sub` (UUID) - User ID
- `email` (STRING) - User email
- `role` (STRING) - User role (from user_metadata)
- `user_metadata` (OBJECT) - Custom user metadata
- `exp` (NUMBER) - Expiration timestamp
- `iat` (NUMBER) - Issued at timestamp
- `iss` (STRING) - Issuer (Supabase project URL)
- `aud` (STRING) - Audience

**If Session Revocation Needed** (future):
- Create `public.revoked_sessions` table
- Store JWT `jti` (JWT ID) when revoked
- Check revocation list during JWT validation

---

## Database Schema

### Tables in Supabase PostgreSQL

```sql
-- Roles table
CREATE TABLE public.roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- User roles (many-to-many)
CREATE TABLE public.user_roles (
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES public.roles(id) ON DELETE CASCADE,
  assigned_at TIMESTAMP DEFAULT NOW(),
  assigned_by UUID REFERENCES auth.users(id),
  PRIMARY KEY (user_id, role_id)
);

-- Permissions table
CREATE TABLE public.permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) UNIQUE NOT NULL,
  resource VARCHAR(50) NOT NULL,
  action VARCHAR(50) NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Role permissions (many-to-many)
CREATE TABLE public.role_permissions (
  role_id UUID NOT NULL REFERENCES public.roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES public.permissions(id) ON DELETE CASCADE,
  granted_at TIMESTAMP DEFAULT NOW(),
  PRIMARY KEY (role_id, permission_id)
);

-- OAuth provider links
CREATE TABLE public.oauth_provider_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  provider VARCHAR(20) NOT NULL CHECK (provider IN ('discord', 'twitch', 'google', 'facebook')),
  provider_user_id VARCHAR(255) NOT NULL,
  provider_email TEXT,
  access_token TEXT, -- encrypted
  refresh_token TEXT, -- encrypted
  linked_at TIMESTAMP DEFAULT NOW(),
  last_used_at TIMESTAMP,
  UNIQUE (provider, provider_user_id)
);

-- Indexes for performance
CREATE INDEX idx_user_roles_user_id ON public.user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON public.user_roles(role_id);
CREATE INDEX idx_role_permissions_role_id ON public.role_permissions(role_id);
CREATE INDEX idx_oauth_provider_links_user_id ON public.oauth_provider_links(user_id);
CREATE INDEX idx_oauth_provider_links_provider ON public.oauth_provider_links(provider);
```

---

## Relationships Diagram

```
auth.users (Supabase)
  ├── 1:N → user_roles
  │         └── N:1 → roles
  │                   └── N:M → role_permissions
  │                             └── N:1 → permissions
  ├── 1:N → oauth_provider_links
  └── 1:N → password_reset_tokens (if custom)
```

---

## Validation Rules Summary

### User Account
- Email format validation (Supabase)
- Password strength requirements (Supabase)
- Email uniqueness (Supabase)
- User type must be 'streamer' or 'user'

### Roles
- Role name must be unique
- Reserved names cannot be used

### User Roles
- User can have multiple roles
- No duplicate role assignments

### Permissions
- Permission name format: `<resource>:<action>`
- Permission name must be unique

### OAuth Links
- Provider must be valid (discord, twitch, google, facebook)
- Provider user ID must be unique per provider
- Streamers can only link discord/twitch
- Users can only link google/facebook

---

## State Transitions

### User Account Lifecycle
1. **Registration** → Unverified account created
2. **Email Verification** → Account verified, can log in
3. **Login** → Active session (JWT issued)
4. **Role Assignment** → Role added to user (JWT refreshed)
5. **Role Removal** → Role removed (JWT refreshed)
6. **Suspension** → Account suspended (via app_metadata)
7. **Deletion** → Account soft-deleted (Supabase)

### OAuth Link Lifecycle
1. **OAuth Flow** → Provider account linked
2. **Login via OAuth** → `last_used_at` updated
3. **Unlink** → OAuth link removed (user can still use email/password)

---

## Data Access Patterns

### Common Queries (for sqlc)

1. **Get user roles**:
   ```sql
   SELECT r.name, r.description
   FROM roles r
   JOIN user_roles ur ON r.id = ur.role_id
   WHERE ur.user_id = :user_id
   ```

2. **Get user permissions** (via roles):
   ```sql
   SELECT DISTINCT p.name, p.resource, p.action
   FROM permissions p
   JOIN role_permissions rp ON p.id = rp.permission_id
   JOIN user_roles ur ON rp.role_id = ur.role_id
   WHERE ur.user_id = :user_id
   ```

3. **Check if user has permission**:
   ```sql
   SELECT EXISTS(
     SELECT 1
     FROM permissions p
     JOIN role_permissions rp ON p.id = rp.permission_id
     JOIN user_roles ur ON rp.role_id = ur.role_id
     WHERE ur.user_id = :user_id AND p.name = :permission_name
   )
   ```

4. **Get OAuth provider link**:
   ```sql
   SELECT * FROM oauth_provider_links
   WHERE user_id = :user_id AND provider = :provider
   ```

---

## Migration Strategy

1. **Initial Migration**: Create roles, permissions, and relationship tables
2. **Seed Data**: Insert predefined roles and permissions
3. **Data Migration**: Assign default roles to existing users (if any)
4. **Index Creation**: Add indexes for performance

---

## Security Considerations

1. **Row Level Security (RLS)**: Supabase RLS policies should protect `auth.users` table
2. **Encryption**: OAuth tokens should be encrypted at rest
3. **Access Control**: Only auth service should write to role/permission tables
4. **Audit Logging**: Log role assignments and permission changes (future)

