package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/services"
)

// OAuthHandler handles OAuth authentication requests
type OAuthHandler struct {
	oauthService *services.OAuthService
	logger       *logger.Logger
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *services.OAuthService, appLogger *logger.Logger) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		logger:       appLogger,
	}
}

// InitiateOAuthRequest represents OAuth initiation request
type InitiateOAuthRequest struct {
	UserType    string `form:"userType" binding:"required,oneof=streamer user"`
	RedirectURI string `form:"redirect_uri" binding:"required,url"`
}

// InitiateOAuth handles GET /v1/auth/oauth/{provider}
func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "OAuth provider is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Get user type and redirect URI from query params
	userType := c.Query("userType")
	redirectURI := c.Query("redirect_uri")

	if userType == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "userType is required (streamer or user)")
		errors.WriteError(c.Writer, appErr)
		return
	}

	if redirectURI == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "redirect_uri is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call OAuth service
	input := services.InitiateOAuthFlowInput{
		Provider:    provider,
		UserType:    userType,
		RedirectURI: redirectURI,
	}

	output, err := h.oauthService.InitiateOAuthFlow(input)
	if err != nil {
		h.logger.Error("OAuth initiation service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to initiate OAuth flow")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Redirect to OAuth provider
	c.Redirect(http.StatusFound, output.AuthURL)
}

// OAuthCallbackResponse represents OAuth callback response
type OAuthCallbackResponse struct {
	User struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		UserType string `json:"userType"`
		Role     string `json:"role"`
	} `json:"user"`
	Session struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int    `json:"expiresIn"`
	} `json:"session"`
}

// OAuthCallback handles GET /v1/auth/oauth/{provider}/callback
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "OAuth provider is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Get code and state from query params
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "OAuth authorization code is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call OAuth service
	input := services.HandleOAuthCallbackInput{
		Provider: provider,
		Code:     code,
		State:    state,
	}

	output, err := h.oauthService.HandleOAuthCallback(input)
	if err != nil {
		h.logger.Error("OAuth callback service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to process OAuth callback")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Build response
	response := OAuthCallbackResponse{}
	response.User.ID = output.UserID
	response.User.Email = output.Email
	response.User.UserType = output.UserType
	response.User.Role = output.Role
	response.Session.AccessToken = output.Session.AccessToken
	response.Session.RefreshToken = output.Session.RefreshToken
	response.Session.ExpiresIn = output.Session.ExpiresIn

	c.JSON(http.StatusOK, response)
}

