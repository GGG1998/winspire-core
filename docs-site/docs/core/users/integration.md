# Users Integration Guide

## Overview

This guide describes how other domains can integrate with the Users subdomain.

## Integration Methods

### REST API

For synchronous operations, use the Users REST API:

```bash
GET /v1/users/{userId}
Authorization: Bearer {token}
```

See [REST API Documentation](./api/openapi.yaml) for full specification.

### Events

Subscribe to user-related events for asynchronous updates:

#### User Created Event

Published when a new user is created.

```yaml
event: user.created
payload:
  userId: string (UUID)
  email: string
  createdAt: string (ISO 8601)
```

#### User Updated Event

Published when user data changes.

```yaml
event: user.updated
payload:
  userId: string (UUID)
  changes:
    field: oldValue -> newValue
  updatedAt: string (ISO 8601)
```

#### User Deleted Event

Published when a user is soft-deleted.

```yaml
event: user.deleted
payload:
  userId: string (UUID)
  deletedAt: string (ISO 8601)
```

## Authentication

All API calls require authentication using JWT Bearer tokens:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

Obtain tokens from the Auth subdomain.

## Rate Limiting

- **Standard tier**: 100 requests/minute
- **Premium tier**: 1000 requests/minute

Rate limit headers:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

## Error Handling

Use standard error codes from [Platform Error Codes](/platform/shared-schemas/error-codes.yaml).

Example error response:
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User not found",
    "requestId": "req_123456789"
  }
}
```

## Best Practices

1. **Cache user data** - Reduce API calls by caching
2. **Subscribe to events** - Stay updated with async events
3. **Handle rate limits** - Implement exponential backoff
4. **Use shared schemas** - Reference `/platform/shared-schemas/user-common.yaml`
5. **Idempotency** - Use idempotency keys for mutations

## Support

For integration support, consult:
- [Platform Integration Guidelines](/platform/integration-guidelines)
- [API Standards](/platform/api-standards)

