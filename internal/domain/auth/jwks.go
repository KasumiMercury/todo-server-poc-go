package auth

import (
	"time"
)

type JWKsEndpoint struct {
	url string
}

func NewJWKsEndpoint(url string) (*JWKsEndpoint, error) {
	if url == "" {
		return nil, ErrInvalidJWKsEndpoint
	}

	return &JWKsEndpoint{url: url}, nil
}

func (j *JWKsEndpoint) URL() string {
	return j.url
}

type JWKsCacheConfig struct {
	cacheDuration  time.Duration
	refreshPadding time.Duration
}

func NewJWKsCacheConfig(cacheDuration, refreshPadding time.Duration) *JWKsCacheConfig {
	return &JWKsCacheConfig{
		cacheDuration:  cacheDuration,
		refreshPadding: refreshPadding,
	}
}

func (c *JWKsCacheConfig) CacheDuration() time.Duration {
	return c.cacheDuration
}

func (c *JWKsCacheConfig) RefreshPadding() time.Duration {
	return c.refreshPadding
}

type TokenValidationResult struct {
	isValid bool
	userID  string
	claims  map[string]interface{}
	err     error
}

func NewTokenValidationResult(isValid bool, userID string, claims map[string]interface{}, err error) *TokenValidationResult {
	return &TokenValidationResult{
		isValid: isValid,
		userID:  userID,
		claims:  claims,
		err:     err,
	}
}

func (r *TokenValidationResult) IsValid() bool {
	return r.isValid
}

func (r *TokenValidationResult) UserID() string {
	return r.userID
}

func (r *TokenValidationResult) Claims() map[string]interface{} {
	return r.claims
}

func (r *TokenValidationResult) Error() error {
	return r.err
}
