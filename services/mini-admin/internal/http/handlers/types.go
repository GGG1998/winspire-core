package handlers

import (
	"time"

	"github.com/google/uuid"
)

// === Response Types ===

// GameResponse represents a game in API responses.
type GameResponse struct {
	ID                string    `json:"id"`
	Slug              string    `json:"slug"`
	Name              string    `json:"name"`
	Description       *string   `json:"description,omitempty"`
	LogoURL           *string   `json:"logoUrl,omitempty"`
	S3Path            string    `json:"s3Path"`
	Version           string    `json:"version"`
	VersioningEnabled bool      `json:"versioningEnabled"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// GamesListResponse represents paginated list response.
type GamesListResponse struct {
	Games []GameResponse `json:"games"`
	Total int            `json:"total"`
}

// === Request Types ===

// CreateGameRequest represents the request body for creating a game.
type CreateGameRequest struct {
	Slug              string  `json:"slug" binding:"required,min=2,max=100"`
	Name              string  `json:"name" binding:"required,min=2,max=200"`
	Description       *string `json:"description"`
	LogoURL           *string `json:"logoUrl"`
	Version           string  `json:"version" binding:"required"`
	VersioningEnabled bool    `json:"versioningEnabled"`
}

// UpdateGameRequest represents the request body for updating a game.
type UpdateGameRequest struct {
	Slug              *string `json:"slug" binding:"omitempty,min=2,max=100"`
	Name              *string `json:"name" binding:"omitempty,min=2,max=200"`
	Description       *string `json:"description"`
	LogoURL           *string `json:"logoUrl"`
	Version           *string `json:"version"`
	VersioningEnabled *bool   `json:"versioningEnabled"`
	IsActive          *bool   `json:"isActive"`
}

// UploadFilesRequest represents metadata for file uploads
type UploadFilesRequest struct {
	GameID uuid.UUID `json:"gameId" binding:"required"`
}

// === Error Response ===

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}
