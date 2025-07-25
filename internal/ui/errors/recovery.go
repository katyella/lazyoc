package errors

import (
	"context"
	"math"
	"time"

	"github.com/katyella/lazyoc/internal/constants"
)

// RetryStrategy defines the retry behavior for different error types
type RetryStrategy struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Retryable     func(error) bool
}

// DefaultRetryStrategy returns a default retry strategy
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxAttempts:   constants.DefaultRetryAttempts,
		InitialDelay:  constants.DefaultInitialDelay,
		MaxDelay:      constants.DefaultMaxDelay,
		BackoffFactor: constants.DefaultBackoffFactor,
		Retryable:     IsRetryableError,
	}
}

// ConnectionRetryStrategy returns a retry strategy optimized for connection errors
func ConnectionRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxAttempts:   5,
		InitialDelay:  constants.ConnectionInitialDelay,
		MaxDelay:      constants.ConnectionMaxDelay,
		BackoffFactor: 1.5,
		Retryable:     IsConnectionError,
	}
}

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	Strategy     *RetryStrategy
	OnRetry      func(attempt int, err error)
	OnSuccess    func(attempt int)
	OnFinalError func(err error, attempts int)
}

// RetryOperation executes an operation with retry logic
func RetryOperation(ctx context.Context, config *RetryConfig, operation func() error) error {
	if config.Strategy == nil {
		config.Strategy = DefaultRetryStrategy()
	}

	var lastErr error
	delay := config.Strategy.InitialDelay

	for attempt := 1; attempt <= config.Strategy.MaxAttempts; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			// Success
			if config.OnSuccess != nil {
				config.OnSuccess(attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !config.Strategy.Retryable(err) {
			// Not retryable, fail immediately
			if config.OnFinalError != nil {
				config.OnFinalError(err, attempt)
			}
			return err
		}

		// Don't retry on last attempt
		if attempt == config.Strategy.MaxAttempts {
			break
		}

		// Notify about retry
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		// Wait before retry (with context cancellation support)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Strategy.BackoffFactor)
		if delay > config.Strategy.MaxDelay {
			delay = config.Strategy.MaxDelay
		}
	}

	// All attempts failed
	if config.OnFinalError != nil {
		config.OnFinalError(lastErr, config.Strategy.MaxAttempts)
	}

	return lastErr
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a UserFriendlyError
	if ufe, ok := err.(*UserFriendlyError); ok {
		return ufe.Retryable
	}

	// Map the error and check retryability
	mapped := MapKubernetesError(err)
	return mapped.Retryable
}

// IsConnectionError determines if an error is connection-related
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a UserFriendlyError
	if ufe, ok := err.(*UserFriendlyError); ok {
		return ufe.Category == ErrorCategoryConnection || ufe.Category == ErrorCategoryNetwork
	}

	// Map the error and check category
	mapped := MapKubernetesError(err)
	return mapped.Category == ErrorCategoryConnection || mapped.Category == ErrorCategoryNetwork
}

// BackoffDelay calculates exponential backoff delay
func BackoffDelay(attempt int, initialDelay time.Duration, maxDelay time.Duration, factor float64) time.Duration {
	delay := time.Duration(float64(initialDelay) * math.Pow(factor, float64(attempt-1)))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// RecoveryAction represents an action that can be taken to recover from an error
type RecoveryAction struct {
	Name        string
	Description string
	Action      func() error
	Automatic   bool // Whether this action can be performed automatically
}

// GetRecoveryActions returns suggested recovery actions for an error
func GetRecoveryActions(err error) []RecoveryAction {
	if err == nil {
		return nil
	}

	var actions []RecoveryAction

	// Map to user-friendly error if needed
	var ufe *UserFriendlyError
	if mapped, ok := err.(*UserFriendlyError); ok {
		ufe = mapped
	} else {
		ufe = MapKubernetesError(err)
	}

	// Add category-specific recovery actions
	switch ufe.Category {
	case ErrorCategoryConnection, ErrorCategoryNetwork:
		actions = append(actions, RecoveryAction{
			Name:        "Retry Connection",
			Description: "Attempt to reconnect to the cluster",
			Automatic:   true,
		})
		actions = append(actions, RecoveryAction{
			Name:        "Check Network",
			Description: "Verify network connectivity to the cluster",
			Automatic:   false,
		})

	case ErrorCategoryAuthentication:
		actions = append(actions, RecoveryAction{
			Name:        "Re-authenticate",
			Description: "Run 'oc login' to refresh authentication",
			Automatic:   false,
		})

	case ErrorCategoryProject:
		actions = append(actions, RecoveryAction{
			Name:        "Refresh Projects",
			Description: "Reload the project list",
			Automatic:   true,
		})
		actions = append(actions, RecoveryAction{
			Name:        "Switch Project",
			Description: "Try switching to a different project",
			Automatic:   false,
		})

	case ErrorCategoryResource:
		actions = append(actions, RecoveryAction{
			Name:        "Refresh Resources",
			Description: "Reload the resource list",
			Automatic:   true,
		})
	}

	// Always add a generic refresh action
	actions = append(actions, RecoveryAction{
		Name:        "Refresh Application",
		Description: "Refresh all data and reconnect",
		Automatic:   false,
	})

	return actions
}
