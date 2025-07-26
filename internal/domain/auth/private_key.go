package auth

import (
	"crypto/rsa"
	"time"
)

type PrivateKeyFile struct {
	filePath string
}

func NewPrivateKeyFile(filePath string) (*PrivateKeyFile, error) {
	if filePath == "" {
		return nil, ErrInvalidPrivateKeyFile
	}
	return &PrivateKeyFile{filePath: filePath}, nil
}

func (p *PrivateKeyFile) FilePath() string {
	return p.filePath
}

type KeyFormat int

const (
	KeyFormatRSAPEM KeyFormat = iota
	KeyFormatPKCS8PEM
	KeyFormatRSADER
	KeyFormatPKCS8DER
	KeyFormatECDSAPEM
	KeyFormatUnknown
)

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
	default:
		return "UNKNOWN"
	}
}

type LoadedPrivateKey struct {
	key      *rsa.PrivateKey
	format   KeyFormat
	loadedAt time.Time
}

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

type PrivateKeyLoader interface {
	LoadPrivateKey(file *PrivateKeyFile) (*LoadedPrivateKey, error)
	SupportedFormats() []KeyFormat
}
