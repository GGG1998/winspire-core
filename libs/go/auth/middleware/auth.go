package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/winspire/winspire-core/libs/go/auth/jwt"
	"github.com/winspire/winspire-core/libs/go/auth/types"
)

// Config holds configuration for the JWT middleware
type Config struct {
	JWTSecret string
	Issuer    string
	Audience  string
}

// ValidateJWTMiddleware creates a Gin middleware that validates JWT tokens
func ValidateJWTMiddleware(cfg Config) gin.HandlerFunc {
	validator := jwt.NewValidator(cfg.JWTSecret, cfg.Issuer, cfg.Audience)

	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Check for Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Extract role from claims
		role := jwt.ExtractRole(claims)
		roles := []types.Role{}
		if role != "" {
			roles = append(roles, types.Role(role))
		}

		// Create user context
		user := &types.UserContext{
			ID:    types.UserID(claims.Sub),
			Email: types.Email(claims.Email),
			Roles: roles,
		}

		// Add user context to request
		c.Request = c.Request.WithContext(types.WithUserContext(c.Request.Context(), user))

		// Store user in Gin context for easy access
		c.Set("user", user)

		c.Next()
	}
}

