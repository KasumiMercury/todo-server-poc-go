package auth

import "errors"

var (
	ErrInvalidJWKsEndpoint     = errors.New("invalid JWKs endpoint URL")
	ErrJWKsClientError         = errors.New("JWKs client error")
	ErrTokenValidation         = errors.New("token validation failed")
	ErrKeyNotFound             = errors.New("key not found in JWKs")
	ErrInvalidPrivateKeyFile   = errors.New("invalid private key file path")
	ErrPrivateKeyFileNotFound  = errors.New("private key file not found")
	ErrPrivateKeyFileReadError = errors.New("failed to read private key file")
	ErrPrivateKeyParseError    = errors.New("failed to parse private key")
	ErrUnsupportedKeyFormat    = errors.New("unsupported private key format")
	ErrPrivateKeyLoaderError   = errors.New("private key loader error")
)
