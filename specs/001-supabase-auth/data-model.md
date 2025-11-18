# Data Model: Supabase Authentication Integration

**Date**: 2025-11-18  
**Feature**: Supabase Authentication Integration

## Overview

This document defines the data model for the authentication system. We rely entirely on Supabase's built-in structures for authentication and identity management. No custom database tables are needed.

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
  - `role` (STRING) - Primary role stored in JWT claims (e.g., 'streamer', 'user', 'admin')
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
- One-to-many with `auth.identities` (OAuth provider links)

---

### 2. OAuth Identity (Supabase Auth)

**Source**: Supabase `auth.identities` table (managed by Supabase)  
**Description**: OAuth provider links managed by Supabase. Supports automatic and manual identity linking.

**Fields** (from Supabase):
- `id` (UUID, PRIMARY KEY) - Identity unique identifier
- `user_id` (UUID, NOT NULL, REFERENCES auth.users(id)) - Linked user account
- `provider` (TEXT, NOT NULL) - OAuth provider name ('discord', 'twitch', 'google', 'facebook', etc.)
- `identity_data` (JSONB) - Provider-specific data (provider user ID, email, etc.)
- `created_at` (TIMESTAMP) - When identity was linked
- `updated_at` (TIMESTAMP) - Last update timestamp

**Validation Rules**:
- Provider must be a valid OAuth provider configured in Supabase
- Automatic linking: Supabase links identities with the same verified email
- Manual linking: Users can link identities via `linkIdentity()` API
- Streamers can only link: 'discord', 'twitch'
- Users can only link: 'google', 'facebook'

**State Transitions**:
- **Created**: OAuth account linked to user (automatic or manual)
- **Deleted**: OAuth account unlinked (user can still use email/password)

**Relationships**:
- Many-to-one with `auth.users` (via `user_id`)

**API Usage**:
- Get identities: `supabaseClient.Auth.GetUserIdentities(userID)`
- Link identity: `supabaseClient.Auth.LinkIdentity(provider, token)`
- Unlink identity: `supabaseClient.Auth.UnlinkIdentity(identityID)`

See: [Supabase Identity Linking Documentation](https://supabase.com/docs/guides/auth/auth-identity-linking)

---

### 3. Authentication Session

**Source**: Supabase manages sessions (JWT-based)  
**Description**: Active user sessions (stateless, stored in JWT)

**Note**: Supabase uses stateless JWT sessions. No sessions table needed.

**JWT Claims** (in Supabase JWT):
- `sub` (UUID) - User ID
- `email` (STRING) - User email
- `role` (STRING) - User role (from user_metadata.role)
- `user_metadata` (OBJECT) - Custom user metadata
  - `role` (STRING) - User role
  - `user_type` (STRING) - 'streamer' or 'user'
  - `display_name` (STRING, optional)
- `app_metadata` (OBJECT) - Application metadata
- `exp` (NUMBER) - Expiration timestamp
- `iat` (NUMBER) - Issued at timestamp
- `iss` (STRING) - Issuer (Supabase project URL)
- `aud` (STRING) - Audience

**If Session Revocation Needed** (future):
- Create `public.revoked_sessions` table
- Store JWT `jti` (JWT ID) when revoked
- Check revocation list during JWT validation

---

### 4. Password Reset Token

**Source**: Supabase handles internally  
**Description**: Temporary tokens for password recovery

**Note**: Supabase handles password reset tokens internally. No custom table needed.

**Usage**:
- Request reset: `supabaseClient.Auth.ResetPasswordForEmail(email)`
- Confirm reset: `supabaseClient.Auth.UpdateUser(password)`

---

## Database Schema

### Supabase Built-in Tables

We rely entirely on Supabase's built-in authentication tables:

1. **`auth.users`** - User accounts
   - Managed by Supabase
   - Contains email, password, metadata
   - Roles stored in `user_metadata.role`

2. **`auth.identities`** - OAuth provider links
   - Managed by Supabase
   - Automatically tracks all OAuth providers per user
   - Supports automatic linking by email
   - Supports manual linking via API

3. **Password Reset Tokens** - Handled internally by Supabase

4. **Sessions** - JWT-based, stateless (no table)

### No Custom Tables Required

We do not create any custom tables. All authentication and identity management is handled by Supabase's built-in structures.

---

## Relationships Diagram

```
auth.users (Supabase)
  └── 1:N → auth.identities (OAuth provider links)
```

---

## Validation Rules Summary

### User Account
- Email format validation (Supabase)
- Password strength requirements (Supabase)
- Email uniqueness (Supabase)
- User type must be 'streamer' or 'user'
- Role stored in `user_metadata.role` (e.g., 'streamer', 'user', 'admin')

### OAuth Identities
- Provider must be valid (discord, twitch, google, facebook)
- Automatic linking by verified email (Supabase handles)
- Manual linking via API (Supabase handles)
- Streamers can only link: 'discord', 'twitch'
- Users can only link: 'google', 'facebook'

---

## State Transitions

### User Account Lifecycle
1. **Registration** → Unverified account created
2. **Email Verification** → Account verified, can log in
3. **Login** → Active session (JWT issued with role in claims)
4. **Role Update** → Update `user_metadata.role`, new JWT issued
5. **Suspension** → Account suspended (via app_metadata)
6. **Deletion** → Account soft-deleted (Supabase)

### OAuth Identity Lifecycle
1. **OAuth Flow** → Identity automatically or manually linked to user
2. **Login via OAuth** → Identity used for authentication
3. **Unlink** → Identity removed (user can still use email/password)

---

## Data Access Patterns

### Using Supabase Go Client

**Get User**:
```go
user, err := supabaseClient.Auth.GetUser(accessToken)
```

**Get User Identities (OAuth providers)**:
```go
identities, err := supabaseClient.Auth.GetUserIdentities(userID)
```

**Link OAuth Identity**:
```go
err := supabaseClient.Auth.LinkIdentity(provider, token)
```

**Unlink OAuth Identity**:
```go
err := supabaseClient.Auth.UnlinkIdentity(identityID)
```

**Update User Role**:
```go
// Update user_metadata.role
user, err := supabaseClient.Auth.UpdateUser(map[string]interface{}{
    "user_metadata": map[string]interface{}{
        "role": "streamer",
    },
})
```

---

## Role Management

### Role Storage

Roles are stored in `user_metadata.role` field:
- Set during registration based on `user_type`
- Updated via Supabase Admin API or service role key
- Included in JWT claims automatically
- Read from JWT in middleware (no database query needed)

### Supported Roles

- `streamer` - Users who stream content
- `user` - Regular platform users  
- `admin` - System administrators (future)

### Role Assignment

Roles are assigned by updating `user_metadata.role`:
```go
// Using service role key
supabaseClient.Auth.Admin.UpdateUserByID(userID, map[string]interface{}{
    "user_metadata": map[string]interface{}{
        "role": "streamer",
    },
})
```

---

## Security Considerations

1. **Row Level Security (RLS)**: Supabase RLS policies protect `auth.users` and `auth.identities` tables
2. **JWT Validation**: All services validate JWTs using Supabase JWT secret
3. **Role-based Access**: Roles stored in JWT claims, validated in middleware
4. **Identity Linking**: Supabase handles security for automatic and manual linking
5. **Password Reset**: Handled securely by Supabase

---

## Migration Strategy

**No migrations needed** - We rely entirely on Supabase's built-in structures.

If custom tables are needed in the future (e.g., for advanced RBAC), they can be added via migrations at that time.

---

## References

- [Supabase Auth Documentation](https://supabase.com/docs/guides/auth)
- [Supabase Identity Linking](https://supabase.com/docs/guides/auth/auth-identity-linking)
- [Supabase JWT Claims](https://supabase.com/docs/guides/auth/jwts)
- [Supabase User Management](https://supabase.com/docs/guides/auth/managing-user-data)
