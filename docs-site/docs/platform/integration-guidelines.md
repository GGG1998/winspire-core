# Integration Guidelines

## Domain Integration Patterns

### Synchronous Integration (REST)

Use for:
- Direct queries
- Real-time operations
- Client-facing APIs

```yaml
# Example: User service calling Auth service
GET /v1/auth/validate
Authorization: Bearer {token}
```

### Asynchronous Integration (Events)

Use for:
- Domain events
- Background processing
- Eventual consistency

```yaml
# Example: User created event
event: user.created
payload:
  userId: "123"
  email: "user@example.com"
  timestamp: "2024-01-01T00:00:00Z"
```

## Best Practices

1. **Loose Coupling** - Domains should be independent
2. **Contract-First** - Define APIs before implementation
3. **Versioning** - Always version your APIs
4. **Idempotency** - Make operations idempotent
5. **Error Handling** - Use standard error codes
6. **Documentation** - Keep OpenAPI specs up to date

## Anti-Patterns

❌ Direct database access between domains
❌ Tight coupling through shared libraries
❌ Synchronous calls for non-critical operations
❌ Breaking API changes without versioning

