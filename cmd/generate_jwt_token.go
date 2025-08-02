package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT token generation utility
func main() {
	// Set token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create JWT claims
	claims := jwt.MapClaims{
		"sub":  "test-user",           // Subject (User ID)
		"exp":  expirationTime.Unix(), // Expiration time
		"iat":  time.Now().Unix(),     // Issued at time
		"name": "Test User",           // Custom claim
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	secretKey := "secret-key-for-testing"

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error generating JWT token: %v\n", err)
		return
	}

	// Display results
	fmt.Println("=== JWT Token Generated ===")
	fmt.Printf("Token: %s\n", tokenString)
	fmt.Printf("Expires at: %s\n", expirationTime.Format(time.RFC3339))
	fmt.Println("\nUsage example:")
	fmt.Println("curl -H \"Authorization: Bearer " + tokenString + "\" http://localhost:8080/tasks")
	fmt.Println("\n=== Copy this token for testing ===")
	fmt.Println(tokenString)
}
