package jwt

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims from Supabase
type Claims struct {
	Sub          string                 `json:"sub"`
	Email        string                 `json:"email"`
	Role         string                 `json:"role"`
	Nickname     string                 `json:"nickname"`
	UserType     string                 `json:"user_type"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	jwt.RegisteredClaims
}

// ParseToken parses a JWT token string and returns the claims
func ParseToken(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractRole extracts the role from claims
func ExtractRole(claims *Claims) string {
	if claims.Role != "" {
		return claims.Role
	}
	
	// Try to get role from user_metadata
	if role, ok := claims.UserMetadata["role"].(string); ok {
		return role
	}

	return ""
}

// ExtractUserType extracts the user type from claims
func ExtractUserType(claims *Claims) string {
	if claims.UserType != "" {
		return claims.UserType
	}
	
	// Fallback to user_metadata if not in top-level claims
	if userType, ok := claims.UserMetadata["user_type"].(string); ok {
		return userType
	}
	return ""
}

// ExtractNickname extracts nickname from claims
func ExtractNickname(claims *Claims) string {
	if claims.Nickname != "" {
		return claims.Nickname
	}
	
	// Fallback to user_metadata if not in top-level claims
	if nick, ok := claims.UserMetadata["nickname"].(string); ok {
		return nick
	}
	return ""
}

// ToJSON converts claims to JSON for debugging
func (c *Claims) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

