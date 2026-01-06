package handlers

import (
	"github.com/google/uuid"
)

// GameResponse represents a game in API responses.
type GameResponse struct {
	ID                string  `json:"id"`
	GameIntegrationID *string `json:"gameIntegrationId,omitempty"`
	Slug              string  `json:"slug"`
	Name              string  `json:"name"`
	StoragePath       string  `json:"storagePath"`
	Description       *string `json:"description,omitempty"`
	LogoURL           *string `json:"logoUrl,omitempty"`
	Version           string  `json:"version"`
	IsActive          bool    `json:"isActive"`
	CreatedAt         string  `json:"createdAt"` // ISO 8601 format
	UpdatedAt         string  `json:"updatedAt"` // ISO 8601 format
}

// GamesListResponse represents the response for listing games.
type GamesListResponse struct {
	Games []GameResponse `json:"games"`
	Total int            `json:"total"`
}

// CreateGameRequest represents the request body for creating a game.
type CreateGameRequest struct {
	GameIntegrationID *uuid.UUID `json:"gameIntegrationId"`
	Slug              string     `json:"slug" binding:"required,min=2,max=50"`
	Name              string     `json:"name" binding:"required,min=2,max=100"`
	StoragePath       string     `json:"storagePath" binding:"required"`
	Description       *string    `json:"description"`
	LogoURL           *string    `json:"logoUrl"`
	Version           string     `json:"version" binding:"required"`
}

// UpdateGameRequest represents the request body for updating a game.
type UpdateGameRequest struct {
	GameIntegrationID *uuid.UUID `json:"gameIntegrationId"`
	Slug              *string    `json:"slug" binding:"omitempty,min=2,max=50"`
	Name              *string    `json:"name" binding:"omitempty,min=2,max=100"`
	Description       *string    `json:"description"`
	LogoURL           *string    `json:"logoUrl"`
	StoragePath       *string    `json:"storagePath"`
	Version           *string    `json:"version"`
	IsActive          *bool      `json:"isActive"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
