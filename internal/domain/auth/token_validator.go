package auth

type TokenValidator interface {
	ValidateToken(tokenString string) *TokenValidationResult
	Name() string
}

type AuthenticationStrategy interface {
	TokenValidator
	IsConfigured() bool
	Priority() int
}

type AuthenticationResult struct {
	Strategy AuthenticationStrategy
	Result   *TokenValidationResult
}

func NewAuthenticationResult(strategy AuthenticationStrategy, result *TokenValidationResult) *AuthenticationResult {
	return &AuthenticationResult{
		Strategy: strategy,
		Result:   result,
	}
}

func (a *AuthenticationResult) IsValid() bool {
	return a.Result.IsValid()
}

func (a *AuthenticationResult) UserID() string {
	return a.Result.UserID()
}

func (a *AuthenticationResult) Claims() map[string]interface{} {
	return a.Result.Claims()
}

func (a *AuthenticationResult) Error() error {
	return a.Result.Error()
}

func (a *AuthenticationResult) StrategyName() string {
	return a.Strategy.Name()
}
