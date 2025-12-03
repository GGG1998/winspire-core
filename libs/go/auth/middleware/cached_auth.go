package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/winspire/winspire-core/libs/go/auth/jwt"
	"github.com/winspire/winspire-core/libs/go/auth/types"
)

// ValidateJWTMiddlewareWithCache creates a Gin middleware that validates JWT tokens with Redis caching
func ValidateJWTMiddlewareWithCache(cfg Config, redisClient *redis.Client) gin.HandlerFunc {
	// Create base validator
	validator := jwt.NewValidator(cfg.JWTSecret, cfg.Issuer, cfg.Audience)

	// Wrap with caching
	cachedValidator := jwt.NewCachedValidator(validator, redisClient)

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

		// Validate token with cache
		ctx := c.Request.Context()
		claims, err := cachedValidator.ValidateToken(ctx, tokenString)
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

		// Extract nickname and user_type from claims
		nickname := jwt.ExtractNickname(claims)
		userType := jwt.ExtractUserType(claims)

		// Create user context
		user := &types.UserContext{
			ID:       types.UserID(claims.Sub),
			Email:    types.Email(claims.Email),
			Nickname: nickname,
			UserType: userType,
			Roles:    roles,
		}

		// Add user context to request
		c.Request = c.Request.WithContext(types.WithUserContext(c.Request.Context(), user))

		// Store user in Gin context for easy access
		c.Set("user", user)

		c.Next()
	}
}

// LogoutHandler invalidates a JWT token from the cache
func LogoutHandler(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]

				// Create a cached validator just to invalidate the token
				validator := jwt.NewValidator("", "", "")
				cachedValidator := jwt.NewCachedValidator(validator, redisClient)

				// Invalidate the token in cache
				_ = cachedValidator.InvalidateToken(context.Background(), tokenString)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	}
}
