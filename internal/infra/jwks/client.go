package jwks

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

// Client provides JWT token validation using JSON Web Key Sets (JWKs).
// It manages a cache of public keys fetched from a JWKs endpoint for token verification.
type Client struct {
	cache       *jwk.Cache
	endpoint    *auth.JWKsEndpoint
	cacheConfig *auth.JWKsCacheConfig
}

// NewClient creates a new JWKs client with caching capabilities.
// It registers the JWKs endpoint with the cache for automatic key refresh.
func NewClient(endpoint *auth.JWKsEndpoint, cacheConfig *auth.JWKsCacheConfig) (*Client, error) {
	if endpoint == nil {
		return nil, auth.ErrInvalidJWKsEndpoint
	}

	ctx := context.Background()

	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create JWKs cache - %v", auth.ErrJWKsClientError, err)
	}

	if err := cache.Register(ctx, endpoint.URL(), jwk.WithMinInterval(cacheConfig.RefreshPadding())); err != nil {
		return nil, fmt.Errorf("%w: failed to register JWKs endpoint - %v", auth.ErrJWKsClientError, err)
	}

	return &Client{
		cache:       cache,
		endpoint:    endpoint,
		cacheConfig: cacheConfig,
	}, nil
}

// ValidateToken validates a JWT token using the cached JWKs.
// It fetches the appropriate public key from the cache and verifies the token signature.
func (c *Client) ValidateToken(tokenString string) *auth.TokenValidationResult {
	ctx := context.Background()

	keySet, err := c.cache.Lookup(ctx, c.endpoint.URL())
	if err != nil {
		return auth.NewTokenValidationResult(false, "",
			fmt.Errorf("%w: failed to get key set - %v", auth.ErrJWKsClientError, err))
	}

	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keySet))
	if err != nil {
		return auth.NewTokenValidationResult(false, "",
			fmt.Errorf("%w: failed to parse token - %v", auth.ErrTokenValidation, err))
	}

	userID := ""
	if sub, ok := token.Subject(); ok {
		userID = sub
	}

	return auth.NewTokenValidationResult(true, userID, nil)
}

// Refresh manually refreshes the JWKs cache from the endpoint.
func (c *Client) Refresh(ctx context.Context) error {
	_, err := c.cache.Refresh(ctx, c.endpoint.URL())
	if err != nil {
		return fmt.Errorf("%w: failed to refresh JWKs cache - %v", auth.ErrJWKsClientError, err)
	}

	return nil
}
