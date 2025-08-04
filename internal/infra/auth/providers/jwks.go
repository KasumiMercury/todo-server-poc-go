package providers

import (
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

// JWKsStrategy implements JWT token validation using JSON Web Key Sets (JWKs).
// It fetches public keys from a JWKs endpoint to validate JWT tokens.
type JWKsStrategy struct {
	name       string
	configured bool
	jwksClient *jwks.Client
}

// NewJWKsStrategy creates a new JWKsStrategy instance from the provided configuration.
// It returns a configured strategy if a JWKs endpoint URL is provided, otherwise returns an unconfigured strategy.
func NewJWKsStrategy(cfg config.Config) (*JWKsStrategy, error) {
	if cfg.Auth.JWKs.EndpointURL != "" {
		endpoint, err := auth.NewJWKsEndpoint(cfg.Auth.JWKs.EndpointURL)
		if err != nil {
			return nil, err
		}

		cacheConfig := auth.NewJWKsCacheConfig(
			time.Duration(cfg.Auth.JWKs.CacheDuration)*time.Second,
			time.Duration(cfg.Auth.JWKs.RefreshPadding)*time.Second,
		)

		jwksClient, err := jwks.NewClient(endpoint, cacheConfig)
		if err != nil {
			return nil, err
		}

		return &JWKsStrategy{
			name:       JWKsStrategyName,
			configured: true,
			jwksClient: jwksClient,
		}, nil
	}

	return &JWKsStrategy{
		name:       JWKsStrategyName,
		configured: false,
		jwksClient: nil,
	}, nil
}

// ValidateToken validates a JWT token using the JWKs endpoint.
// It returns an error result if the strategy is not configured.
func (s *JWKsStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured || s.jwksClient == nil {
		return auth.NewTokenValidationResult(false, "", auth.ErrProviderNotConfigured)
	}

	return s.jwksClient.ValidateToken(tokenString)
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

// GetClient returns the underlying JWKs client for testing or advanced usage.
func (s *JWKsStrategy) GetClient() *jwks.Client {
	return s.jwksClient
}
