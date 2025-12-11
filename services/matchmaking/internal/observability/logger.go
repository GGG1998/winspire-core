package observability

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// LogLevel represents logging severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Logger provides structured logging to CloudWatch Logs
type Logger struct {
	level  LogLevel
	format string // "json" or "text"
}

// NewLogger creates a new logger
func NewLogger(level, format string) *Logger {
	return &Logger{
		level:  LogLevel(level),
		format: format,
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     time.Time         `json:"timestamp"`
	Level         string            `json:"level"`
	Message       string            `json:"message"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	TournamentID  string            `json:"tournament_id,omitempty"`
	MatchID       string            `json:"match_id,omitempty"`
	PlayerID      string            `json:"player_id,omitempty"`
	EventType     string            `json:"event_type,omitempty"`
	ErrorType     string            `json:"error_type,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	if l.shouldLog(LogLevelDebug) {
		l.log(LogLevelDebug, message, fields)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	if l.shouldLog(LogLevelInfo) {
		l.log(LogLevelInfo, message, fields)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	if l.shouldLog(LogLevelWarn) {
		l.log(LogLevelWarn, message, fields)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	if l.shouldLog(LogLevelError) {
		l.log(LogLevelError, message, fields)
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     string(level),
		Message:   message,
		Fields:    fields,
	}

	// Extract common fields
	if correlationID, ok := fields["correlation_id"].(string); ok {
		entry.CorrelationID = correlationID
	}
	if tournamentID, ok := fields["tournament_id"].(string); ok {
		entry.TournamentID = tournamentID
	}
	if matchID, ok := fields["match_id"].(string); ok {
		entry.MatchID = matchID
	}
	if playerID, ok := fields["player_id"].(string); ok {
		entry.PlayerID = playerID
	}
	if eventType, ok := fields["event_type"].(string); ok {
		entry.EventType = eventType
	}
	if errorType, ok := fields["error_type"].(string); ok {
		entry.ErrorType = errorType
	}

	if l.format == "json" {
		jsonBytes, _ := json.Marshal(entry)
		log.Println(string(jsonBytes))
	} else {
		log.Printf("[%s] %s %v", level, message, fields)
	}
}

// shouldLog checks if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	return levels[level] >= levels[l.level]
}

// SetOutput sets the logger output (for testing)
func (l *Logger) SetOutput(w *os.File) {
	log.SetOutput(w)
}










