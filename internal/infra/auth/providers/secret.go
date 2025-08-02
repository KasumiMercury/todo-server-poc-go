package providers

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

const (
	SecretStrategyName = "Secret"
	SecretPriority     = 100 // Lowest priority (fallback)
)

type SecretStrategy struct {
	name       string
	configured bool
	secretKey  string
}

func NewSecretStrategy(cfg config.Config) (*SecretStrategy, error) {
	strategy := &SecretStrategy{
		name:       SecretStrategyName,
		configured: cfg.Auth.JWTSecret != "",
		secretKey:  cfg.Auth.JWTSecret,
	}

	return strategy, nil
}

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

func (s *SecretStrategy) IsConfigured() bool {
	return s.configured
}

func (s *SecretStrategy) Priority() int {
	return SecretPriority
}

func (s *SecretStrategy) GetSecretKey() string {
	return s.secretKey
}
