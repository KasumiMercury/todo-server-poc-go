package providers

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/keyloader"
)

const (
	// PrivateKeyStrategyName is the identifier for the private key authentication strategy.
	PrivateKeyStrategyName = "PrivateKey"
	// PrivateKeyPriority defines the priority of the private key strategy (highest priority).
	PrivateKeyPriority = 300 // Highest priority
)

// PrivateKeyStrategy implements JWT token validation using a private key file.
// It validates tokens by verifying their signature against the private key's public key.
type PrivateKeyStrategy struct {
	name             string
	configured       bool
	privateKeyLoader auth.PrivateKeyLoader
	loadedPrivateKey *auth.LoadedPrivateKey
}

// NewPrivateKeyStrategy creates a new PrivateKeyStrategy instance from the provided configuration.
// It loads the private key from the specified file path if provided.
func NewPrivateKeyStrategy(cfg config.Config) (*PrivateKeyStrategy, error) {
	if cfg.Auth.PrivateKeyFilePath != "" {
		privateKeyLoader := keyloader.NewFileLoader()

		privateKeyFile, err := auth.NewPrivateKeyFile(cfg.Auth.PrivateKeyFilePath)
		if err != nil {
			return nil, err
		}

		loadedKey, err := privateKeyLoader.LoadPrivateKey(privateKeyFile)
		if err != nil {
			return nil, err
		}

		return &PrivateKeyStrategy{
			name:             PrivateKeyStrategyName,
			configured:       true,
			privateKeyLoader: privateKeyLoader,
			loadedPrivateKey: loadedKey,
		}, nil
	}

	return &PrivateKeyStrategy{
		name:             PrivateKeyStrategyName,
		configured:       false,
		privateKeyLoader: nil,
		loadedPrivateKey: nil,
	}, nil
}

// ValidateToken validates a JWT token using the loaded private key.
// It verifies the token signature using RSA signing method and extracts the user ID from the 'sub' claim.
func (s *PrivateKeyStrategy) ValidateToken(tokenString string) *auth.TokenValidationResult {
	if !s.configured || s.loadedPrivateKey == nil {
		return auth.NewTokenValidationResult(false, "", auth.ErrProviderNotConfigured)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, auth.ErrInvalidTokenSignature
		}

		return &s.loadedPrivateKey.Key().PublicKey, nil
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

func (s *PrivateKeyStrategy) Name() string {
	return s.name
}

// IsConfigured returns whether this strategy is properly configured with a private key.
func (s *PrivateKeyStrategy) IsConfigured() bool {
	return s.configured
}

func (s *PrivateKeyStrategy) Priority() int {
	return PrivateKeyPriority
}

// GetKeyFormat returns the format of the loaded private key.
func (s *PrivateKeyStrategy) GetKeyFormat() auth.KeyFormat {
	if s.loadedPrivateKey != nil {
		return s.loadedPrivateKey.Format()
	}

	return auth.KeyFormatUnknown
}
