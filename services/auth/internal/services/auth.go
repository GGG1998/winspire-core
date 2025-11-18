package services

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	supabaseclient "github.com/winspire/winspire-core/services/auth/internal/supabase"
)

// AuthService handles user authentication (login, refresh)
type AuthService struct {
	supabaseClient *supabaseclient.Client
	logger         *logger.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(supabaseClient *supabaseclient.Client, appLogger *logger.Logger) *AuthService {
	return &AuthService{
		supabaseClient: supabaseClient,
		logger:         appLogger,
	}
}

// LoginUserInput represents login input
type LoginUserInput struct {
	Email    string
	Password string
}

// LoginUserOutput represents login output
type LoginUserOutput struct {
	UserID   string
	Email    string
	UserType string
	Role     string
	Session  *Session
}

// LoginUser authenticates a user with email and password
func (s *AuthService) LoginUser(input LoginUserInput) (*LoginUserOutput, error) {
	s.logger.Info("Login attempt for email: %s", input.Email)

	// Call Supabase Auth SignIn
	client := s.supabaseClient.GetClient()
	
	tokenResponse, err := client.Auth.SignInWithEmailPassword(input.Email, input.Password)
	if err != nil {
		s.logger.Error("Login failed for email %s: %v", input.Email, err)
		
		// Check for specific error types
		if strings.Contains(err.Error(), "Invalid login credentials") ||
		   strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "credentials") {
			return nil, errors.NewError(errors.ErrorCodeUnauthorized, "Invalid email or password")
		}
		
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	// Check if email is verified
	if tokenResponse.User.ID == (uuid.UUID{}) {
		return nil, errors.NewError(errors.ErrorCodeInternal, "User not found in response")
	}

	if tokenResponse.User.EmailConfirmedAt == nil {
		s.logger.Warn("Login attempt with unverified email: %s", input.Email)
		return nil, errors.NewError(errors.ErrorCodeEmailNotVerified, "Email address not verified. Please check your email and verify your account.")
	}

	// Extract user metadata
	userType := "user"
	role := "user"
	if tokenResponse.User.UserMetadata != nil {
		if ut, ok := tokenResponse.User.UserMetadata["user_type"].(string); ok {
			userType = ut
		}
		if r, ok := tokenResponse.User.UserMetadata["role"].(string); ok {
			role = r
		}
	}

	// Build session
	session := &Session{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresIn:    tokenResponse.ExpiresIn,
	}

	s.logger.Info("User logged in successfully: %s (ID: %s, Role: %s)", 
		tokenResponse.User.Email, tokenResponse.User.ID, role)

	return &LoginUserOutput{
		UserID:   tokenResponse.User.ID.String(),
		Email:    tokenResponse.User.Email,
		UserType: userType,
		Role:     role,
		Session:  session,
	}, nil
}

// RefreshTokenInput represents refresh token input
type RefreshTokenInput struct {
	RefreshToken string
}

// RefreshTokenOutput represents refresh token output
type RefreshTokenOutput struct {
	Session *Session
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(input RefreshTokenInput) (*RefreshTokenOutput, error) {
	s.logger.Info("Token refresh attempt")

	client := s.supabaseClient.GetClient()
	
	// Use Supabase Auth RefreshToken
	tokenResponse, err := client.Auth.RefreshToken(input.RefreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed: %v", err)
		
		if strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "expired") ||
		   strings.Contains(err.Error(), "revoked") {
			return nil, errors.NewError(errors.ErrorCodeUnauthorized, "Invalid or expired refresh token")
		}
		
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return nil, errors.NewError(errors.ErrorCodeInternal, "No session returned from token refresh")
	}

	session := &Session{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresIn:    tokenResponse.ExpiresIn,
	}

	s.logger.Info("Token refreshed successfully")

	return &RefreshTokenOutput{
		Session: session,
	}, nil
}

// LogoutUserInput represents logout input
type LogoutUserInput struct {
	AccessToken string
}

// LogoutUser logs out a user and invalidates their session
func (s *AuthService) LogoutUser(input LogoutUserInput) error {
	s.logger.Info("Logout attempt")

	client := s.supabaseClient.GetClient()
	
	// Call Supabase Auth Logout
	// Note: Supabase uses stateless JWT, so logout is mainly client-side
	// We can call Logout to invalidate refresh tokens on the server
	// Logout requires authentication token, so we need to set it first
	authenticatedClient := client.Auth.WithToken(input.AccessToken)
	err := authenticatedClient.Logout()
	if err != nil {
		s.logger.Error("Logout failed: %v", err)
		
		// Even if SignOut fails, we consider logout successful from client perspective
		// The JWT will expire naturally, and refresh token invalidation is best-effort
		if strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "expired") {
			// Token already invalid/expired - logout is effectively successful
			s.logger.Info("Token already invalid/expired - logout successful")
			return nil
		}
		
		// Log error but don't fail logout (stateless JWT means client can just discard token)
		s.logger.Warn("Logout service error (non-critical): %v", err)
	}

	s.logger.Info("User logged out successfully")
	return nil
}

