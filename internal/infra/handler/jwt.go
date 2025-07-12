package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates JWT tokens from Authorization header
func JWTMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			details := "Authorization header is required"
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			details := "Authorization header must start with 'Bearer '"
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			details := "Token is required"
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			details := "Invalid token: " + err.Error()
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			c.Abort()
			return
		}

		if !token.Valid {
			details := "Token is not valid"
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError("Unauthorized", &details))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Store user information in context for later use
			c.Set("user_id", claims["sub"])
			c.Set("jwt_claims", claims)
		}

		c.Next()
	}
}
