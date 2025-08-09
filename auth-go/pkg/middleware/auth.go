package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/auth-go/pkg/oauth2"
	"github.com/erpmicroservices/auth-go/pkg/jwt"
)

// AuthMiddleware provides authentication middleware for HTTP requests
type AuthMiddleware struct {
	oauth2Client *oauth2.Client
	jwtParser    *jwt.Parser
	validator    *oauth2.Validator
	config       *AuthConfig
}

// AuthConfig holds configuration for authentication middleware
type AuthConfig struct {
	TokenHeader         string
	TokenQueryParam     string
	SkipPaths          []string
	RequireAuth        bool
	AllowBearerToken   bool
	AllowQueryToken    bool
	CacheUserInfo      bool
	UserInfoTTL        time.Duration
}

// DefaultAuthConfig returns default configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		TokenHeader:      "Authorization",
		TokenQueryParam:  "access_token",
		RequireAuth:      true,
		AllowBearerToken: true,
		AllowQueryToken:  false, // Disabled by default for security
		CacheUserInfo:    true,
		UserInfoTTL:      5 * time.Minute,
	}
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(oauth2Client *oauth2.Client, jwtParser *jwt.Parser, validator *oauth2.Validator, config *AuthConfig) *AuthMiddleware {
	if config == nil {
		config = DefaultAuthConfig()
	}
	
	return &AuthMiddleware{
		oauth2Client: oauth2Client,
		jwtParser:    jwtParser,
		validator:    validator,
		config:       config,
	}
}

// RequireAuth is a middleware that requires authentication for all requests
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be skipped
		if am.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		token, err := am.extractToken(r)
		if err != nil {
			log.Debug().Err(err).Str("path", r.URL.Path).Msg("Failed to extract token")
			am.unauthorized(w, "Authentication required")
			return
		}

		if token == "" {
			log.Debug().Str("path", r.URL.Path).Msg("No token provided")
			am.unauthorized(w, "Authentication required")
			return
		}

		// Validate the token
		validationResult, err := am.validator.ValidateToken(r.Context(), token)
		if err != nil {
			log.Warn().Err(err).Str("path", r.URL.Path).Msg("Token validation failed")
			am.unauthorized(w, "Invalid token")
			return
		}

		if !validationResult.Valid {
			log.Warn().Str("path", r.URL.Path).Msg("Token is not valid")
			am.unauthorized(w, "Invalid token")
			return
		}

		// Add authentication context to request
		ctx := am.addAuthContextToRequest(r.Context(), validationResult, token)
		
		log.Debug().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("subject", am.getSubjectFromResult(validationResult)).
			Msg("Request authenticated successfully")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that optionally validates authentication
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := am.extractToken(r)
		if err != nil || token == "" {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate the token
		validationResult, err := am.validator.ValidateToken(r.Context(), token)
		if err != nil || !validationResult.Valid {
			// Invalid token, continue without authentication
			log.Debug().Err(err).Str("path", r.URL.Path).Msg("Optional authentication failed")
			next.ServeHTTP(w, r)
			return
		}

		// Add authentication context to request
		ctx := am.addAuthContextToRequest(r.Context(), validationResult, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts the authentication token from the request
func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	// Try Authorization header first
	if am.config.AllowBearerToken {
		authHeader := r.Header.Get(am.config.TokenHeader)
		if authHeader != "" {
			// Check for Bearer token
			if strings.HasPrefix(authHeader, "Bearer ") {
				return strings.TrimPrefix(authHeader, "Bearer "), nil
			}
			// Return the full header value (might be other token types)
			return authHeader, nil
		}
	}

	// Try query parameter
	if am.config.AllowQueryToken {
		token := r.URL.Query().Get(am.config.TokenQueryParam)
		if token != "" {
			return token, nil
		}
	}

	return "", nil
}

// shouldSkipPath checks if the given path should skip authentication
func (am *AuthMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range am.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// addAuthContextToRequest adds authentication information to the request context
func (am *AuthMiddleware) addAuthContextToRequest(ctx context.Context, result *oauth2.ValidationResult, token string) context.Context {
	authCtx := &AuthContext{
		Authenticated: true,
		Token:         token,
		ValidationResult: result,
	}

	// Extract user information based on validation type
	if result.JWT != nil && result.Claims != nil {
		// JWT validation was successful
		authCtx.Subject = result.Claims.Subject
		authCtx.Email = result.Claims.Email
		authCtx.Name = result.Claims.Name
		authCtx.Roles = result.Claims.Roles
		authCtx.Authorities = result.Claims.Authorities
		authCtx.OrganizationID = result.Claims.OrganizationID
		authCtx.DepartmentID = result.Claims.DepartmentID
	} else if result.Introspection != nil {
		// Introspection validation was successful
		authCtx.Subject = result.Introspection.Subject
		authCtx.Username = result.Introspection.Username
		authCtx.Roles = result.Introspection.Roles
		authCtx.Authorities = result.Introspection.Authorities
		authCtx.ClientID = result.Introspection.ClientID
	}

	return context.WithValue(ctx, AuthContextKeyValue, authCtx)
}

// getSubjectFromResult extracts subject from validation result for logging
func (am *AuthMiddleware) getSubjectFromResult(result *oauth2.ValidationResult) string {
	if result.Claims != nil {
		return result.Claims.Subject
	}
	if result.Introspection != nil {
		return result.Introspection.Subject
	}
	return "unknown"
}

// unauthorized sends an unauthorized response
func (am *AuthMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","message":"` + message + `"}`))
}

// AuthContext holds authentication information for a request
type AuthContext struct {
	Authenticated    bool
	Token           string
	Subject         string
	Username        string
	Email           string
	Name            string
	Roles           []string
	Authorities     []string
	OrganizationID  string
	DepartmentID    string
	ClientID        string
	ValidationResult *oauth2.ValidationResult
}

// HasRole checks if the authenticated user has a specific role
func (ac *AuthContext) HasRole(role string) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the authenticated user has any of the specified roles
func (ac *AuthContext) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if ac.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAuthority checks if the authenticated user has a specific authority
func (ac *AuthContext) HasAuthority(authority string) bool {
	for _, a := range ac.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}

// HasAnyAuthority checks if the authenticated user has any of the specified authorities
func (ac *AuthContext) HasAnyAuthority(authorities ...string) bool {
	for _, authority := range authorities {
		if ac.HasAuthority(authority) {
			return true
		}
	}
	return false
}

// IsServiceAccount returns true if this is a service account
func (ac *AuthContext) IsServiceAccount() bool {
	return ac.HasAuthority("SERVICE") || ac.Subject == ac.ClientID
}

// GetDisplayName returns a suitable display name for the user
func (ac *AuthContext) GetDisplayName() string {
	if ac.Name != "" {
		return ac.Name
	}
	if ac.Username != "" {
		return ac.Username
	}
	if ac.Email != "" {
		return ac.Email
	}
	return ac.Subject
}

// AuthContextKey is the context key for authentication context
type AuthContextKey string

const AuthContextKeyValue AuthContextKey = "auth"

// GetAuthContext retrieves the authentication context from a request context
func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(AuthContextKeyValue).(*AuthContext)
	return authCtx, ok
}

// MustGetAuthContext retrieves the authentication context or panics if not found
func MustGetAuthContext(ctx context.Context) *AuthContext {
	authCtx, ok := GetAuthContext(ctx)
	if !ok {
		panic("authentication context not found - ensure auth middleware is applied")
	}
	return authCtx
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	authCtx, ok := GetAuthContext(ctx)
	return ok && authCtx.Authenticated
}

// GetSubject retrieves the subject from the authentication context
func GetSubject(ctx context.Context) string {
	if authCtx, ok := GetAuthContext(ctx); ok {
		return authCtx.Subject
	}
	return ""
}

// GetRoles retrieves the roles from the authentication context
func GetRoles(ctx context.Context) []string {
	if authCtx, ok := GetAuthContext(ctx); ok {
		return authCtx.Roles
	}
	return nil
}

// GetAuthorities retrieves the authorities from the authentication context
func GetAuthorities(ctx context.Context) []string {
	if authCtx, ok := GetAuthContext(ctx); ok {
		return authCtx.Authorities
	}
	return nil
}

// HasRole checks if the authenticated user has a specific role
func HasRole(ctx context.Context, role string) bool {
	if authCtx, ok := GetAuthContext(ctx); ok {
		return authCtx.HasRole(role)
	}
	return false
}

// HasAuthority checks if the authenticated user has a specific authority
func HasAuthority(ctx context.Context, authority string) bool {
	if authCtx, ok := GetAuthContext(ctx); ok {
		return authCtx.HasAuthority(authority)
	}
	return false
}