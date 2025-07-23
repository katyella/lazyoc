package logging

import (
	"io"
	"log"
	"os"

	"github.com/katyella/lazyoc/internal/constants"
)

// Logger levels
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// SetupLogger creates a logger for the application
// In debug mode, logs go to a file, otherwise they're discarded
func SetupLogger(debug bool) *log.Logger {
	if debug {
		// Create or append to debug log file
		file, err := os.OpenFile(constants.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, constants.LogFilePermissions)
		if err != nil {
			// Fallback to stderr if file creation fails
			return log.New(os.Stderr, "[LazyOC] ", log.LstdFlags|log.Lshortfile)
		}
		return log.New(file, "[LazyOC] ", log.LstdFlags|log.Lshortfile)
	}

	// In production, discard logs to avoid interfering with TUI
	return log.New(io.Discard, "", 0)
}

// Debug logs a debug message
func Debug(logger *log.Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("["+LevelDebug+"] "+msg, args...)
	}
}

// Info logs an info message
func Info(logger *log.Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("["+LevelInfo+"] "+msg, args...)
	}
}

// Warn logs a warning message
func Warn(logger *log.Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("["+LevelWarn+"] "+msg, args...)
	}
}

// Error logs an error message
func Error(logger *log.Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("["+LevelError+"] "+msg, args...)
	}
}