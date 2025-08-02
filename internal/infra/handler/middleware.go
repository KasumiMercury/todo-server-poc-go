package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/KasumiMercury/todo-server-poc-go/internal/infra/auth"
)

type AuthenticationMiddleware struct {
	authService *auth.AuthenticationService
}

func NewAuthenticationMiddleware(authService *auth.AuthenticationService) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		authService: authService,
	}
}

func (m *AuthenticationMiddleware) MiddlewareFunc() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")

			tokenString, err := m.authService.ExtractTokenFromHeader(authHeader)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", strPtr(err.Error())))
			}

			result := m.authService.ValidateToken(tokenString)
			if !result.IsValid() {
				errorMessage := "Invalid token"
				if result.Error() != nil {
					errorMessage = "Invalid token: " + result.Error().Error()
				}

				return c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &errorMessage))
			}

			// Store authentication information in context
			c.Set("user_id", result.UserID())
			c.Set("jwt_claims", result.Claims())
			c.Set("auth_strategy", result.StrategyName())

			return next(c)
		}
	}
}

func strPtr(s string) *string {
	return &s
}
