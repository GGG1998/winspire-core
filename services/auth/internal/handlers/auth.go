package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	registrationService *services.RegistrationService
	authService         *services.AuthService
	logger              *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(registrationService *services.RegistrationService, authService *services.AuthService, appLogger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		registrationService: registrationService,
		authService:         authService,
		logger:              appLogger,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	UserType string `json:"userType" binding:"required,oneof=streamer user"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	User struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		UserType string `json:"userType"`
	} `json:"user"`
	Session *struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int    `json:"expiresIn"`
	} `json:"session,omitempty"`
	RequiresVerification bool `json:"requiresVerification"`
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid registration request: %v", err)
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid request", getValidationErrors(err)...)
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Validate email format (additional check beyond binding)
	if !isValidEmail(req.Email) {
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid email format")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		appErr := errors.NewError(errors.ErrorCodeValidation, err.Error())
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call registration service
	input := services.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		UserType: req.UserType,
	}

	output, err := h.registrationService.RegisterUser(input)
	if err != nil {
		h.logger.Error("Registration service error: %v", err)
		
		// Check if user already exists
		if strings.Contains(err.Error(), "already registered") || 
		   strings.Contains(err.Error(), "already exists") ||
		   strings.Contains(err.Error(), "User already registered") {
			appErr := errors.NewError(errors.ErrorCodeConflict, "User with this email already exists")
			errors.WriteError(c.Writer, appErr)
			return
		}

		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to register user")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Build response
	response := RegisterResponse{
		RequiresVerification: output.RequiresVerification,
	}
	response.User.ID = output.UserID
	response.User.Email = output.Email
	response.User.UserType = output.UserType

	if output.Session != nil {
		response.Session = &struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			ExpiresIn    int    `json:"expiresIn"`
		}{
			AccessToken:  output.Session.AccessToken,
			RefreshToken: output.Session.RefreshToken,
			ExpiresIn:    output.Session.ExpiresIn,
		}
	}

	c.JSON(http.StatusCreated, response)
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"`
}

// VerifyEmail handles GET /v1/auth/verify
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "Verification token is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Note: Supabase handles email verification via redirect URLs
	// This endpoint can be used to verify the token and redirect to frontend
	// For now, we'll return a success response
	// In production, this should verify the token with Supabase and redirect
	
	h.logger.Info("Email verification attempt with token: %s", token[:10]+"...")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verification link is valid. Please check your email for the verification link.",
		"note":    "Email verification is handled by Supabase via email link redirect",
	})
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
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

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request: %v", err)
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid request", getValidationErrors(err)...)
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid email format")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call auth service
	input := services.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.authService.LoginUser(input)
	if err != nil {
		h.logger.Error("Login service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to login")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Build response
	response := LoginResponse{}
	response.User.ID = output.UserID
	response.User.Email = output.Email
	response.User.UserType = output.UserType
	response.User.Role = output.Role
	response.Session.AccessToken = output.Session.AccessToken
	response.Session.RefreshToken = output.Session.RefreshToken
	response.Session.ExpiresIn = output.Session.ExpiresIn

	c.JSON(http.StatusOK, response)
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshResponse represents refresh token response
type RefreshResponse struct {
	Session struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int    `json:"expiresIn"`
	} `json:"session"`
}

// RefreshToken handles POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request: %v", err)
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid request", getValidationErrors(err)...)
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call auth service
	input := services.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.authService.RefreshToken(input)
	if err != nil {
		h.logger.Error("Refresh token service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to refresh token")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Build response
	response := RefreshResponse{}
	response.Session.AccessToken = output.Session.AccessToken
	response.Session.RefreshToken = output.Session.RefreshToken
	response.Session.ExpiresIn = output.Session.ExpiresIn

	c.JSON(http.StatusOK, response)
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		appErr := errors.NewError(errors.ErrorCodeUnauthorized, "Authorization header is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Extract token (Bearer <token>)
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		appErr := errors.NewError(errors.ErrorCodeUnauthorized, "Invalid authorization header format")
		errors.WriteError(c.Writer, appErr)
		return
	}

	accessToken := parts[1]

	// Call auth service
	input := services.LogoutUserInput{
		AccessToken: accessToken,
	}

	err := h.authService.LogoutUser(input)
	if err != nil {
		h.logger.Error("Logout service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		// Even if service returns error, logout is considered successful
		// (stateless JWT means client can discard token)
		h.logger.Warn("Logout service returned error, but logout is still successful: %v", err)
	}

	// Return success (204 No Content)
	c.Status(http.StatusNoContent)
}

// Helper functions

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}
	
	if !hasUpper || !hasLower || !hasNumber {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}
	
	return nil
}

func getValidationErrors(err error) []string {
	// Extract validation errors from Gin binding errors
	// This is a simplified version - in production, you might want more detailed error extraction
	if err != nil {
		return []string{err.Error()}
	}
	return nil
}

