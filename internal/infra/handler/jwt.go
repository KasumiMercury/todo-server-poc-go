package handler

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

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
