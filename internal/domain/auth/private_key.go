package auth

import (
	"crypto/rsa"
	"time"
)

// PrivateKeyFile represents a private key file location.
// It encapsulates the file path where a private key is stored.
type PrivateKeyFile struct {
	filePath string
}

// NewPrivateKeyFile creates a new PrivateKeyFile with the provided file path.
// It returns an error if the file path is empty.
func NewPrivateKeyFile(filePath string) (*PrivateKeyFile, error) {
	if filePath == "" {
		return nil, ErrInvalidPrivateKeyFile
	}

	return &PrivateKeyFile{filePath: filePath}, nil
}

// FilePath returns the file path of the private key file.
func (p *PrivateKeyFile) FilePath() string {
	return p.filePath
}

// KeyFormat represents the format of a private key file.
type KeyFormat int

const (
	KeyFormatRSAPEM KeyFormat = iota
	KeyFormatPKCS8PEM
	KeyFormatRSADER
	KeyFormatPKCS8DER
	KeyFormatECDSAPEM
	KeyFormatUnknown
)

// String returns the string representation of the KeyFormat.
func (k KeyFormat) String() string {
	switch k {
	case KeyFormatRSAPEM:
		return "RSA_PEM"
	case KeyFormatPKCS8PEM:
		return "PKCS8_PEM"
	case KeyFormatRSADER:
		return "RSA_DER"
	case KeyFormatPKCS8DER:
		return "PKCS8_DER"
	case KeyFormatECDSAPEM:
		return "ECDSA_PEM"
	case KeyFormatUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

// LoadedPrivateKey represents a successfully loaded private key with metadata.
type LoadedPrivateKey struct {
	key      *rsa.PrivateKey
	format   KeyFormat
	loadedAt time.Time
}

// NewLoadedPrivateKey creates a new LoadedPrivateKey with the provided key and format.
// It automatically sets the loaded timestamp to the current time.
func NewLoadedPrivateKey(key *rsa.PrivateKey, format KeyFormat) *LoadedPrivateKey {
	return &LoadedPrivateKey{
		key:      key,
		format:   format,
		loadedAt: time.Now(),
	}
}

func (l *LoadedPrivateKey) Key() *rsa.PrivateKey {
	return l.key
}

func (l *LoadedPrivateKey) Format() KeyFormat {
	return l.format
}

func (l *LoadedPrivateKey) LoadedAt() time.Time {
	return l.loadedAt
}

// PrivateKeyLoader defines the interface for loading private keys from files.
// It supports multiple key formats and provides information about supported formats.
type PrivateKeyLoader interface {
	// LoadPrivateKey loads a private key from the specified file.
	LoadPrivateKey(file *PrivateKeyFile) (*LoadedPrivateKey, error)
	// SupportedFormats returns the list of key formats supported by this loader.
	SupportedFormats() []KeyFormat
}
