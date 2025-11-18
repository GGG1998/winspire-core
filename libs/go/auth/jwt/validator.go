package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Validator validates Supabase JWT tokens
type Validator struct {
	secret     []byte
	issuer     string
	audience   string
}

// NewValidator creates a new JWT validator
func NewValidator(secret, issuer, audience string) *Validator {
	return &Validator{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *Validator) ValidateToken(tokenString string) (*Claims, error) {
	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	}, jwt.WithValidMethods([]string{"HS256", "RS256"}))

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Validate issuer if provided
	if v.issuer != "" && claims.Issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, claims.Issuer)
	}

	// Validate audience if provided
	if v.audience != "" {
		audiences, err := claims.GetAudience()
		if err != nil || len(audiences) == 0 {
			return nil, fmt.Errorf("token missing audience")
		}
		found := false
		for _, aud := range audiences {
			if aud == v.audience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid audience: expected %s", v.audience)
		}
	}

	return claims, nil
}

