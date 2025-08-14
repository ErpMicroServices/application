package middleware

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/erpmicroservices/products-endpoint-graphql/internal/config"
)

// NewRequestLogger creates a new request logging middleware
func NewRequestLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("user_agent", r.UserAgent()).
				Str("remote_addr", r.RemoteAddr).
				Msg("Request received")
			next.ServeHTTP(w, r)
		})
	}
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(cfg config.AuthConfig, logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement authentication logic
			// For now, just pass through if auth is disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Add authentication logic here
			logger.Debug().Msg("Authentication middleware - implementation pending")
			next.ServeHTTP(w, r)
		})
	}
}