package services

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	supabaseclient "github.com/winspire/winspire-core/services/auth/internal/supabase"
	"github.com/supabase-community/gotrue-go/types"
)

// RegistrationService handles user registration
type RegistrationService struct {
	supabaseClient *supabaseclient.Client
	logger         *logger.Logger
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(supabaseClient *supabaseclient.Client, appLogger *logger.Logger) *RegistrationService {
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
	
	// Use Supabase Auth Signup with email and password
	signupReq := types.SignupRequest{
		Email:    input.Email,
		Password: input.Password,
		Data:     userMetadata,
	}
	
	signupResponse, err := client.Auth.Signup(signupReq)
	if err != nil {
		s.logger.Error("Registration failed for email %s: %v", input.Email, err)
		// Check for specific Supabase errors
		if strings.Contains(err.Error(), "already registered") || 
		   strings.Contains(err.Error(), "User already registered") ||
		   strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// SignupResponse can contain either User (if autoconfirm is off) or Session (if autoconfirm is on)
	var requiresVerification bool
	var session *Session
	var userID string
	var email string
	
	// Check if we have a Session (autoconfirm is on)
	if signupResponse.Session.AccessToken != "" {
		session = &Session{
			AccessToken:  signupResponse.Session.AccessToken,
			RefreshToken: signupResponse.Session.RefreshToken,
			ExpiresIn:    signupResponse.Session.ExpiresIn,
		}
		userID = signupResponse.Session.User.ID.String()
		email = signupResponse.Session.User.Email
		requiresVerification = signupResponse.Session.User.EmailConfirmedAt == nil
	} else if signupResponse.User.ID != (uuid.UUID{}) {
		// Only User (autoconfirm is off)
		userID = signupResponse.User.ID.String()
		email = signupResponse.User.Email
		requiresVerification = signupResponse.User.EmailConfirmedAt == nil
	} else {
		userID = ""
		email = input.Email
		requiresVerification = true
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

