package providers

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

const (
	// SecretStrategyName is the identifier for the secret-based authentication strategy.
	SecretStrategyName = "Secret"
	// SecretPriority defines the priority of the secret strategy (lowest priority, used as fallback).
	SecretPriority = 100 // Lowest priority (fallback)
)

// SecretStrategy implements JWT token validation using a shared secret key.
// It validates tokens using HMAC signing method and serves as a fallback authentication strategy.
type SecretStrategy struct {
	name       string
	configured bool
	secretKey  string
}

// NewSecretStrategy creates a new SecretStrategy instance from the provided configuration.
// It is configured if a JWT secret is provided in the configuration.
func NewSecretStrategy(cfg config.Config) (*SecretStrategy, error) {
	strategy := &SecretStrategy{
		name:       SecretStrategyName,
		configured: cfg.Auth.JWTSecret != "",
		secretKey:  cfg.Auth.JWTSecret,
	}

	return strategy, nil
}

// ValidateToken validates a JWT token using the shared secret key.
// It verifies the token signature using HMAC signing method and extracts the user ID from the 'sub' claim.
func (s *SecretStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured {
		return auth.NewTokenValidationResult(false, "", auth.ErrProviderNotConfigured)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrInvalidTokenSignature
		}

		return []byte(s.secretKey), nil
	})
	if err != nil {
		return auth.NewTokenValidationResult(false, "", err)
	}

	if !token.Valid {
		return auth.NewTokenValidationResult(false, "", auth.ErrTokenValidation)
	}

	claims := make(map[string]interface{})

	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
		for k, v := range mapClaims {
			claims[k] = v
		}
	}

	userID := ""
	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	}

	return auth.NewTokenValidationResult(true, userID, nil)
}

func (s *SecretStrategy) Name() string {
	return s.name
}

// IsConfigured returns whether this strategy is properly configured with a secret key.
func (s *SecretStrategy) IsConfigured() bool {
	return s.configured
}

// Priority returns the priority of this authentication strategy.
func (s *SecretStrategy) Priority() int {
	return SecretPriority
}

// GetSecretKey returns the secret key used for token validation (for testing purposes).
func (s *SecretStrategy) GetSecretKey() string {
	return s.secretKey
}
