package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	infraAuth "github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/keyloader"
)

func TestPrivateKeyFileLoader(t *testing.T) {
	tests := []struct {
		name     string
		keyType  string
		expected auth.KeyFormat
	}{
		{"RSA PEM format", "rsa_pem", auth.KeyFormatRSAPEM},
		{"PKCS#8 PEM format", "pkcs8_pem", auth.KeyFormatPKCS8PEM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			tempFile, err := os.CreateTemp("", "test_key_*.pem")
			require.NoError(t, err)

			defer func() {
				if err := os.Remove(tempFile.Name()); err != nil {
					t.Logf("Failed to remove temp file %s: %v", tempFile.Name(), err)
				}
			}()

			var (
				keyBytes  []byte
				blockType string
			)

			switch tt.keyType {
			case "rsa_pem":
				keyBytes = x509.MarshalPKCS1PrivateKey(privateKey)
				blockType = "RSA PRIVATE KEY"
			case "pkcs8_pem":
				keyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
				require.NoError(t, err)

				blockType = "PRIVATE KEY"
			}

			pemBlock := &pem.Block{
				Type:  blockType,
				Bytes: keyBytes,
			}

			err = pem.Encode(tempFile, pemBlock)
			require.NoError(t, err)

			if err := tempFile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Test loading
			loader := keyloader.NewFileLoader()
			privateKeyFile, err := auth.NewPrivateKeyFile(tempFile.Name())
			require.NoError(t, err)

			loadedKey, err := loader.LoadPrivateKey(privateKeyFile)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, loadedKey.Format())
			assert.NotNil(t, loadedKey.Key())
			assert.Equal(t, privateKey.N, loadedKey.Key().N)
		})
	}
}

func TestPrivateKeyFileLoader_Errors(t *testing.T) {
	loader := keyloader.NewFileLoader()

	t.Run("file not found", func(t *testing.T) {
		privateKeyFile, err := auth.NewPrivateKeyFile("/nonexistent/path")
		require.NoError(t, err)

		_, err = loader.LoadPrivateKey(privateKeyFile)
		assert.ErrorIs(t, err, auth.ErrPrivateKeyFileNotFound)
	})

	t.Run("invalid file format", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "invalid_key_*")
		require.NoError(t, err)

		defer func() {
			if err := os.Remove(tempFile.Name()); err != nil {
				t.Logf("Failed to remove temp file %s: %v", tempFile.Name(), err)
			}
		}()

		_, err = tempFile.WriteString("invalid key content")
		require.NoError(t, err)

		if err := tempFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		privateKeyFile, err := auth.NewPrivateKeyFile(tempFile.Name())
		require.NoError(t, err)

		_, err = loader.LoadPrivateKey(privateKeyFile)
		assert.ErrorIs(t, err, auth.ErrPrivateKeyParseError)
	})

	t.Run("nil file", func(t *testing.T) {
		_, err := loader.LoadPrivateKey(nil)
		assert.ErrorIs(t, err, auth.ErrPrivateKeyLoaderError)
	})
}

func TestAuthenticationServiceWithPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tempFile, err := os.CreateTemp("", "test_jwt_key_*.pem")
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Logf("Failed to remove temp file %s: %v", tempFile.Name(), err)
		}
	}()

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	err = pem.Encode(tempFile, pemBlock)
	require.NoError(t, err)

	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	cfg := config.Config{
		Auth: config.AuthConfig{
			PrivateKeyFilePath: tempFile.Name(),
			JWTSecret:          "fallback-secret",
		},
	}

	authService, err := infraAuth.NewAuthenticationService(cfg)
	require.NoError(t, err)

	t.Run("valid token with private key", func(t *testing.T) {
		testUserID := uuid.New().String()
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": testUserID,
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		result := authService.ValidateToken(tokenString)
		assert.True(t, result.IsValid())
		assert.Equal(t, testUserID, result.UserID())
		assert.Equal(t, "PrivateKey", result.StrategyName())
	})

	t.Run("invalid token", func(t *testing.T) {
		result := authService.ValidateToken("invalid.token.here")
		assert.False(t, result.IsValid())
		assert.NotNil(t, result.Error())
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": uuid.New().String(),
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
			"iat": time.Now().Add(-2 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		result := authService.ValidateToken(tokenString)
		assert.False(t, result.IsValid())
		assert.NotNil(t, result.Error())
	})
}

func TestAuthenticationPriority(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tempFile, err := os.CreateTemp("", "test_priority_key_*.pem")
	require.NoError(t, err)

	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Logf("Failed to remove temp file %s: %v", tempFile.Name(), err)
		}
	}()

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	err = pem.Encode(tempFile, pemBlock)
	require.NoError(t, err)

	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	t.Run("private key has priority over string secret", func(t *testing.T) {
		cfg := config.Config{
			Auth: config.AuthConfig{
				PrivateKeyFilePath: tempFile.Name(),
				JWTSecret:          "string-secret",
			},
		}

		authService, err := infraAuth.NewAuthenticationService(cfg)
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "private-key-user",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		privateKeyToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		result := authService.ValidateToken(privateKeyToken)
		assert.True(t, result.IsValid())
		assert.Equal(t, "private-key-user", result.UserID())
		assert.Equal(t, "PrivateKey", result.StrategyName())
	})
}
