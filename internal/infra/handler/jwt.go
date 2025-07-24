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
)

type JWTService struct {
	secretKey  string
	jwksClient *jwks.Client
	useJWKs    bool
}

func NewJWTService(cfg config.Config) (*JWTService, error) {
	service := &JWTService{
		secretKey: cfg.JWTSecret,
		useJWKs:   cfg.JWKs.EndpointURL != "",
	}

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

func (j *JWTService) validateToken(tokenString string) *auth.TokenValidationResult {
	if j.useJWKs && j.jwksClient != nil {
		result := j.jwksClient.ValidateToken(tokenString)
		if result.IsValid() {
			return result
		}
	}

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
	config := echojwt.Config{
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
	return echojwt.WithConfig(config)
}

func JWTMiddlewareWithService(jwtService *JWTService) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				details := "Missing authorization header"
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			}

			tokenString := ""
			if len(auth) > 7 && auth[:7] == "Bearer " {
				tokenString = auth[7:]
			} else {
				details := "Invalid authorization header format"
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			}

			result := jwtService.validateToken(tokenString)
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
