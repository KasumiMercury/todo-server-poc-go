package auth

import (
	"sort"
	"strings"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth/providers"
)

// AuthenticationService manages multiple authentication strategies and handles JWT token validation.
// It supports multiple authentication providers (private key, JWKs, secret) with priority-based selection.
type AuthenticationService struct {
	strategies []auth.AuthenticationStrategy
}

// NewAuthenticationService creates a new AuthenticationService with configured authentication strategies.
// It initializes all available providers and sorts them by priority (highest first).
func NewAuthenticationService(cfg config.Config) (*AuthenticationService, error) {
	service := &AuthenticationService{
		strategies: make([]auth.AuthenticationStrategy, 0),
	}

	// Create and add configured providers
	if err := service.initializeProviders(cfg); err != nil {
		return nil, err
	}

	// Sort providers by priority (higher priority first)
	sort.Slice(service.strategies, func(i, j int) bool {
		return service.strategies[i].Priority() > service.strategies[j].Priority()
	})

	return service, nil
}

// NewAuthenticationServiceWithProviders creates a new AuthenticationService with provided strategies.
func NewAuthenticationServiceWithProviders(strategies []auth.AuthenticationStrategy) (*AuthenticationService, error) {
	if len(strategies) == 0 {
		return nil, auth.ErrNoValidProvider
	}

	service := &AuthenticationService{
		strategies: strategies,
	}

	// Sort providers by priority (higher priority first)
	sort.Slice(service.strategies, func(i, j int) bool {
		return service.strategies[i].Priority() > service.strategies[j].Priority()
	})

	return service, nil
}

func (s *AuthenticationService) initializeProviders(cfg config.Config) error {
	// Initialize providers based on configuration
	providerFactories := []func(config.Config) (auth.AuthenticationStrategy, error){
		func(cfg config.Config) (auth.AuthenticationStrategy, error) {
			return providers.NewPrivateKeyStrategy(cfg)
		},
		func(cfg config.Config) (auth.AuthenticationStrategy, error) {
			return providers.NewJWKsStrategy(cfg)
		},
		func(cfg config.Config) (auth.AuthenticationStrategy, error) {
			return providers.NewSecretStrategy(cfg)
		},
	}

	for _, providerFactory := range providerFactories {
		provider, err := providerFactory(cfg)
		if err != nil {
			return err
		}

		if provider.IsConfigured() {
			s.strategies = append(s.strategies, provider)
		}
	}

	if len(s.strategies) == 0 {
		return auth.ErrNoValidProvider
	}

	return nil
}

// ValidateToken validates a JWT token using the configured authentication strategies.
// It tries each strategy in priority order and returns the first successful validation result.
func (s *AuthenticationService) ValidateToken(tokenString string) *auth.AuthenticationResult {
	if tokenString == "" {
		return auth.NewAuthenticationResult(nil,
			auth.NewTokenValidationResult(false, "", auth.ErrInvalidTokenFormat))
	}

	var lastError error

	for _, strategy := range s.strategies {
		result := strategy.ValidateToken(tokenString)
		if result.IsValid() {
			return auth.NewAuthenticationResult(strategy, result)
		}

		if result.Error() != nil {
			lastError = result.Error()
		}
	}

	// All providers failed
	if lastError == nil {
		lastError = auth.ErrAllProvidersFailed
	}

	return auth.NewAuthenticationResult(nil,
		auth.NewTokenValidationResult(false, "", lastError))
}

// ExtractTokenFromHeader extracts a JWT token from the Authorization header.
// It expects the header to be in the format "Bearer <token>".
func (s *AuthenticationService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", auth.ErrMissingAuthorizationHeader
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", auth.ErrInvalidAuthorizationFormat
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", auth.ErrInvalidAuthorizationFormat
	}

	return token, nil
}

// GetConfiguredProviders returns the names of all configured authentication providers.
func (s *AuthenticationService) GetConfiguredProviders() []string {
	names := make([]string, len(s.strategies))
	for i, provider := range s.strategies {
		names[i] = provider.Name()
	}

	return names
}

// GetProviderCount returns the number of configured authentication providers.
func (s *AuthenticationService) GetProviderCount() int {
	return len(s.strategies)
}
