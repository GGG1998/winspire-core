package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/supabase"
	"github.com/supabase-community/supabase-go"
)

// OAuthService handles OAuth authentication flows
type OAuthService struct {
	supabaseClient *supabase.Client
	supabaseURL    string
	logger         *logger.Logger
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(supabaseClient *supabase.Client, supabaseURL string, appLogger *logger.Logger) *OAuthService {
	return &OAuthService{
		supabaseClient: supabaseClient,
		supabaseURL:    supabaseURL,
		logger:         appLogger,
	}
}

// ValidateProvider validates that the provider is allowed for the user type
func ValidateProvider(provider, userType string) error {
	provider = strings.ToLower(provider)
	userType = strings.ToLower(userType)

	// Streamers can only use Discord and Twitch
	if userType == "streamer" {
		if provider != "discord" && provider != "twitch" {
			return errors.NewError(errors.ErrorCodeValidation, 
				fmt.Sprintf("Provider '%s' is not allowed for streamers. Use 'discord' or 'twitch'.", provider))
		}
	}

	// Users can only use Google and Facebook
	if userType == "user" {
		if provider != "google" && provider != "facebook" {
			return errors.NewError(errors.ErrorCodeValidation,
				fmt.Sprintf("Provider '%s' is not allowed for users. Use 'google' or 'facebook'.", provider))
		}
	}

	// Validate provider is one of the supported ones
	supportedProviders := []string{"discord", "twitch", "google", "facebook"}
	isSupported := false
	for _, p := range supportedProviders {
		if provider == p {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return errors.NewError(errors.ErrorCodeValidation,
			fmt.Sprintf("Unsupported OAuth provider: %s. Supported providers: discord, twitch, google, facebook", provider))
	}

	return nil
}

// InitiateOAuthFlowInput represents OAuth initiation input
type InitiateOAuthFlowInput struct {
	Provider   string
	UserType   string
	RedirectURI string
}

// InitiateOAuthFlowOutput represents OAuth initiation output
type InitiateOAuthFlowOutput struct {
	AuthURL string
}

// InitiateOAuthFlow generates OAuth URL for the provider
func (s *OAuthService) InitiateOAuthFlow(input InitiateOAuthFlowInput) (*InitiateOAuthFlowOutput, error) {
	s.logger.Info("OAuth flow initiation for provider: %s, userType: %s", input.Provider, input.UserType)

	// Validate provider
	if err := ValidateProvider(input.Provider, input.UserType); err != nil {
		return nil, err
	}

	// Validate redirect URI
	if input.RedirectURI == "" {
		return nil, errors.NewError(errors.ErrorCodeValidation, "Redirect URI is required")
	}

	// Build OAuth URL using Supabase Auth
	// Supabase handles OAuth flow, we just need to redirect to the right URL
	// Format: {supabaseURL}/auth/v1/authorize?provider={provider}&redirect_to={redirectURI}
	
	authURL, err := url.Parse(fmt.Sprintf("%s/auth/v1/authorize", s.supabaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth URL: %w", err)
	}

	q := authURL.Query()
	q.Set("provider", strings.ToLower(input.Provider))
	q.Set("redirect_to", input.RedirectURI)
	authURL.RawQuery = q.Encode()

	s.logger.Info("OAuth URL generated for provider: %s", input.Provider)

	return &InitiateOAuthFlowOutput{
		AuthURL: authURL.String(),
	}, nil
}

// HandleOAuthCallbackInput represents OAuth callback input
type HandleOAuthCallbackInput struct {
	Provider   string
	Code       string
	State      string
}

// HandleOAuthCallbackOutput represents OAuth callback output
type HandleOAuthCallbackOutput struct {
	UserID   string
	Email    string
	UserType string
	Role     string
	Session  *Session
}

// HandleOAuthCallback processes OAuth callback and exchanges code for tokens
func (s *OAuthService) HandleOAuthCallback(input HandleOAuthCallbackInput) (*HandleOAuthCallbackOutput, error) {
	s.logger.Info("OAuth callback processing for provider: %s", input.Provider)

	// Validate code is present
	if input.Code == "" {
		return nil, errors.NewError(errors.ErrorCodeValidation, "OAuth authorization code is required")
	}

	client := s.supabaseClient.GetClient()

	// Supabase OAuth callback flow:
	// 1. User is redirected to Supabase callback URL with code
	// 2. Supabase exchanges code for tokens and creates/links user
	// 3. Supabase redirects to our callback URL with session
	// 4. We need to extract the session from the callback
	
	// Note: The actual OAuth flow is handled by Supabase
	// The callback URL receives the session token or user info
	// We may need to use ExchangeCodeForSession or similar method
	
	// For now, we'll use a simplified approach
	// In production, Supabase redirects to callback with access_token in URL or session
	// We need to handle the session extraction based on Supabase's callback format
	
	// Try to exchange code for session
	// This depends on Supabase Go client API - may need adjustment
	sessionResponse, err := client.Auth.ExchangeCodeForSession(client.Context, input.Code)
	if err != nil {
		s.logger.Error("OAuth callback processing failed: %v", err)
		
		if strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "expired") ||
		   strings.Contains(err.Error(), "code") {
			return nil, errors.NewError(errors.ErrorCodeBadRequest, "Invalid or expired OAuth authorization code")
		}
		
		return nil, fmt.Errorf("failed to process OAuth callback: %w", err)
	}

	// Extract user and session from response
	if sessionResponse.User == nil {
		return nil, errors.NewError(errors.ErrorCodeInternal, "User not found in OAuth response")
	}

	// Extract user information from session response
	authResponse := sessionResponse

	// Extract user metadata
	userType := "user"
	role := "user"
	if authResponse.User.UserMetadata != nil {
		if ut, ok := authResponse.User.UserMetadata["user_type"].(string); ok {
			userType = ut
		}
		if r, ok := authResponse.User.UserMetadata["role"].(string); ok {
			role = r
		}
	}

	// Build session
	var session *Session
	if authResponse.Session != nil {
		session = &Session{
			AccessToken:  authResponse.Session.AccessToken,
			RefreshToken: authResponse.Session.RefreshToken,
			ExpiresIn:    authResponse.Session.ExpiresIn,
		}
	} else {
		return nil, errors.NewError(errors.ErrorCodeInternal, "No session returned from OAuth authentication")
	}

	s.logger.Info("OAuth authentication successful for provider: %s, user: %s", 
		input.Provider, authResponse.User.Email)

	return &HandleOAuthCallbackOutput{
		UserID:   authResponse.User.ID,
		Email:    authResponse.User.Email,
		UserType: userType,
		Role:     role,
		Session:  session,
	}, nil
}

// GetUserIdentitiesInput represents input for getting user identities
type GetUserIdentitiesInput struct {
	UserID string
}

// GetUserIdentitiesOutput represents output for getting user identities
type GetUserIdentitiesOutput struct {
	Identities []OAuthIdentity
}

// OAuthIdentity represents an OAuth provider identity
type OAuthIdentity struct {
	ID       string
	Provider string
	UserID   string
}

// GetUserIdentities retrieves all OAuth identities linked to a user
func (s *OAuthService) GetUserIdentities(input GetUserIdentitiesInput) (*GetUserIdentitiesOutput, error) {
	s.logger.Info("Getting OAuth identities for user: %s", input.UserID)

	client := s.supabaseClient.GetClient()
	
	// Use Supabase Auth GetUserIdentities
	identities, err := client.Auth.GetUserIdentities(client.Context, input.UserID)
	if err != nil {
		s.logger.Error("Failed to get user identities: %v", err)
		return nil, fmt.Errorf("failed to get user identities: %w", err)
	}

	result := make([]OAuthIdentity, 0, len(identities))
	for _, identity := range identities {
		result = append(result, OAuthIdentity{
			ID:       identity.ID,
			Provider: identity.Provider,
			UserID:   identity.UserID,
		})
	}

	return &GetUserIdentitiesOutput{
		Identities: result,
	}, nil
}

// UnlinkIdentityInput represents input for unlinking an identity
type UnlinkIdentityInput struct {
	IdentityID string
}

// UnlinkIdentity unlinks an OAuth identity from a user
func (s *OAuthService) UnlinkIdentity(input UnlinkIdentityInput) error {
	s.logger.Info("Unlinking OAuth identity: %s", input.IdentityID)

	client := s.supabaseClient.GetClient()
	
	// Use Supabase Auth UnlinkIdentity
	err := client.Auth.UnlinkIdentity(client.Context, input.IdentityID)
	if err != nil {
		s.logger.Error("Failed to unlink identity: %v", err)
		
		if strings.Contains(err.Error(), "not found") ||
		   strings.Contains(err.Error(), "invalid") {
			return errors.NewError(errors.ErrorCodeNotFound, "OAuth identity not found")
		}
		
		return fmt.Errorf("failed to unlink identity: %w", err)
	}

	s.logger.Info("OAuth identity unlinked successfully: %s", input.IdentityID)
	return nil
}

