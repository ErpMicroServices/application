package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
	"github.com/erpmicroservices/graphql-federation-gateway/internal/federation"
)

// Server represents the GraphQL federation gateway server
type Server struct {
	config      *config.Config
	httpServer  *http.Server
	federation  *federation.Gateway
}

// NewServer creates a new federation gateway server
func NewServer(cfg *config.Config, fedGateway *federation.Gateway) (*Server, error) {
	s := &Server{
		config:     cfg,
		federation: fedGateway,
	}

	// Setup HTTP router
	router := s.setupRouter()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return s, nil
}

// setupRouter configures the HTTP router with all routes and middleware
func (s *Server) setupRouter() chi.Router {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(s.config.Server.WriteTimeout))

	// CORS middleware
	if s.config.Server.EnableCORS {
		corsHandler := cors.New(cors.Options{
			AllowedOrigins: s.config.Server.CORSOrigins,
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
			ExposedHeaders: []string{"Link"},
			AllowCredentials: true,
			MaxAge: 300,
		})
		r.Use(corsHandler.Handler)
	}

	// Health check endpoint
	r.Get(s.config.Monitoring.HealthCheckPath, s.handleHealth)

	// GraphQL endpoint
	r.Post("/graphql", s.handleGraphQL)

	// GraphQL playground (development only)
	if s.config.Server.EnablePlayground && s.config.Environment != "production" {
		r.Get("/", s.handlePlayground)
	}

	// Metrics endpoint (on different port in production)
	if s.config.Monitoring.EnableMetrics {
		r.Get("/metrics", s.handleMetrics)
	}

	return r
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	log.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting HTTP server")

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down HTTP server")

	return s.httpServer.Shutdown(ctx)
}

// HTTP Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement comprehensive health check
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GraphQL request handling
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"errors":[{"message":"GraphQL handler not yet implemented"}]}`))
}

func (s *Server) handlePlayground(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GraphQL playground
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
</head>
<body>
    <h1>GraphQL Federation Gateway</h1>
    <p>Playground will be implemented with GraphQL handler</p>
    <p>GraphQL endpoint: <a href="/graphql">/graphql</a></p>
</body>
</html>
	`))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement Prometheus metrics handler
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("# Metrics handler not yet implemented\n"))
}