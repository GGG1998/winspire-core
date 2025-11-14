# Users Data Model

## Entities

### User

Main user entity representing a system user.

```yaml
User:
  id: UUID (Primary Key)
  email: String (Unique, Not Null)
  firstName: String
  lastName: String
  status: UserStatus (Enum)
  role: UserRole (Enum)
  createdAt: DateTime
  updatedAt: DateTime
  lastLoginAt: DateTime (Nullable)
```

### UserProfile

Extended user profile information.

```yaml
UserProfile:
  userId: UUID (Foreign Key -> User)
  bio: Text (Nullable)
  avatarUrl: String (Nullable)
  phoneNumber: String (Nullable)
  timezone: String
  language: String
  preferences: JSON
```

## Relationships

- User 1:1 UserProfile
- User references shared schemas from `/platform/shared-schemas/user-common.yaml`

## Constraints

- Email must be unique
- Email must be validated before account activation
- Soft delete: status = 'deleted' instead of physical deletion

