package jwks

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
)

type Client struct {
	cache       *jwk.Cache
	endpoint    *auth.JWKsEndpoint
	cacheConfig *auth.JWKsCacheConfig
}

func NewClient(endpoint *auth.JWKsEndpoint, cacheConfig *auth.JWKsCacheConfig) (*Client, error) {
	if endpoint == nil {
		return nil, auth.ErrInvalidJWKsEndpoint
	}

	ctx := context.Background()
	cache := jwk.NewCache(ctx)

	err := cache.Register(endpoint.URL(), jwk.WithMinRefreshInterval(cacheConfig.RefreshPadding()))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to register JWKs endpoint - %v", auth.ErrJWKsClientError, err)
	}

	return &Client{
		cache:       cache,
		endpoint:    endpoint,
		cacheConfig: cacheConfig,
	}, nil
}

func (c *Client) ValidateToken(tokenString string) *auth.TokenValidationResult {
	ctx := context.Background()

	keySet, err := c.cache.Get(ctx, c.endpoint.URL())
	if err != nil {
		return auth.NewTokenValidationResult(false, "", nil,
			fmt.Errorf("%w: failed to get key set - %v", auth.ErrJWKsClientError, err))
	}

	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keySet))
	if err != nil {
		return auth.NewTokenValidationResult(false, "", nil,
			fmt.Errorf("%w: failed to parse token - %v", auth.ErrTokenValidation, err))
	}

	claims, err := token.AsMap(ctx)
	if err != nil {
		return auth.NewTokenValidationResult(false, "", nil,
			fmt.Errorf("%w: failed to get claims - %v", auth.ErrTokenValidation, err))
	}

	userID := ""
	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	}

	return auth.NewTokenValidationResult(true, userID, claims, nil)
}

func (c *Client) Refresh(ctx context.Context) error {
	_, err := c.cache.Refresh(ctx, c.endpoint.URL())
	if err != nil {
		return fmt.Errorf("%w: failed to refresh JWKs cache - %v", auth.ErrJWKsClientError, err)
	}
	return nil
}
