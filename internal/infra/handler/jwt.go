package handler

import (
	"net/http"

	"github.com/KasumiMercury/todo-server-poc-go/internal/config"
	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/jwks"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware returns an Echo JWT middleware for validating JWT tokens
func JWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	var jwtConfig echojwt.Config

	switch cfg.JWTVerificationMode {
	case "jwks":
		if cfg.JWKsURL == "" {
			panic("JWKS URL is required when using JWKS verification mode")
		}

		jwksClient, err := jwks.NewClient(cfg.JWKsURL)
		if err != nil {
			panic("Failed to create JWKS client: " + err.Error())
		}

		jwtConfig = echojwt.Config{
			KeyFunc: jwksClient.GetKeyFunc(),
			ErrorHandler: func(c echo.Context, err error) error {
				details := "Invalid token: " + err.Error()
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			},
			SuccessHandler: func(c echo.Context) {
				user := c.Get("user").(*jwt.Token)
				if claims, ok := user.Claims.(jwt.MapClaims); ok {
					c.Set("user_id", claims["sub"])
					c.Set("jwt_claims", claims)
				}
			},
		}
	default: // "secret" mode
		jwtConfig = echojwt.Config{
			SigningKey: []byte(cfg.JWTSecret),
			ErrorHandler: func(c echo.Context, err error) error {
				details := "Invalid token: " + err.Error()
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			},
			SuccessHandler: func(c echo.Context) {
				user := c.Get("user").(*jwt.Token)
				if claims, ok := user.Claims.(jwt.MapClaims); ok {
					c.Set("user_id", claims["sub"])
					c.Set("jwt_claims", claims)
				}
			},
		}
	}

	return echojwt.WithConfig(jwtConfig)
}
