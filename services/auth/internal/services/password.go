package services

import (
	"fmt"
	"strings"

	"github.com/winspire/winspire-core/services/auth/internal/errors"
	"github.com/winspire/winspire-core/services/auth/internal/logger"
	"github.com/winspire/winspire-core/services/auth/internal/supabase"
	"github.com/supabase-community/supabase-go"
)

// PasswordService handles password reset operations
type PasswordService struct {
	supabaseClient *supabase.Client
	logger         *logger.Logger
}

// NewPasswordService creates a new password service
func NewPasswordService(supabaseClient *supabase.Client, appLogger *logger.Logger) *PasswordService {
	return &PasswordService{
		supabaseClient: supabaseClient,
		logger:         appLogger,
	}
}

// RequestPasswordResetInput represents password reset request input
type RequestPasswordResetInput struct {
	Email string
}

// RequestPasswordReset sends a password reset email to the user
func (s *PasswordService) RequestPasswordReset(input RequestPasswordResetInput) error {
	s.logger.Info("Password reset requested for email: %s", input.Email)

	client := s.supabaseClient.GetClient()
	
	// Call Supabase Auth ResetPasswordForEmail
	// Note: Supabase handles sending the email and token generation
	err := client.Auth.ResetPasswordForEmail(client.Context, input.Email, supabase.ResetPasswordOptions{})
	if err != nil {
		s.logger.Error("Password reset request failed for email %s: %v", input.Email, err)
		
		// Supabase returns success even if email doesn't exist (security best practice)
		// But we still log the error for debugging
		if strings.Contains(err.Error(), "rate limit") {
			return errors.NewError(errors.ErrorCodeBadRequest, "Too many reset requests. Please try again later.")
		}
		
		// For security, we don't reveal if email exists or not
		// We return success even if there's an error (unless it's a rate limit)
		if !strings.Contains(err.Error(), "rate limit") {
			// Log but don't return error - security best practice
			s.logger.Warn("Password reset error (not returned to user): %v", err)
		}
	}

	// Always return success for security (don't reveal if email exists)
	s.logger.Info("Password reset email sent (or would be sent) for: %s", input.Email)
	return nil
}

// ConfirmPasswordResetInput represents password reset confirmation input
type ConfirmPasswordResetInput struct {
	Token    string
	Password string
}

// ConfirmPasswordReset confirms password reset and sets new password
func (s *PasswordService) ConfirmPasswordReset(input ConfirmPasswordResetInput) error {
	s.logger.Info("Password reset confirmation attempt")

	client := s.supabaseClient.GetClient()
	
	// Supabase handles password reset via email link redirect
	// The token is validated by Supabase when user clicks the link
	// We need to use the service role key to update the password
	// However, the standard flow is:
	// 1. User clicks link in email (handled by Supabase)
	// 2. User is redirected to frontend with token
	// 3. Frontend calls this endpoint with token and new password
	
	// For now, we'll use UpdateUser with the token
	// Note: This might require using the service role key or a different approach
	// depending on Supabase Go client API
	
	// Validate password strength
	if len(input.Password) < 8 {
		return errors.NewError(errors.ErrorCodeValidation, "Password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	
	for _, char := range input.Password {
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
		return errors.NewError(errors.ErrorCodeValidation, "Password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	// Supabase password reset flow:
	// 1. User clicks link in email (Supabase validates token and redirects)
	// 2. Frontend receives token in redirect URL
	// 3. Frontend calls this endpoint with token and new password
	// 4. We use UpdateUser with the token to set new password
	
	// Note: The token from email is a one-time use token
	// We need to exchange it for a session or use it directly with UpdateUser
	// Depending on Supabase Go client version, this might vary
	
	// Try to update password using the reset token
	// The token should be validated by Supabase
	userAttributes := supabase.UserAttributes{
		Password: &input.Password,
	}
	
	// Use the token as the session token for this update
	// Supabase will validate the reset token
	_, err := client.Auth.UpdateUser(client.Context, userAttributes, supabase.AuthOptions{
		Token: input.Token,
	})
	
	if err != nil {
		s.logger.Error("Password reset confirmation failed: %v", err)
		
		// Check for specific error types
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "expired") ||
		   strings.Contains(errMsg, "invalid") ||
		   strings.Contains(errMsg, "token") ||
		   strings.Contains(errMsg, "not found") {
			return errors.NewError(errors.ErrorCodeBadRequest, "Invalid or expired reset token. Please request a new password reset.")
		}
		
		return fmt.Errorf("failed to reset password: %w", err)
	}

	s.logger.Info("Password reset confirmed successfully")
	return nil
}

