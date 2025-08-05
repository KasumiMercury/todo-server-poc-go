package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrivateKeyFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		filePath      string
		expectedError error
	}{
		{
			name:          "valid file path",
			filePath:      "/path/to/private.key",
			expectedError: nil,
		},
		{
			name:          "valid relative path",
			filePath:      "./keys/private.pem",
			expectedError: nil,
		},
		{
			name:          "empty file path",
			filePath:      "",
			expectedError: ErrInvalidPrivateKeyFile,
		},
		{
			name:          "windows path",
			filePath:      "C:\\keys\\private.key",
			expectedError: nil,
		},
		{
			name:          "path with spaces",
			filePath:      "/path/to/my private key.pem",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			keyFile, err := NewPrivateKeyFile(tt.filePath)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, keyFile)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, keyFile)
				assert.Equal(t, tt.filePath, keyFile.FilePath())
			}
		})
	}
}

func TestPrivateKeyFile_FilePath(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedPath := "/etc/ssl/private/jwt-signing.key"
	keyFile, err := NewPrivateKeyFile(expectedPath)
	require.NoError(t, err)
	require.NotNil(t, keyFile)

	// Act
	actualPath := keyFile.FilePath()

	// Assert
	assert.Equal(t, expectedPath, actualPath)
}

func TestKeyFormat_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		format         KeyFormat
		expectedString string
	}{
		{
			name:           "RSA PEM format",
			format:         KeyFormatRSAPEM,
			expectedString: "RSA_PEM",
		},
		{
			name:           "PKCS8 PEM format",
			format:         KeyFormatPKCS8PEM,
			expectedString: "PKCS8_PEM",
		},
		{
			name:           "RSA DER format",
			format:         KeyFormatRSADER,
			expectedString: "RSA_DER",
		},
		{
			name:           "PKCS8 DER format",
			format:         KeyFormatPKCS8DER,
			expectedString: "PKCS8_DER",
		},
		{
			name:           "ECDSA PEM format",
			format:         KeyFormatECDSAPEM,
			expectedString: "ECDSA_PEM",
		},
		{
			name:           "unknown format",
			format:         KeyFormatUnknown,
			expectedString: "UNKNOWN",
		},
		{
			name:           "invalid format value",
			format:         KeyFormat(999),
			expectedString: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := tt.format.String()

			// Assert
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestNewLoadedPrivateKey(t *testing.T) {
	t.Parallel()

	// Generate a test RSA key
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name           string
		key            *rsa.PrivateKey
		format         KeyFormat
		expectedFormat KeyFormat
	}{
		{
			name:           "RSA key with PEM format",
			key:            testKey,
			format:         KeyFormatRSAPEM,
			expectedFormat: KeyFormatRSAPEM,
		},
		{
			name:           "RSA key with PKCS8 format",
			key:            testKey,
			format:         KeyFormatPKCS8PEM,
			expectedFormat: KeyFormatPKCS8PEM,
		},
		{
			name:           "RSA key with DER format",
			key:            testKey,
			format:         KeyFormatRSADER,
			expectedFormat: KeyFormatRSADER,
		},
		{
			name:           "RSA key with ECDSA format",
			key:            testKey,
			format:         KeyFormatECDSAPEM,
			expectedFormat: KeyFormatECDSAPEM,
		},
		{
			name:           "RSA key with unknown format",
			key:            testKey,
			format:         KeyFormatUnknown,
			expectedFormat: KeyFormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			beforeTime := time.Now()

			// Act
			loadedKey := NewLoadedPrivateKey(tt.key, tt.format)
			afterTime := time.Now()

			// Assert
			require.NotNil(t, loadedKey)
			assert.Equal(t, tt.key, loadedKey.Key())
			assert.Equal(t, tt.expectedFormat, loadedKey.Format())

			// Check that LoadedAt is between before and after times
			loadedAt := loadedKey.LoadedAt()
			assert.True(t, loadedAt.After(beforeTime) || loadedAt.Equal(beforeTime))
			assert.True(t, loadedAt.Before(afterTime) || loadedAt.Equal(afterTime))
		})
	}
}

func TestLoadedPrivateKey_Methods(t *testing.T) {
	t.Parallel()

	// Arrange
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	expectedFormat := KeyFormatPKCS8PEM
	beforeTime := time.Now()
	loadedKey := NewLoadedPrivateKey(testKey, expectedFormat)
	afterTime := time.Now()

	// Act & Assert
	assert.Equal(t, testKey, loadedKey.Key())
	assert.Equal(t, expectedFormat, loadedKey.Format())

	loadedAt := loadedKey.LoadedAt()
	assert.True(t, loadedAt.After(beforeTime) || loadedAt.Equal(beforeTime))
	assert.True(t, loadedAt.Before(afterTime) || loadedAt.Equal(afterTime))
}

func TestLoadedPrivateKey_KeyIntegrity(t *testing.T) {
	t.Parallel()

	// Arrange
	originalKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	loadedKey := NewLoadedPrivateKey(originalKey, KeyFormatRSAPEM)

	// Act
	retrievedKey := loadedKey.Key()

	// Assert
	assert.Equal(t, originalKey, retrievedKey)
	assert.Equal(t, originalKey.N, retrievedKey.N)
	assert.Equal(t, originalKey.E, retrievedKey.E)
	assert.Equal(t, originalKey.D, retrievedKey.D)
}

func TestLoadedPrivateKey_TimeConsistency(t *testing.T) {
	t.Parallel()

	// Arrange
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	loadedKey := NewLoadedPrivateKey(testKey, KeyFormatRSAPEM)

	// Act
	firstCall := loadedKey.LoadedAt()
	time.Sleep(1 * time.Millisecond) // Small delay
	secondCall := loadedKey.LoadedAt()

	// Assert - LoadedAt should return the same time on multiple calls
	assert.Equal(t, firstCall, secondCall)
}

func TestKeyFormatEnumValues(t *testing.T) {
	t.Parallel()

	// Arrange & Assert - Verify enum values are properly defined
	assert.Equal(t, KeyFormat(0), KeyFormatRSAPEM)
	assert.Equal(t, KeyFormat(1), KeyFormatPKCS8PEM)
	assert.Equal(t, KeyFormat(2), KeyFormatRSADER)
	assert.Equal(t, KeyFormat(3), KeyFormatPKCS8DER)
	assert.Equal(t, KeyFormat(4), KeyFormatECDSAPEM)
	assert.Equal(t, KeyFormat(5), KeyFormatUnknown)
}

// Mock implementation of PrivateKeyLoader for testing interface compliance
type mockPrivateKeyLoader struct {
	supportedFormats []KeyFormat
	loadError        error
	loadedKey        *LoadedPrivateKey
}

func (m *mockPrivateKeyLoader) LoadPrivateKey(file *PrivateKeyFile) (*LoadedPrivateKey, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	return m.loadedKey, nil
}

func (m *mockPrivateKeyLoader) SupportedFormats() []KeyFormat {
	return m.supportedFormats
}

func TestPrivateKeyLoader_Interface(t *testing.T) {
	t.Parallel()

	// Arrange
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	expectedFormats := []KeyFormat{KeyFormatRSAPEM, KeyFormatPKCS8PEM}
	expectedLoadedKey := NewLoadedPrivateKey(testKey, KeyFormatRSAPEM)

	loader := &mockPrivateKeyLoader{
		supportedFormats: expectedFormats,
		loadedKey:        expectedLoadedKey,
	}

	keyFile, err := NewPrivateKeyFile("/test/path/key.pem")
	require.NoError(t, err)

	// Act
	loadedKey, err := loader.LoadPrivateKey(keyFile)
	supportedFormats := loader.SupportedFormats()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedLoadedKey, loadedKey)
	assert.Equal(t, expectedFormats, supportedFormats)
}

func TestPrivateKeyLoader_InterfaceError(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedError := ErrPrivateKeyFileNotFound
	loader := &mockPrivateKeyLoader{
		loadError: expectedError,
	}

	keyFile, err := NewPrivateKeyFile("/nonexistent/key.pem")
	require.NoError(t, err)

	// Act
	loadedKey, err := loader.LoadPrivateKey(keyFile)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, loadedKey)
}
