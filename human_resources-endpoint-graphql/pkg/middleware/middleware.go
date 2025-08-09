package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/erpmicroservices/human_resources-endpoint-graphql/internal/config"
)

// NewRequestLogger creates a request logging middleware
func NewRequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  &zerologWrapper{logger},
		NoColor: true,
	})
}

// NewAuthMiddleware creates an authentication middleware
func NewAuthMiddleware(cfg config.AuthConfig, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// For now, allow requests without auth header
				// In production, you might want to reject these
				next.ServeHTTP(w, r)
				return
			}

			// Parse and validate token
			// This is a placeholder - implement your JWT validation logic here
			userID, roles, err := validateToken(authHeader, cfg)
			if err != nil {
				logger.Error().Err(err).Msg("Token validation failed")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user information to context
			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "user_roles", roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken validates JWT token and extracts user information
func validateToken(authHeader string, cfg config.AuthConfig) (string, []string, error) {
	// This is a placeholder implementation
	// In a real application, you would:
	// 1. Parse the Bearer token from authHeader
	// 2. Validate JWT signature using cfg.JWTSecret
	// 3. Extract user ID and roles from token claims
	// 4. Return user information or error

	// For now, return dummy data for development
	return "user123", []string{"HR_USER"}, nil
}

// zerologWrapper wraps zerolog.Logger to implement middleware.LoggerInterface
type zerologWrapper struct {
	logger zerolog.Logger
}

func (z *zerologWrapper) Print(v ...interface{}) {
	z.logger.Info().Msg(fmt.Sprint(v...))
}