package auth

// TokenValidator defines the interface for JWT token validation.
// It provides methods to validate tokens and identify the validator.
type TokenValidator interface {
	// ValidateToken validates a JWT token string and returns the validation result.
	ValidateToken(tokenString string) *TokenValidationResult
	// Name returns the name of the token validator.
	Name() string
}

// AuthenticationStrategy defines the interface for authentication strategies.
type AuthenticationStrategy interface {
	TokenValidator
	// IsConfigured returns whether the authentication strategy is properly configured.
	IsConfigured() bool
	// Priority returns the priority of this authentication strategy (higher values have higher priority).
	Priority() int
}

// AuthenticationResult represents the result of an authentication attempt.
type AuthenticationResult struct {
	Strategy AuthenticationStrategy
	Result   *TokenValidationResult
}

// NewAuthenticationResult creates a new AuthenticationResult with the given strategy and validation result.
func NewAuthenticationResult(strategy AuthenticationStrategy, result *TokenValidationResult) *AuthenticationResult {
	return &AuthenticationResult{
		Strategy: strategy,
		Result:   result,
	}
}

// IsValid returns whether the authentication result is valid.
func (a *AuthenticationResult) IsValid() bool {
	return a.Result.IsValid()
}

// UserID returns the user ID from the authentication result.
func (a *AuthenticationResult) UserID() string {
	return a.Result.UserID()
}

// Error returns any error that occurred during authentication.
func (a *AuthenticationResult) Error() error {
	return a.Result.Error()
}

// StrategyName returns the name of the authentication strategy used.
func (a *AuthenticationResult) StrategyName() string {
	return a.Strategy.Name()
}
