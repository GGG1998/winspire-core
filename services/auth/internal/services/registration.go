package services

import (
	"fmt"
	"strings"

	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/supabase"
	"github.com/supabase-community/supabase-go"
)

// RegistrationService handles user registration
type RegistrationService struct {
	supabaseClient *supabase.Client
	logger         *logger.Logger
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(supabaseClient *supabase.Client, appLogger *logger.Logger) *RegistrationService {
	return &RegistrationService{
		supabaseClient: supabaseClient,
		logger:         appLogger,
	}
}

// RegisterUserInput represents registration input
type RegisterUserInput struct {
	Email    string
	Password string
	UserType string // "streamer" or "user"
}

// RegisterUserOutput represents registration output
type RegisterUserOutput struct {
	UserID    string
	Email     string
	UserType  string
	Session   *Session
	RequiresVerification bool
}

// Session represents authentication session (shared across services)
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// RegisterUser registers a new user with Supabase
func (s *RegistrationService) RegisterUser(input RegisterUserInput) (*RegisterUserOutput, error) {
	s.logger.Info("Registration attempt for email: %s, userType: %s", input.Email, input.UserType)

	// Validate user type
	if input.UserType != "streamer" && input.UserType != "user" {
		return nil, fmt.Errorf("invalid user type: %s (must be 'streamer' or 'user')", input.UserType)
	}

	// Prepare user metadata with role
	// Role is set based on user_type (streamer -> 'streamer', user -> 'user')
	role := input.UserType
	userMetadata := map[string]interface{}{
		"role":       role,
		"user_type":  input.UserType,
	}

	// Call Supabase Auth SignUp
	client := s.supabaseClient.GetClient()
	
	// Use Supabase Auth SignUp with email and password
	authResponse, err := client.Auth.SignUp(client.Context, supabase.UserCredentials{
		Email:    input.Email,
		Password: input.Password,
		Data:     userMetadata,
	})
	if err != nil {
		s.logger.Error("Registration failed for email %s: %v", input.Email, err)
		// Check for specific Supabase errors
		if strings.Contains(err.Error(), "already registered") || 
		   strings.Contains(err.Error(), "User already registered") {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Check if email verification is required
	requiresVerification := authResponse.User == nil || authResponse.User.EmailConfirmedAt == nil

	var session *Session
	if authResponse.Session != nil {
		session = &Session{
			AccessToken:  authResponse.Session.AccessToken,
			RefreshToken: authResponse.Session.RefreshToken,
			ExpiresIn:    authResponse.Session.ExpiresIn,
		}
	}

	userID := ""
	email := input.Email
	if authResponse.User != nil {
		userID = authResponse.User.ID
		email = authResponse.User.Email
	}

	s.logger.Info("User registered successfully: %s (ID: %s, Verified: %v)", 
		email, userID, !requiresVerification)

	return &RegisterUserOutput{
		UserID:              userID,
		Email:               email,
		UserType:            input.UserType,
		Session:             session,
		RequiresVerification: requiresVerification,
	}, nil
}

