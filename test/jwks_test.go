package test

import (
	"testing"
	"time"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

func TestNewJWKsEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL",
			url:         "https://example.com/.well-known/jwks.json",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := auth.NewJWKsEndpoint(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}

				if endpoint != nil {
					t.Errorf("expected nil endpoint but got %v", endpoint)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if endpoint == nil {
					t.Errorf("expected endpoint but got nil")
				}

				if endpoint.URL() != tt.url {
					t.Errorf("expected URL %s but got %s", tt.url, endpoint.URL())
				}
			}
		})
	}
}

func TestNewJWKsCacheConfig(t *testing.T) {
	cacheDuration := 3600 * time.Second
	refreshPadding := 300 * time.Second

	config := auth.NewJWKsCacheConfig(cacheDuration, refreshPadding)

	if config.CacheDuration() != cacheDuration {
		t.Errorf("expected cache duration %v but got %v", cacheDuration, config.CacheDuration())
	}

	if config.RefreshPadding() != refreshPadding {
		t.Errorf("expected refresh padding %v but got %v", refreshPadding, config.RefreshPadding())
	}
}

func TestTokenValidationResult(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user123",
		"iat": 1234567890,
	}

	result := auth.NewTokenValidationResult(true, "user123", claims, nil)

	if !result.IsValid() {
		t.Errorf("expected valid token")
	}

	if result.UserID() != "user123" {
		t.Errorf("expected user ID 'user123' but got '%s'", result.UserID())
	}

	if result.Claims()["sub"] != "user123" {
		t.Errorf("expected sub claim 'user123' but got '%v'", result.Claims()["sub"])
	}

	if result.Error() != nil {
		t.Errorf("expected no error but got %v", result.Error())
	}
}
