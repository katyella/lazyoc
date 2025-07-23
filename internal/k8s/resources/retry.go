package resources

import (
	"context"
	"fmt"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/katyella/lazyoc/internal/constants"
)

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
	Delay     time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are generally retryable
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Kubernetes API errors
	if statusErr, ok := err.(*errors.StatusError); ok {
		switch statusErr.ErrStatus.Code {
		case 429: // Too Many Requests
			return true
		case constants.HTTPStatusInternalServerError, 
			constants.HTTPStatusBadGateway, 
			constants.HTTPStatusServiceUnavailable, 
			constants.HTTPStatusGatewayTimeout: // Server errors
			return true
		case 408: // Request Timeout
			return true
		default:
			return false
		}
	}

	// Context errors (except cancellation)
	if err == context.DeadlineExceeded {
		return true
	}

	return false
}

// GetRetryDelay returns appropriate delay for retry based on error type
func GetRetryDelay(err error, attempt int) time.Duration {
	baseDelay := time.Duration(attempt) * time.Second

	// Check for rate limiting
	if statusErr, ok := err.(*errors.StatusError); ok {
		if statusErr.ErrStatus.Code == 429 {
			// For rate limiting, use exponential backoff with jitter
			return time.Duration(1<<uint(attempt)) * time.Second
		}
	}

	// For other errors, use linear backoff with max
	if baseDelay > 30*time.Second {
		return 30 * time.Second
	}

	return baseDelay
}

// RetryOperation performs an operation with retry logic
func RetryOperation[T any](ctx context.Context, maxRetries int, operation func(ctx context.Context) (T, error)) (T, error) {
	var lastErr error
	var result T

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := operation(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			break
		}

		// Calculate delay
		delay := GetRetryDelay(err, attempt+1)

		// Check if context allows for the delay
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

// WithRetry wraps a ResourceClient with retry logic
type RetryWrapper struct {
	client     ResourceClient
	maxRetries int
}

// NewRetryWrapper creates a new retry wrapper
func NewRetryWrapper(client ResourceClient, maxRetries int) *RetryWrapper {
	return &RetryWrapper{
		client:     client,
		maxRetries: maxRetries,
	}
}

// ListPods with retry logic
func (r *RetryWrapper) ListPods(ctx context.Context, opts ListOptions) (*ResourceList[PodInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[PodInfo], error) {
		return r.client.ListPods(ctx, opts)
	})
}

// GetPod with retry logic
func (r *RetryWrapper) GetPod(ctx context.Context, namespace, name string) (*PodInfo, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*PodInfo, error) {
		return r.client.GetPod(ctx, namespace, name)
	})
}

// ListServices with retry logic
func (r *RetryWrapper) ListServices(ctx context.Context, opts ListOptions) (*ResourceList[ServiceInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[ServiceInfo], error) {
		return r.client.ListServices(ctx, opts)
	})
}

// GetService with retry logic
func (r *RetryWrapper) GetService(ctx context.Context, namespace, name string) (*ServiceInfo, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ServiceInfo, error) {
		return r.client.GetService(ctx, namespace, name)
	})
}

// ListDeployments with retry logic
func (r *RetryWrapper) ListDeployments(ctx context.Context, opts ListOptions) (*ResourceList[DeploymentInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[DeploymentInfo], error) {
		return r.client.ListDeployments(ctx, opts)
	})
}

// GetDeployment with retry logic
func (r *RetryWrapper) GetDeployment(ctx context.Context, namespace, name string) (*DeploymentInfo, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*DeploymentInfo, error) {
		return r.client.GetDeployment(ctx, namespace, name)
	})
}

// ListConfigMaps with retry logic
func (r *RetryWrapper) ListConfigMaps(ctx context.Context, opts ListOptions) (*ResourceList[ConfigMapInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[ConfigMapInfo], error) {
		return r.client.ListConfigMaps(ctx, opts)
	})
}

// GetConfigMap with retry logic
func (r *RetryWrapper) GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMapInfo, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ConfigMapInfo, error) {
		return r.client.GetConfigMap(ctx, namespace, name)
	})
}

// ListSecrets with retry logic
func (r *RetryWrapper) ListSecrets(ctx context.Context, opts ListOptions) (*ResourceList[SecretInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[SecretInfo], error) {
		return r.client.ListSecrets(ctx, opts)
	})
}

// GetSecret with retry logic
func (r *RetryWrapper) GetSecret(ctx context.Context, namespace, name string) (*SecretInfo, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*SecretInfo, error) {
		return r.client.GetSecret(ctx, namespace, name)
	})
}

// ListNamespaces with retry logic
func (r *RetryWrapper) ListNamespaces(ctx context.Context) (*ResourceList[NamespaceInfo], error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*ResourceList[NamespaceInfo], error) {
		return r.client.ListNamespaces(ctx)
	})
}

// GetCurrentNamespace (no retry needed)
func (r *RetryWrapper) GetCurrentNamespace() string {
	return r.client.GetCurrentNamespace()
}

// SetCurrentNamespace (no retry needed)
func (r *RetryWrapper) SetCurrentNamespace(namespace string) error {
	return r.client.SetCurrentNamespace(namespace)
}

// GetNamespaceContext with retry logic
func (r *RetryWrapper) GetNamespaceContext() (*NamespaceContext, error) {
	ctx := context.Background()
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (*NamespaceContext, error) {
		return r.client.GetNamespaceContext()
	})
}

// TestConnection with retry logic
func (r *RetryWrapper) TestConnection(ctx context.Context) error {
	_, err := RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, r.client.TestConnection(ctx)
	})
	return err
}

// GetServerInfo with retry logic
func (r *RetryWrapper) GetServerInfo(ctx context.Context) (map[string]interface{}, error) {
	return RetryOperation(ctx, r.maxRetries, func(ctx context.Context) (map[string]interface{}, error) {
		return r.client.GetServerInfo(ctx)
	})
}