package auth

import "errors"

var (
	ErrInvalidJWKsEndpoint = errors.New("invalid JWKs endpoint URL")
	ErrJWKsClientError     = errors.New("JWKs client error")
	ErrTokenValidation     = errors.New("token validation failed")
	ErrKeyNotFound         = errors.New("key not found in JWKs")
)
