package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/services"
)

// PasswordHandler handles password reset requests
type PasswordHandler struct {
	passwordService *services.PasswordService
	logger          *logger.Logger
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(passwordService *services.PasswordService, appLogger *logger.Logger) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
		logger:          appLogger,
	}
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestPasswordReset handles POST /v1/auth/password/reset
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid password reset request: %v", err)
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

	// Call password service
	input := services.RequestPasswordResetInput{
		Email: req.Email,
	}

	err := h.passwordService.RequestPasswordReset(input)
	if err != nil {
		h.logger.Error("Password reset service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to process password reset request")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Always return success (security best practice - don't reveal if email exists)
	c.Status(http.StatusNoContent)
}

// PasswordResetConfirmRequest represents password reset confirmation request
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ConfirmPasswordReset handles POST /v1/auth/password/reset/confirm
func (h *PasswordHandler) ConfirmPasswordReset(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid password reset confirmation request: %v", err)
		appErr := errors.NewError(errors.ErrorCodeValidation, "Invalid request", getValidationErrors(err)...)
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		appErr := errors.NewError(errors.ErrorCodeValidation, err.Error())
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Validate token is not empty
	if req.Token == "" {
		appErr := errors.NewError(errors.ErrorCodeValidation, "Reset token is required")
		errors.WriteError(c.Writer, appErr)
		return
	}

	// Call password service
	input := services.ConfirmPasswordResetInput{
		Token:    req.Token,
		Password: req.Password,
	}

	err := h.passwordService.ConfirmPasswordReset(input)
	if err != nil {
		h.logger.Error("Password reset confirmation service error: %v", err)
		
		// Check if it's an AppError (from service)
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteError(c.Writer, appErr)
			return
		}
		
		appErr := errors.NewError(errors.ErrorCodeInternal, "Failed to reset password")
		errors.WriteError(c.Writer, appErr)
		return
	}

	c.Status(http.StatusNoContent)
}

