package auth

import (
	"time"
)

// JWKsEndpoint represents a JSON Web Key Set endpoint configuration.
// It encapsulates the URL where JWKs can be fetched for token validation.
type JWKsEndpoint struct {
	url string
}

// NewJWKsEndpoint creates a new JWKsEndpoint with the provided URL.
// It returns an error if the URL is empty.
func NewJWKsEndpoint(url string) (*JWKsEndpoint, error) {
	if url == "" {
		return nil, ErrInvalidJWKsEndpoint
	}

	return &JWKsEndpoint{url: url}, nil
}

// URL returns the JWKs endpoint URL.
func (j *JWKsEndpoint) URL() string {
	return j.url
}

// JWKsCacheConfig holds configuration for JWKs caching behavior.
// It defines how long keys are cached and when to refresh them.
type JWKsCacheConfig struct {
	cacheDuration  time.Duration
	refreshPadding time.Duration
}

// NewJWKsCacheConfig creates a new JWKsCacheConfig with the specified durations.
func NewJWKsCacheConfig(cacheDuration, refreshPadding time.Duration) *JWKsCacheConfig {
	return &JWKsCacheConfig{
		cacheDuration:  cacheDuration,
		refreshPadding: refreshPadding,
	}
}

// CacheDuration returns the duration for which keys are cached.
func (c *JWKsCacheConfig) CacheDuration() time.Duration {
	return c.cacheDuration
}

// RefreshPadding returns the minimum interval between cache refreshes.
func (c *JWKsCacheConfig) RefreshPadding() time.Duration {
	return c.refreshPadding
}

// TokenValidationResult represents the result of JWT token validation.
// It contains the validation status, extracted user ID, and any error that occurred.
type TokenValidationResult struct {
	isValid bool
	userID  string
	err     error
}

// NewTokenValidationResult creates a new TokenValidationResult with the provided values.
func NewTokenValidationResult(isValid bool, userID string, err error) *TokenValidationResult {
	return &TokenValidationResult{
		isValid: isValid,
		userID:  userID,
		err:     err,
	}
}

func (r *TokenValidationResult) IsValid() bool {
	return r.isValid
}

func (r *TokenValidationResult) UserID() string {
	return r.userID
}

func (r *TokenValidationResult) Error() error {
	return r.err
}
