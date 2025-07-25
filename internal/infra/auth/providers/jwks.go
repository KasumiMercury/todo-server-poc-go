package providers

import (
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/jwks"
)

const (
	JWKsStrategyName = "JWKs"
	JWKsPriority     = 200 // Medium priority
)

type JWKsStrategy struct {
	name       string
	configured bool
	jwksClient *jwks.Client
}

func NewJWKsStrategy(cfg config.Config) (*JWKsStrategy, error) {
	strategy := &JWKsStrategy{
		name:       JWKsStrategyName,
		configured: cfg.Auth.JWKs.EndpointURL != "",
	}

	if strategy.configured {
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

		strategy.jwksClient = jwksClient
	}

	return strategy, nil
}

func (s *JWKsStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured || s.jwksClient == nil {
		return auth.NewTokenValidationResult(false, "", nil, auth.ErrProviderNotConfigured)
	}

	return s.jwksClient.ValidateToken(tokenString)
}

func (s *JWKsStrategy) Name() string {
	return s.name
}

func (s *JWKsStrategy) IsConfigured() bool {
	return s.configured
}

func (s *JWKsStrategy) Priority() int {
	return JWKsPriority
}

func (s *JWKsStrategy) GetClient() *jwks.Client {
	return s.jwksClient
}
