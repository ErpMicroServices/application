// Package middleware provides HTTP middleware helpers for common concerns in the ERP microservices system.
// It includes authentication, authorization, logging, CORS, rate limiting, and error handling middleware.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/erpmicroservices/common-go/pkg/errors"
	"github.com/erpmicroservices/common-go/pkg/logging"
	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// Middleware represents a middleware function.
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middleware functions together.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// CORS Configuration
type CORSConfig struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAge             int
	OptionsPassthrough bool
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:     []string{},
		AllowCredentials:   false,
		MaxAge:             86400, // 24 hours
		OptionsPassthrough: false,
	}
}

// CORS creates a CORS middleware with the given configuration.
func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && len(config.AllowedOrigins) > 0 {
				allowed := false
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ","))
			}

			// Handle preflight request
			if r.Method == "OPTIONS" {
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
				}
				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
				}
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}

				if !config.OptionsPassthrough {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogging creates a request logging middleware.
func RequestLogging(logger *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Create response writer wrapper to capture status and size
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Add request context
			ctx := logging.WithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapper, r)

			// Log request
			duration := time.Since(start)

			fields := logging.RequestLogFields{
				Method:        r.Method,
				URL:           r.URL.String(),
				UserAgent:     r.UserAgent(),
				IP:            getClientIP(r),
				StatusCode:    wrapper.statusCode,
				ResponseSize:  wrapper.responseSize,
				Duration:      duration,
				RequestID:     requestID,
				CorrelationID: logging.GetCorrelationIDFromContext(ctx),
				UserID:        logging.GetUserIDFromContext(ctx).String(),
			}

			logger.LogRequest(fields)
		})
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture response details.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseSize += int64(size)
	return size, err
}

// getClientIP extracts the client IP address from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// Authentication middleware
type AuthConfig struct {
	TokenHeader      string
	TokenPrefix      string
	SkipPaths        []string
	RequiredScopes   []string
	UserContextKey   string
	ScopesContextKey string
}

// DefaultAuthConfig returns a default authentication configuration.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		TokenHeader:      "Authorization",
		TokenPrefix:      "Bearer ",
		SkipPaths:        []string{"/health", "/metrics"},
		RequiredScopes:   []string{},
		UserContextKey:   "user",
		ScopesContextKey: "scopes",
	}
}

// TokenValidator represents a function that validates authentication tokens.
type TokenValidator func(token string) (*AuthInfo, error)

// AuthInfo contains authentication information extracted from a token.
type AuthInfo struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expiresAt"`
	IssuedAt  time.Time `json:"issuedAt"`
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
}

// HasScope returns true if the user has the specified scope.
func (ai *AuthInfo) HasScope(scope string) bool {
	for _, s := range ai.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyScope returns true if the user has any of the specified scopes.
func (ai *AuthInfo) HasAnyScope(scopes []string) bool {
	for _, scope := range scopes {
		if ai.HasScope(scope) {
			return true
		}
	}
	return false
}

// IsExpired returns true if the token is expired.
func (ai *AuthInfo) IsExpired() bool {
	return time.Now().After(ai.ExpiresAt)
}

// Authentication creates an authentication middleware.
func Authentication(config AuthConfig, validator TokenValidator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			for _, skipPath := range config.SkipPaths {
				if r.URL.Path == skipPath {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token from header
			authHeader := r.Header.Get(config.TokenHeader)
			if authHeader == "" {
				writeErrorResponse(w, errors.Unauthorized("missing authorization header"))
				return
			}

			token := authHeader
			if config.TokenPrefix != "" && strings.HasPrefix(authHeader, config.TokenPrefix) {
				token = authHeader[len(config.TokenPrefix):]
			}

			// Validate token
			authInfo, err := validator(token)
			if err != nil {
				writeErrorResponse(w, errors.Unauthorized("invalid token"))
				return
			}

			// Check expiration
			if authInfo.IsExpired() {
				writeErrorResponse(w, errors.Unauthorized("token expired"))
				return
			}

			// Check required scopes
			if len(config.RequiredScopes) > 0 && !authInfo.HasAnyScope(config.RequiredScopes) {
				writeErrorResponse(w, errors.Forbidden("insufficient permissions"))
				return
			}

			// Add auth info to context
			ctx := context.WithValue(r.Context(), config.UserContextKey, authInfo)
			ctx = context.WithValue(ctx, config.ScopesContextKey, authInfo.Scopes)
			ctx = logging.WithUserIDContext(ctx, authInfo.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Rate Limiting middleware
type RateLimitConfig struct {
	RequestsPerWindow int
	Window            time.Duration
	KeyExtractor      func(*http.Request) string
}

// DefaultRateLimitConfig returns a default rate limiting configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		KeyExtractor:      defaultKeyExtractor,
	}
}

// defaultKeyExtractor extracts the client IP as the rate limiting key.
func defaultKeyExtractor(r *http.Request) string {
	return getClientIP(r)
}

// rateLimitEntry represents a rate limiting entry.
type rateLimitEntry struct {
	count     int
	resetTime time.Time
	mutex     sync.Mutex
}

// RateLimit creates a rate limiting middleware.
func RateLimit(config RateLimitConfig) Middleware {
	entries := make(map[string]*rateLimitEntry)
	mutex := sync.RWMutex{}

	// Cleanup expired entries periodically
	go func() {
		ticker := time.NewTicker(config.Window)
		defer ticker.Stop()

		for range ticker.C {
			mutex.Lock()
			now := time.Now()
			for key, entry := range entries {
				entry.mutex.Lock()
				if now.After(entry.resetTime) {
					delete(entries, key)
				}
				entry.mutex.Unlock()
			}
			mutex.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyExtractor(r)
			now := time.Now()

			mutex.RLock()
			entry, exists := entries[key]
			mutex.RUnlock()

			if !exists {
				entry = &rateLimitEntry{
					count:     0,
					resetTime: now.Add(config.Window),
				}
				mutex.Lock()
				entries[key] = entry
				mutex.Unlock()
			}

			entry.mutex.Lock()
			defer entry.mutex.Unlock()

			// Reset if window has passed
			if now.After(entry.resetTime) {
				entry.count = 0
				entry.resetTime = now.Add(config.Window)
			}

			// Check rate limit
			if entry.count >= config.RequestsPerWindow {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(entry.resetTime.Unix(), 10))
				writeErrorResponse(w, errors.RateLimited("rate limit exceeded"))
				return
			}

			// Increment counter
			entry.count++

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(config.RequestsPerWindow-entry.count))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(entry.resetTime.Unix(), 10))

			next.ServeHTTP(w, r)
		})
	}
}

// Error handling middleware
func ErrorHandling(logger *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error().
						Interface("panic", err).
						Str("url", r.URL.String()).
						Str("method", r.Method).
						Msg("Panic recovered")

					erpErr := errors.Internal("internal server error")
					writeErrorResponse(w, erpErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Timeout middleware
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					writeErrorResponse(w, errors.Timeout("request"))
				}
			}
		})
	}
}

// Security headers middleware
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next.ServeHTTP(w, r)
		})
	}
}

// Compression middleware (basic implementation)
func Compression() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// For a full implementation, you would use a compression library
			// This is a placeholder that sets the header but doesn't actually compress
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(w, r)
		})
	}
}

// Health check middleware
func HealthCheck(path string, checker func() map[string]interface{}) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == path {
				health := checker()
				w.Header().Set("Content-Type", "application/json")

				status := health["status"].(string)
				if status != "healthy" {
					w.WriteHeader(http.StatusServiceUnavailable)
				}

				json.NewEncoder(w).Encode(health)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Correlation ID middleware
func CorrelationID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate correlation ID
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Add to response header
			w.Header().Set("X-Correlation-ID", correlationID)

			// Add to context
			ctx := logging.WithCorrelationID(r.Context(), correlationID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Content type validation middleware
func ContentType(allowedTypes ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				contentType := r.Header.Get("Content-Type")
				if contentType == "" {
					writeErrorResponse(w, errors.Validation("missing Content-Type header"))
					return
				}

				// Remove charset and other parameters
				if semicolon := strings.Index(contentType, ";"); semicolon != -1 {
					contentType = contentType[:semicolon]
				}

				allowed := false
				for _, allowedType := range allowedTypes {
					if contentType == allowedType {
						allowed = true
						break
					}
				}

				if !allowed {
					writeErrorResponse(w, errors.Validation(fmt.Sprintf("unsupported Content-Type: %s", contentType)))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to write error responses
func writeErrorResponse(w http.ResponseWriter, err *errors.ERPError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.GetHTTPStatus())

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
		},
	}

	if err.UserMessage != "" {
		response["error"].(map[string]interface{})["userMessage"] = err.UserMessage
	}

	if err.CorrelationID != "" {
		response["correlationId"] = err.CorrelationID
	}

	json.NewEncoder(w).Encode(response)
}

// Context helper functions

// GetAuthInfoFromContext extracts authentication info from context.
func GetAuthInfoFromContext(ctx context.Context) (*AuthInfo, bool) {
	authInfo, ok := ctx.Value("user").(*AuthInfo)
	return authInfo, ok
}

// GetUserIDFromContext extracts user ID from context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if authInfo, ok := GetAuthInfoFromContext(ctx); ok {
		return authInfo.UserID, true
	}
	return uuid.UUID{}, false
}

// RequireAuth is a helper to ensure authentication is present.
func RequireAuth(ctx context.Context) (*AuthInfo, error) {
	authInfo, ok := GetAuthInfoFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("authentication required")
	}
	return authInfo, nil
}

// RequireScope is a helper to ensure a specific scope is present.
func RequireScope(ctx context.Context, scope string) error {
	authInfo, err := RequireAuth(ctx)
	if err != nil {
		return err
	}

	if !authInfo.HasScope(scope) {
		return errors.Forbidden(fmt.Sprintf("required scope: %s", scope))
	}

	return nil
}
