package types

import (
	"context"
)

// UserID represents a user identifier
type UserID string

// Email represents a user email
type Email string

// Role represents a user role
type Role string

// UserContext holds user information extracted from JWT
type UserContext struct {
	ID       UserID
	Email    Email
	Nickname string
	UserType string
	Roles    []Role
}

type contextKey string

const userContextKey contextKey = "user"

// WithUserContext adds user context to the request context
func WithUserContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext extracts user context from the request context
func UserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	return user, ok
}
