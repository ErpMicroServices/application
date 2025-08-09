package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/auth-go/internal/config"
	"github.com/erpmicroservices/auth-go/pkg/cache"
	"github.com/erpmicroservices/auth-go/pkg/jwt"
	"github.com/erpmicroservices/auth-go/pkg/middleware"
	"github.com/erpmicroservices/auth-go/pkg/oauth2"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Configure logging
	setupLogging(cfg)

	log.Info().
		Str("service", cfg.Service.Name).
		Str("version", "1.0.0").
		Str("port", cfg.Service.Port).
		Msg("Starting auth-go example service")

	// Initialize components
	tokenCache := cache.NewInMemoryTokenCache(cfg.Cache.TTL, cfg.Cache.CleanupInterval)
	oauth2Client := oauth2.NewClient(&cfg.OAuth2, tokenCache)
	jwtParser := jwt.NewParser(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.SigningKey)
	validator := oauth2.NewValidator(oauth2Client, cfg.OAuth2.JWKSURL, cfg.JWT.Issuer, cfg.JWT.Audience)

	// Initialize middleware
	authConfig := middleware.DefaultAuthConfig()
	authConfig.SkipPaths = []string{"/health", "/metrics", "/oauth2/", "/graphql-playground"}
	authMiddleware := middleware.NewAuthMiddleware(oauth2Client, jwtParser, validator, authConfig)

	rbacMiddleware := middleware.NewRBACMiddleware(middleware.DefaultRBACConfig())

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", healthHandler)
	
	// Metrics endpoint
	mux.HandleFunc("/metrics", metricsHandler(tokenCache))
	
	// OAuth2 endpoints
	setupOAuth2Endpoints(mux, oauth2Client, &cfg.OAuth2)
	
	// Protected API endpoints
	setupAPIEndpoints(mux, authMiddleware, rbacMiddleware)
	
	// GraphQL endpoints (with authentication)
	setupGraphQLEndpoints(mux, oauth2Client, authMiddleware)

	// Setup CORS if needed
	handler := setupCORS(mux, &cfg.CORS)
	
	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Service.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", server.Addr).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}
	
	log.Info().Msg("Server shutdown complete")
}

// setupLogging configures the logging system
func setupLogging(cfg *config.Config) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	if cfg.Logging.Format == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	// Add service context
	log.Logger = log.With().
		Str("service", cfg.Service.Name).
		Logger()
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"service": "auth-go-example",
	})
}

// metricsHandler provides basic metrics about the service
func metricsHandler(tokenCache cache.TokenCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cache_size": tokenCache.Size(),
			"timestamp": time.Now().Unix(),
		})
	}
}

// setupOAuth2Endpoints sets up OAuth2 related endpoints
func setupOAuth2Endpoints(mux *http.ServeMux, client *oauth2.Client, cfg *config.OAuth2Config) {
	// OAuth2 authorization URL
	mux.HandleFunc("/oauth2/authorize", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			state = "example_state_" + fmt.Sprintf("%d", time.Now().Unix())
		}
		
		authURL := client.GetAuthorizationURL(state)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"authorization_url": authURL,
			"state": state,
		})
	})

	// OAuth2 callback handler
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code required", http.StatusBadRequest)
			return
		}

		token, err := client.ExchangeCodeForToken(r.Context(), code)
		if err != nil {
			log.Error().Err(err).Msg("Failed to exchange code for token")
			http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
	})

	// Token introspection endpoint
	mux.HandleFunc("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromRequest(r)
		if token == "" {
			http.Error(w, "Token required", http.StatusBadRequest)
			return
		}

		introspection, err := client.IntrospectToken(r.Context(), token)
		if err != nil {
			log.Error().Err(err).Msg("Failed to introspect token")
			http.Error(w, "Failed to introspect token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(introspection)
	})
}

// setupAPIEndpoints sets up protected API endpoints
func setupAPIEndpoints(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware) {
	// Protected endpoint that requires authentication
	mux.Handle("/api/profile", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCtx := middleware.MustGetAuthContext(r.Context())
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"subject":         authCtx.Subject,
			"name":           authCtx.Name,
			"email":          authCtx.Email,
			"roles":          authCtx.Roles,
			"authorities":    authCtx.Authorities,
			"organization_id": authCtx.OrganizationID,
			"department_id":   authCtx.DepartmentID,
		})
	})))

	// Admin endpoint that requires ADMIN role
	mux.Handle("/api/admin", authMiddleware.RequireAuth(
		rbacMiddleware.RequireRole("ADMIN")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Welcome, admin!",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})),
	))

	// Manager endpoint that requires MANAGER or ADMIN role
	mux.Handle("/api/manager", authMiddleware.RequireAuth(
		rbacMiddleware.RequireRole("MANAGER", "ADMIN")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := middleware.MustGetAuthContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Welcome, manager!",
				"user": authCtx.GetDisplayName(),
				"roles": authCtx.Roles,
			})
		})),
	))

	// Service endpoint that requires SERVICE authority
	mux.Handle("/api/service", authMiddleware.RequireAuth(
		rbacMiddleware.RequireAuthority("SERVICE")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := middleware.MustGetAuthContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Service-to-service communication",
				"client_id": authCtx.ClientID,
				"authorities": authCtx.Authorities,
			})
		})),
	))
}

// setupGraphQLEndpoints sets up GraphQL endpoints with authentication directives
func setupGraphQLEndpoints(mux *http.ServeMux, oauth2Client *oauth2.Client, authMiddleware *middleware.AuthMiddleware) {
	// Create a simple GraphQL server for demonstration
	// In a real application, you would use your generated GraphQL schema
	
	// GraphQL playground (development only)
	mux.Handle("/graphql-playground", playground.Handler("GraphQL Playground", "/graphql"))
	
	// GraphQL endpoint with authentication middleware
	mux.Handle("/graphql", authMiddleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This would be replaced with your actual GraphQL server
		// For demonstration, we'll return some sample data
		
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authCtx, authenticated := middleware.GetAuthContext(r.Context())
		
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"authenticated": authenticated,
			},
		}
		
		if authenticated {
			response["data"].(map[string]interface{})["user"] = map[string]interface{}{
				"id":      authCtx.Subject,
				"name":    authCtx.GetDisplayName(),
				"roles":   authCtx.Roles,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})))
}

// setupCORS configures CORS if needed
func setupCORS(handler http.Handler, corsConfig *config.CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if origin != "" && isOriginAllowed(origin, corsConfig.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		handler.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if an origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// extractTokenFromRequest extracts token from Authorization header or query parameter
func extractTokenFromRequest(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	// Try query parameter
	return r.URL.Query().Get("access_token")
}

// Example of how to use directives (this would be in your actual GraphQL resolvers)
func exampleDirectiveUsage() {
	// This is just for documentation - in real usage, you would configure your GraphQL server like this:
	
	/*
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{},
		Directives: generated.DirectiveRoot{
			Auth:          directives.Auth(oauth2Client),
			HasRole:       directives.HasRole(oauth2Client),
			HasAuthority:  directives.HasAuthority(oauth2Client),
			HasPermission: directives.HasPermission(oauth2Client),
		},
	}))
	*/
	
	log.Info().Msg("GraphQL directives would be configured in your schema")
}

// Example GraphQL schema (this would be in a .graphql file):
/*
directive @auth(requireEmailVerified: Boolean, requireServiceAccount: Boolean) on FIELD_DEFINITION
directive @hasRole(role: String, roles: [String!], requireAll: Boolean = false) on FIELD_DEFINITION  
directive @hasAuthority(authority: String, authorities: [String!], requireAll: Boolean = false) on FIELD_DEFINITION
directive @hasPermission(roles: [String!], authorities: [String!], requireAll: Boolean = false) on FIELD_DEFINITION

type Query {
  me: User @auth
  adminData: AdminData @hasRole(role: "ADMIN") 
  managerData: ManagerData @hasRole(roles: ["MANAGER", "ADMIN"])
  serviceData: ServiceData @hasAuthority(authority: "SERVICE")
  sensitiveData: SensitiveData @hasPermission(roles: ["ADMIN"], authorities: ["READ_SENSITIVE"])
}

type User {
  id: ID!
  name: String!
  email: String!
}

type AdminData {
  message: String!
}

type ManagerData {
  message: String!
  teamSize: Int!
}

type ServiceData {
  message: String!
  clientId: String!
}

type SensitiveData {
  message: String!
  data: String!
}
*/