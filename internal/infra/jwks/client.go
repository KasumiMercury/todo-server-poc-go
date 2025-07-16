package jwks

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWK struct {
	Kty string   `json:"kty"`
	Use string   `json:"use"`
	Kid string   `json:"kid"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type JWKSet struct {
	Keys []JWK `json:"keys"`
}

type Client struct {
	jwksURL   string
	keys      map[string]*rsa.PublicKey
	mutex     sync.RWMutex
	lastFetch time.Time
	cacheTTL  time.Duration
}

func NewClient(jwksURL string) (*Client, error) {
	if jwksURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	client := &Client{
		jwksURL:  jwksURL,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: time.Hour,
	}

	if err := client.fetchKeys(); err != nil {
		return nil, fmt.Errorf("failed to fetch initial keys: %w", err)
	}

	return client, nil
}

func (c *Client) GetKeyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		c.mutex.RLock()
		needsRefresh := time.Since(c.lastFetch) > c.cacheTTL
		key, exists := c.keys[kid]
		c.mutex.RUnlock()

		if needsRefresh || !exists {
			if err := c.fetchKeys(); err != nil {
				return nil, fmt.Errorf("failed to refresh keys: %w", err)
			}

			c.mutex.RLock()
			key, exists = c.keys[kid]
			c.mutex.RUnlock()
		}

		if !exists {
			return nil, fmt.Errorf("key not found for kid: %s", kid)
		}

		return key, nil
	}
}

func (c *Client) fetchKeys() error {
	resp, err := http.Get(c.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	var jwkSet JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwkSet); err != nil {
		return fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwkSet.Keys {
		if jwk.Kty != "RSA" {
			continue
		}

		key, err := c.parseRSAKey(jwk)
		if err != nil {
			continue
		}

		keys[jwk.Kid] = key
	}

	c.mutex.Lock()
	c.keys = keys
	c.lastFetch = time.Now()
	c.mutex.Unlock()

	return nil
}

func (c *Client) parseRSAKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	if len(jwk.X5c) > 0 {
		certBytes, err := base64.StdEncoding.DecodeString(jwk.X5c[0])
		if err == nil {
			cert, err := x509.ParseCertificate(certBytes)
			if err == nil {
				if rsaKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
					return rsaKey, nil
				}
			}
		}
	}

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
