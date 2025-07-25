// Package logging provides centralized logging functionality for LazyOC.
// It supports different log levels and can be configured for debug or production mode.
// In debug mode, logs are written to a file; in production, they are discarded to avoid interfering with the TUI.
package logging

import (
	"io"
	"log"
	"os"

	"github.com/katyella/lazyoc/internal/constants"
)

// Logger level constants define the available log levels for the application.
const (
	// LevelDebug represents debug-level logging for detailed diagnostic information
	LevelDebug = "DEBUG"
	
	// LevelInfo represents informational logging for general application events
	LevelInfo = "INFO"
	
	// LevelWarn represents warning-level logging for potentially problematic situations
	LevelWarn = "WARN"
	
	// LevelError represents error-level logging for error conditions
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
