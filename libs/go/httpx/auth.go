package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/winspire/winspire-core/libs/go/auth/types"
)

// RequireAuth is a middleware that ensures a user is authenticated
// This expects that JWT validation has already been done by an upstream middleware
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authentication",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if the authenticated user has the specified role
func RequireRole(role types.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		userCtx, ok := user.(*types.UserContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user context",
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, r := range userCtx.Roles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "insufficient permissions",
				"required_role": role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if the authenticated user has any of the specified roles
func RequireAnyRole(roles ...types.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		userCtx, ok := user.(*types.UserContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user context",
			})
			c.Abort()
			return
		}

		hasAnyRole := false
		for _, userRole := range userCtx.Roles {
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasAnyRole = true
					break
				}
			}
			if hasAnyRole {
				break
			}
		}

		if !hasAnyRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "insufficient permissions",
				"required_roles": roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminRole is a convenience wrapper for RequireRole("admin")
func RequireAdminRole() gin.HandlerFunc {
	return RequireRole("admin")
}

// GetUser extracts the user context from the Gin context
func GetUser(c *gin.Context) (*types.UserContext, bool) {
	user, ok := c.Get("user")
	if !ok {
		return nil, false
	}

	userCtx, ok := user.(*types.UserContext)
	return userCtx, ok
}

// MustGetUser extracts the user context from the Gin context and panics if not found
// Use this only in handlers that are protected by RequireAuth middleware
func MustGetUser(c *gin.Context) *types.UserContext {
	user, ok := GetUser(c)
	if !ok {
		panic("user context not found - did you forget to add RequireAuth middleware?")
	}
	return user
}
