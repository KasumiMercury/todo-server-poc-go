package providers

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/keyloader"
)

const (
	PrivateKeyStrategyName = "PrivateKey"
	PrivateKeyPriority     = 300 // Highest priority
)

type PrivateKeyStrategy struct {
	name             string
	configured       bool
	privateKeyLoader auth.PrivateKeyLoader
	loadedPrivateKey *auth.LoadedPrivateKey
}

func NewPrivateKeyStrategy(cfg config.Config) (*PrivateKeyStrategy, error) {
	strategy := &PrivateKeyStrategy{
		name:       PrivateKeyStrategyName,
		configured: cfg.Auth.PrivateKeyFilePath != "",
	}

	if strategy.configured {
		strategy.privateKeyLoader = keyloader.NewFileLoader()

		privateKeyFile, err := auth.NewPrivateKeyFile(cfg.Auth.PrivateKeyFilePath)
		if err != nil {
			return nil, err
		}

		loadedKey, err := strategy.privateKeyLoader.LoadPrivateKey(privateKeyFile)
		if err != nil {
			return nil, err
		}

		strategy.loadedPrivateKey = loadedKey
	}

	return strategy, nil
}

func (s *PrivateKeyStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured || s.loadedPrivateKey == nil {
		return auth.NewTokenValidationResult(false, "", nil, auth.ErrProviderNotConfigured)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, auth.ErrInvalidTokenSignature
		}

		return &s.loadedPrivateKey.Key().PublicKey, nil
	})
	if err != nil {
		return auth.NewTokenValidationResult(false, "", nil, err)
	}

	if !token.Valid {
		return auth.NewTokenValidationResult(false, "", nil, auth.ErrTokenValidation)
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

	return auth.NewTokenValidationResult(true, userID, claims, nil)
}

func (s *PrivateKeyStrategy) Name() string {
	return s.name
}

func (s *PrivateKeyStrategy) IsConfigured() bool {
	return s.configured
}

func (s *PrivateKeyStrategy) Priority() int {
	return PrivateKeyPriority
}

func (s *PrivateKeyStrategy) GetKeyFormat() auth.KeyFormat {
	if s.loadedPrivateKey != nil {
		return s.loadedPrivateKey.Format()
	}

	return auth.KeyFormatUnknown
}
