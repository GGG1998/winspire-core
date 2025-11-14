# User Management Policies

## Business Rules

### User Creation

1. Email must be unique in the system
2. Email must be valid format
3. Password must meet security requirements:
   - Minimum 8 characters
   - At least one uppercase letter
   - At least one number
   - At least one special character
4. New users start with `inactive` status
5. Email verification required before activation

### User Updates

1. Email changes require re-verification
2. Only user themselves or admins can update profile
3. Role changes require admin permission
4. Status transitions must follow state machine

### User Deletion

1. Soft delete by default (status = 'deleted')
2. Hard delete only after 90 days
3. Admin approval required for hard delete
4. Cascade: anonymize personal data, keep audit logs

## State Machine

```
inactive -> active -> suspended -> deleted
         ↑           ↓
         ←-----------←
```

## Validation Rules

- **Email**: RFC 5322 compliant
- **Phone**: E.164 format (optional)
- **Name**: 1-100 characters, no special characters
- **Bio**: Max 500 characters

