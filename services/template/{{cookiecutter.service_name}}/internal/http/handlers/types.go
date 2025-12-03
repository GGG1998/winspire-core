package handlers

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler holds all handler dependencies.
type Handler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// New creates a new Handler with the given dependencies.
func New(pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{
		pool:   pool,
		logger: logger,
	}
}

// Common response types

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}


