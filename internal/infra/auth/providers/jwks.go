package providers

import (
	"context"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/jwks"
)

const (
	// JWKsStrategyName is the identifier for the JWKs authentication strategy.
	JWKsStrategyName = "JWKs"
	// JWKsPriority defines the priority of the JWKs strategy (medium priority).
	JWKsPriority = 200 // Medium priority
)

// JWKSValidator defines the interface for JWT token validation using JWKs.
// This interface allows for easier testing by enabling dependency injection of mock implementations.
type JWKSValidator interface {
	// ValidateToken validates a JWT token and returns the validation result.
	ValidateToken(tokenString string) *auth.TokenValidationResult
	// Refresh manually refreshes the JWKs cache from the endpoint.
	Refresh(ctx context.Context) error
}

// JWKsStrategy implements JWT token validation using JSON Web Key Sets (JWKs).
// It fetches public keys from a JWKs endpoint to validate JWT tokens.
type JWKsStrategy struct {
	name       string
	configured bool
	validator  JWKSValidator
}

// NewJWKsStrategy creates a new JWKsStrategy instance from the provided configuration.
// It returns a configured strategy if a JWKs endpoint URL is provided, otherwise returns an unconfigured strategy.
func NewJWKsStrategy(cfg config.Config) (*JWKsStrategy, error) {
	if cfg.Auth.JWKs.EndpointURL == "" {
		return &JWKsStrategy{
			name:       JWKsStrategyName,
			configured: false,
			validator:  nil,
		}, nil
	}

	endpoint, err := auth.NewJWKsEndpoint(cfg.Auth.JWKs.EndpointURL)
	if err != nil {
		return nil, err
	}

	cacheConfig := auth.NewJWKsCacheConfig(
		time.Duration(cfg.Auth.JWKs.CacheDuration)*time.Second,
		time.Duration(cfg.Auth.JWKs.RefreshPadding)*time.Second,
	)

	validator, err := jwks.NewClient(endpoint, cacheConfig)
	if err != nil {
		return nil, err
	}

	return &JWKsStrategy{
		name:       JWKsStrategyName,
		configured: true,
		validator:  validator,
	}, nil
}

// NewJWKsStrategyWithValidator creates a new JWKsStrategy instance with a custom validator.
func NewJWKsStrategyWithValidator(validator JWKSValidator) *JWKsStrategy {
	configured := validator != nil
	return &JWKsStrategy{
		name:       JWKsStrategyName,
		configured: configured,
		validator:  validator,
	}
}

// ValidateToken validates a JWT token using the JWKs endpoint.
// It returns an error result if the strategy is not configured.
func (s *JWKsStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured || s.validator == nil {
		return auth.NewTokenValidationResult(false, "", auth.ErrProviderNotConfigured)
	}

	return s.validator.ValidateToken(tokenString)
}

// Name returns the name of this authentication strategy.
func (s *JWKsStrategy) Name() string {
	return s.name
}

// IsConfigured returns whether this strategy is properly configured with a JWKs endpoint.
func (s *JWKsStrategy) IsConfigured() bool {
	return s.configured
}

// Priority returns the priority of this authentication strategy.
func (s *JWKsStrategy) Priority() int {
	return JWKsPriority
}

// GetValidator returns the underlying JWKs validator for testing or advanced usage.
func (s *JWKsStrategy) GetValidator() JWKSValidator {
	return s.validator
}
