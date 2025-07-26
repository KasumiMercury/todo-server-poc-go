package auth

import "errors"

var (
	// ErrInvalidJWKsEndpoint JWKs related errors
	ErrInvalidJWKsEndpoint = errors.New("invalid JWKs endpoint URL")
	ErrJWKsClientError     = errors.New("JWKs client error")
	ErrKeyNotFound         = errors.New("key not found in JWKs")

	// ErrInvalidPrivateKeyFile Private key related errors
	ErrInvalidPrivateKeyFile   = errors.New("invalid private key file path")
	ErrPrivateKeyFileNotFound  = errors.New("private key file not found")
	ErrPrivateKeyFileReadError = errors.New("failed to read private key file")
	ErrPrivateKeyParseError    = errors.New("failed to parse private key")
	ErrUnsupportedKeyFormat    = errors.New("unsupported private key format")
	ErrPrivateKeyLoaderError   = errors.New("private key loader error")

	// ErrTokenValidation Token validation errors
	ErrTokenValidation       = errors.New("token validation failed")
	ErrInvalidTokenFormat    = errors.New("invalid token format")
	ErrTokenExpired          = errors.New("token has expired")
	ErrInvalidTokenSignature = errors.New("invalid token signature")

	// ErrNoValidProvider Authentication provider errors
	ErrNoValidProvider       = errors.New("no valid authentication provider found")
	ErrProviderNotConfigured = errors.New("authentication provider not configured")
	ErrAllProvidersFailed    = errors.New("all authentication providers failed")

	// ErrMissingAuthorizationHeader Authentication service errors
	ErrMissingAuthorizationHeader = errors.New("missing authorization header")
	ErrInvalidAuthorizationFormat = errors.New("invalid authorization header format")
)
