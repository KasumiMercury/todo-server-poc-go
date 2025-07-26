package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/handler"
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
			// Generate test RSA key
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			// Create temporary file
			tempFile, err := ioutil.TempFile("", "test_key_*.pem")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())

			// Write key in specified format
			var keyBytes []byte
			var blockType string

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
			tempFile.Close()

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
		tempFile, err := ioutil.TempFile("", "invalid_key_*")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString("invalid key content")
		require.NoError(t, err)
		tempFile.Close()

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

func TestJWTServiceWithPrivateKey(t *testing.T) {
	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create temporary key file
	tempFile, err := ioutil.TempFile("", "test_jwt_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write RSA PEM key
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	err = pem.Encode(tempFile, pemBlock)
	require.NoError(t, err)
	tempFile.Close()

	// Create JWT service with private key file
	cfg := config.Config{
		PrivateKeyFilePath: tempFile.Name(),
		JWTSecret:          "fallback-secret",
	}

	jwtService, err := handler.NewJWTService(cfg)
	require.NoError(t, err)

	t.Run("valid token with private key", func(t *testing.T) {
		// Create a valid JWT token signed with the private key
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "test-user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Validate using JWT service
		result := jwtService.ValidateToken(tokenString)
		assert.True(t, result.IsValid())
		assert.Equal(t, "test-user-123", result.UserID())
		assert.NotNil(t, result.Claims())
	})

	t.Run("invalid token", func(t *testing.T) {
		result := jwtService.ValidateToken("invalid.token.here")
		assert.False(t, result.IsValid())
		assert.NotNil(t, result.Error())
	})

	t.Run("expired token", func(t *testing.T) {
		// Create an expired token
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "test-user-123",
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
			"iat": time.Now().Add(-2 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		result := jwtService.ValidateToken(tokenString)
		assert.False(t, result.IsValid())
		assert.NotNil(t, result.Error())
	})
}

func TestAuthenticationPriority(t *testing.T) {
	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create temporary key file
	tempFile, err := ioutil.TempFile("", "test_priority_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	err = pem.Encode(tempFile, pemBlock)
	require.NoError(t, err)
	tempFile.Close()

	t.Run("private key has priority over string secret", func(t *testing.T) {
		cfg := config.Config{
			PrivateKeyFilePath: tempFile.Name(),
			JWTSecret:          "string-secret",
		}

		jwtService, err := handler.NewJWTService(cfg)
		require.NoError(t, err)

		// Create token signed with private key
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "private-key-user",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		privateKeyToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// Should validate successfully with private key
		result := jwtService.ValidateToken(privateKeyToken)
		assert.True(t, result.IsValid())
		assert.Equal(t, "private-key-user", result.UserID())
	})
}
