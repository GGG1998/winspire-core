package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CachedValidator wraps a Validator with Redis caching for improved performance
type CachedValidator struct {
	validator *Validator
	redis     *redis.Client
	keyPrefix string
}

// NewCachedValidator creates a new cached JWT validator
func NewCachedValidator(validator *Validator, redisClient *redis.Client) *CachedValidator {
	return &CachedValidator{
		validator: validator,
		redis:     redisClient,
		keyPrefix: "jwt:v1:",
	}
}

// ValidateToken validates a JWT token with Redis caching
// Cache key: jwt:v1:{sha256(token)}
// Cache TTL: token expiry time - now (or 1 hour if no expiry)
func (cv *CachedValidator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Generate cache key from token hash
	cacheKey := cv.cacheKey(tokenString)

	// Try to get from cache
	if cv.redis != nil {
		cached, err := cv.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache hit - deserialize claims
			var claims Claims
			if err := json.Unmarshal([]byte(cached), &claims); err == nil {
				// Verify expiration hasn't passed (defense in depth)
				if claims.ExpiresAt == nil || claims.ExpiresAt.Time.After(time.Now()) {
					return &claims, nil
				}
				// Token expired, delete from cache
				cv.redis.Del(ctx, cacheKey)
			}
		}
	}

	// Cache miss or error - validate token
	claims, err := cv.validator.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Cache the validated claims
	if cv.redis != nil {
		go cv.cacheToken(cacheKey, claims)
	}

	return claims, nil
}

// ValidateTokenSync is a synchronous version without context (for backward compatibility)
func (cv *CachedValidator) ValidateTokenSync(tokenString string) (*Claims, error) {
	return cv.ValidateToken(context.Background(), tokenString)
}

// cacheToken stores the validated claims in Redis
func (cv *CachedValidator) cacheToken(cacheKey string, claims *Claims) {
	ctx := context.Background()

	// Serialize claims
	data, err := json.Marshal(claims)
	if err != nil {
		return
	}

	// Calculate TTL based on token expiry
	ttl := 1 * time.Hour // default
	if claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 && remaining < 24*time.Hour {
			ttl = remaining
		}
	}

	// Store in Redis (ignore errors in background goroutine)
	cv.redis.Set(ctx, cacheKey, data, ttl)
}

// cacheKey generates a cache key from the token string
func (cv *CachedValidator) cacheKey(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return cv.keyPrefix + hex.EncodeToString(hash[:])
}

// InvalidateToken removes a token from the cache (useful for logout)
func (cv *CachedValidator) InvalidateToken(ctx context.Context, tokenString string) error {
	if cv.redis == nil {
		return nil
	}
	cacheKey := cv.cacheKey(tokenString)
	return cv.redis.Del(ctx, cacheKey).Err()
}

// Stats returns cache statistics (for monitoring)
func (cv *CachedValidator) Stats(ctx context.Context) (map[string]interface{}, error) {
	if cv.redis == nil {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	// Get Redis info
	info, err := cv.redis.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis stats: %w", err)
	}

	return map[string]interface{}{
		"enabled":    true,
		"redis_info": info,
	}, nil
}

// Close closes the Redis connection (if owned by this validator)
func (cv *CachedValidator) Close() error {
	if cv.redis != nil {
		return cv.redis.Close()
	}
	return nil
}
