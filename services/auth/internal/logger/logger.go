package logger

import (
	"log"
	"os"
)

// Logger provides structured logging
type Logger struct {
	info  *log.Logger
	error *log.Logger
	warn  *log.Logger
	debug *log.Logger
}

// New creates a new logger instance
func New(env string) *Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile

	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", flags),
		error: log.New(os.Stderr, "[ERROR] ", flags),
		warn:  log.New(os.Stdout, "[WARN] ", flags),
		debug: log.New(os.Stdout, "[DEBUG] ", flags),
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.warn.Printf(format, v...)
}

// Debug logs a debug message (only in development)
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

