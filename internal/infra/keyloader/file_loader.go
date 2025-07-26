package keyloader

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

type FileLoader struct{}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (f *FileLoader) LoadPrivateKey(file *auth.PrivateKeyFile) (*auth.LoadedPrivateKey, error) {
	if file == nil {
		return nil, fmt.Errorf("%w: file cannot be nil", auth.ErrPrivateKeyLoaderError)
	}

	data, err := os.ReadFile(file.FilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", auth.ErrPrivateKeyFileNotFound, file.FilePath())
		}
		return nil, fmt.Errorf("%w: %v", auth.ErrPrivateKeyFileReadError, err)
	}

	// Try to load with different formats in priority order (easily configurable)
	formats := f.getFormatPriority()

	for _, format := range formats {
		if key, err := f.tryLoadWithFormat(data, format); err == nil {
			return auth.NewLoadedPrivateKey(key, format), nil
		}
	}

	return nil, fmt.Errorf("%w: unable to parse private key with any supported format", auth.ErrPrivateKeyParseError)
}

// getFormatPriority returns the format priority order (configurable)
func (f *FileLoader) getFormatPriority() []auth.KeyFormat {
	return []auth.KeyFormat{
		auth.KeyFormatRSAPEM,   // Highest priority (RSA PEM)
		auth.KeyFormatPKCS8PEM, // Second priority
		auth.KeyFormatRSADER,   // Third priority
		auth.KeyFormatPKCS8DER, // Lowest priority
	}
}

func (f *FileLoader) SupportedFormats() []auth.KeyFormat {
	return f.getFormatPriority()
}

func (f *FileLoader) tryLoadWithFormat(data []byte, format auth.KeyFormat) (*rsa.PrivateKey, error) {
	switch format {
	case auth.KeyFormatRSAPEM:
		return f.loadRSAPEM(data)
	case auth.KeyFormatPKCS8PEM:
		return f.loadPKCS8PEM(data)
	case auth.KeyFormatRSADER:
		return f.loadRSADER(data)
	case auth.KeyFormatPKCS8DER:
		return f.loadPKCS8DER(data)
	default:
		return nil, fmt.Errorf("%w: format %s", auth.ErrUnsupportedKeyFormat, format.String())
	}
}

func (f *FileLoader) loadRSAPEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("not an RSA private key PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	return key, nil
}

func (f *FileLoader) loadPKCS8PEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("not a PKCS#8 private key PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %v", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	return rsaKey, nil
}

func (f *FileLoader) loadRSADER(data []byte) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA DER private key: %v", err)
	}

	return key, nil
}

func (f *FileLoader) loadPKCS8DER(data []byte) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 DER private key: %v", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	return rsaKey, nil
}
