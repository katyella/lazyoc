package auth

import (
	"context"

	"k8s.io/client-go/rest"
)

// AuthProvider defines the interface for authentication methods
type AuthProvider interface {
	// Authenticate attempts to authenticate and returns a rest.Config
	Authenticate(ctx context.Context) (*rest.Config, error)
	
	// IsValid checks if the current authentication is still valid
	IsValid(ctx context.Context) error
	
	// Refresh attempts to refresh the authentication credentials if possible
	Refresh(ctx context.Context) error
	
	// GetContext returns the current context name if applicable
	GetContext() string
	
	// GetNamespace returns the default namespace for this authentication
	GetNamespace() string
}

// AuthManager manages multiple authentication providers
type AuthManager struct {
	providers []AuthProvider
	active    AuthProvider
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		providers: make([]AuthProvider, 0),
	}
}

// AddProvider adds an authentication provider
func (am *AuthManager) AddProvider(provider AuthProvider) {
	am.providers = append(am.providers, provider)
}

// Authenticate tries each provider until one succeeds
func (am *AuthManager) Authenticate(ctx context.Context) (*rest.Config, error) {
	for _, provider := range am.providers {
		config, err := provider.Authenticate(ctx)
		if err == nil {
			am.active = provider
			return config, nil
		}
	}
	
	return nil, &AuthError{
		Type:    "authentication_failed",
		Message: "all authentication methods failed",
	}
}

// GetActiveProvider returns the currently active provider
func (am *AuthManager) GetActiveProvider() AuthProvider {
	return am.active
}

// IsValid checks if the current authentication is valid
func (am *AuthManager) IsValid(ctx context.Context) error {
	if am.active == nil {
		return &AuthError{
			Type:    "no_active_auth",
			Message: "no active authentication provider",
		}
	}
	
	return am.active.IsValid(ctx)
}

// Refresh attempts to refresh the current authentication
func (am *AuthManager) Refresh(ctx context.Context) error {
	if am.active == nil {
		return &AuthError{
			Type:    "no_active_auth",
			Message: "no active authentication provider",
		}
	}
	
	return am.active.Refresh(ctx)
}