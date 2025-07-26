package handler

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/auth"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/jwks"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/keyloader"
)

type JWTService struct {
	secretKey        string
	jwksClient       *jwks.Client
	privateKeyLoader auth.PrivateKeyLoader
	loadedPrivateKey *auth.LoadedPrivateKey
	useJWKs          bool
	usePrivateKey    bool
}

func NewJWTService(cfg config.Config) (*JWTService, error) {
	service := &JWTService{
		secretKey:     cfg.JWTSecret,
		useJWKs:       cfg.JWKs.EndpointURL != "",
		usePrivateKey: cfg.PrivateKeyFilePath != "",
	}

	// Priority 1: Private key file
	if service.usePrivateKey {
		service.privateKeyLoader = keyloader.NewFileLoader()

		privateKeyFile, err := auth.NewPrivateKeyFile(cfg.PrivateKeyFilePath)
		if err != nil {
			return nil, err
		}

		loadedKey, err := service.privateKeyLoader.LoadPrivateKey(privateKeyFile)
		if err != nil {
			return nil, err
		}

		service.loadedPrivateKey = loadedKey
	}

	// Priority 2: JWKs endpoint
	if service.useJWKs {
		endpoint, err := auth.NewJWKsEndpoint(cfg.JWKs.EndpointURL)
		if err != nil {
			return nil, err
		}

		cacheConfig := auth.NewJWKsCacheConfig(
			time.Duration(cfg.JWKs.CacheDuration)*time.Second,
			time.Duration(cfg.JWKs.RefreshPadding)*time.Second,
		)

		jwksClient, err := jwks.NewClient(endpoint, cacheConfig)
		if err != nil {
			return nil, err
		}

		service.jwksClient = jwksClient
	}

	return service, nil
}

func (j *JWTService) ValidateToken(tokenString string) *auth.TokenValidationResult {
	// Priority 1: Private key file
	if j.usePrivateKey && j.loadedPrivateKey != nil {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &j.loadedPrivateKey.Key().PublicKey, nil
		})

		if err == nil && token.Valid {
			claims := make(map[string]interface{})
			if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
				for k, v := range mapClaims {
					claims[k] = v
				}
			}

			userID := ""
			if sub, ok := claims["sub"].(string); ok {
				userID = sub
			}

			return auth.NewTokenValidationResult(true, userID, claims, nil)
		}
	}

	// Priority 2: JWKs endpoint
	if j.useJWKs && j.jwksClient != nil {
		result := j.jwksClient.ValidateToken(tokenString)
		if result.IsValid() {
			return result
		}
	}

	// Priority 3: String secret (fallback)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil || !token.Valid {
		return auth.NewTokenValidationResult(false, "", nil, err)
	}

	claims := make(map[string]interface{})
	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
		for k, v := range mapClaims {
			claims[k] = v
		}
	}

	userID := ""
	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	}

	return auth.NewTokenValidationResult(true, userID, claims, nil)
}

// JWTMiddleware returns an Echo JWT middleware for validating JWT tokens
func JWTMiddleware(secretKey string) echo.MiddlewareFunc {
	jwtCfg := echojwt.Config{
		SigningKey: []byte(secretKey),
		ErrorHandler: func(c echo.Context, err error) error {
			details := "Invalid token: " + err.Error()
			return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
		},
		SuccessHandler: func(c echo.Context) {
			// Extract user information from JWT token
			user := c.Get("user").(*jwt.Token)
			if claims, ok := user.Claims.(jwt.MapClaims); ok {
				// Store user information in context for later use
				c.Set("user_id", claims["sub"])
				c.Set("jwt_claims", claims)
			}
		},
	}
	return echojwt.WithConfig(jwtCfg)
}

func JWTMiddlewareWithService(jwtService *JWTService) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				details := "Missing authorization header"
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			}

			tokenString := ""
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				details := "Invalid authorization header format"
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			}

			result := jwtService.ValidateToken(tokenString)
			if !result.IsValid() {
				details := "Invalid token: " + result.Error().Error()
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			}

			c.Set("user_id", result.UserID())
			c.Set("jwt_claims", result.Claims())

			return next(c)
		}
	})
}
