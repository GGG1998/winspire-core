package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	supabaseclient "github.com/winspire/winspire-core/services/auth/internal/supabase"
	"github.com/supabase-community/gotrue-go/types"
)

// OAuthService handles OAuth authentication flows
type OAuthService struct {
	supabaseClient *supabaseclient.Client
	supabaseURL    string
	logger         *logger.Logger
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(supabaseClient *supabaseclient.Client, supabaseURL string, appLogger *logger.Logger) *OAuthService {
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
	// For OAuth, Supabase redirects with access_token in URL fragment
	// We need to use Token endpoint with authorization_code grant type
	
	// Exchange code for token using Token endpoint with pkce grant type
	// Note: OAuth flow typically uses PKCE, but we can also use code directly
	tokenReq := types.TokenRequest{
		GrantType: "pkce",
		Code:      input.Code,
	}
	
	tokenResponse, err := client.Auth.Token(tokenReq)
	if err != nil {
		s.logger.Error("OAuth callback processing failed: %v", err)
		
		if strings.Contains(err.Error(), "invalid") ||
		   strings.Contains(err.Error(), "expired") ||
		   strings.Contains(err.Error(), "code") {
			return nil, errors.NewError(errors.ErrorCodeBadRequest, "Invalid or expired OAuth authorization code")
		}
		
		return nil, fmt.Errorf("failed to process OAuth callback: %w", err)
	}

	// Extract user and session from token response
	if tokenResponse.User.ID == (uuid.UUID{}) {
		return nil, errors.NewError(errors.ErrorCodeInternal, "User not found in OAuth response")
	}

	// Extract user information from token response
	authResponse := tokenResponse

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
	if authResponse.AccessToken == "" {
		return nil, errors.NewError(errors.ErrorCodeInternal, "No session returned from OAuth authentication")
	}

	session := &Session{
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		ExpiresIn:    authResponse.ExpiresIn,
	}

	s.logger.Info("OAuth authentication successful for provider: %s, user: %s", 
		input.Provider, authResponse.User.Email)

	return &HandleOAuthCallbackOutput{
		UserID:   tokenResponse.User.ID.String(),
		Email:    tokenResponse.User.Email,
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

	// Note: GetUserIdentities is not directly available in gotrue-go
	// We would need to use Admin API or query auth.identities directly
	// For now, return empty list - this can be implemented later if needed
	// TODO: Implement GetUserIdentities using Admin API or direct DB query
	// For now, return empty list
	identities := []types.Identity{}

	result := make([]OAuthIdentity, 0, len(identities))
	for _, identity := range identities {
		result = append(result, OAuthIdentity{
			ID:       identity.ID,
			Provider: identity.Provider,
			UserID:   identity.UserID.String(),
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

	// Note: UnlinkIdentity is not directly available in gotrue-go
	// We would need to use Admin API or direct DB query
	// For now, return error - this can be implemented later if needed
	// TODO: Implement UnlinkIdentity using Admin API or direct DB query
	err := fmt.Errorf("UnlinkIdentity not yet implemented - requires Admin API or direct DB access")
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

