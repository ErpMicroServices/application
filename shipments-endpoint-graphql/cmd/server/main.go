package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/erpmicroservices/shipments-endpoint-graphql/graph"
	"github.com/erpmicroservices/shipments-endpoint-graphql/graph/generated"
	"github.com/erpmicroservices/shipments-endpoint-graphql/internal/config"
	"github.com/erpmicroservices/shipments-endpoint-graphql/pkg/directives"
	"github.com/erpmicroservices/shipments-endpoint-graphql/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg.Logging)
	logger := log.With().Str("service", "shipments-graphql-api").Logger()

	logger.Info().
		Str("version", getVersion()).
		Str("environment", cfg.Server.Environment).
		Int("port", cfg.Server.Port).
		Msg("Starting Human Resources GraphQL API")

	// Setup database connection
	db, err := setupDatabase(cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to setup database")
	}
	defer db.Close()

	// Setup GraphQL server
	graphqlServer := setupGraphQLServer(cfg, db, logger)

	// Setup HTTP server
	httpServer := setupHTTPServer(cfg, graphqlServer, logger)

	// Start server with graceful shutdown
	if err := runServer(httpServer, logger); err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}

	logger.Info().Msg("Server shutdown complete")
}

func setupLogging(cfg config.LoggingConfig) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	if cfg.Format == "console" || cfg.EnableConsole {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		log.Logger = log.Output(os.Stdout)
	}

	// Add common fields
	log.Logger = log.Logger.With().
		Timestamp().
		Str("service", "shipments-graphql-api").
		Logger()
}

func setupDatabase(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("Database connection established")

	return db, nil
}

func setupGraphQLServer(cfg *config.Config, db *sqlx.DB, logger zerolog.Logger) *handler.Server {
	// Create resolver with dependencies
	resolver := &graph.Resolver{
		DB:     db,
		Logger: logger.With().Str("component", "graphql-resolver").Logger(),
	}

	// Create executable schema with directives
	executableSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
		Directives: generated.DirectiveRoot{
			Auth:       directives.Auth,
			HasRole:    directives.HasRole,
			ReadOnly:   directives.ReadOnly,
			Complexity: directives.Complexity,
		},
	})

	// Create GraphQL server
	srv := handler.New(executableSchema)

	// Configure transports
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{})

	// Configure extensions
	if cfg.GraphQL.EnableDataLoader {
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New(100),
		})
	}

	// Add complexity limits
	srv.Use(extension.FixedComplexityLimit(cfg.GraphQL.ComplexityLimit))

	// Add introspection
	if cfg.GraphQL.Introspection {
		srv.Use(extension.Introspection{})
	}

	// Add Apollo tracing for performance monitoring
	if cfg.GraphQL.EnableTracing {
		srv.Use(apollotracing.Tracer{})
	}

	// Add query cache
	if cfg.GraphQL.CacheEnabled {
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New(1000),
		})
	}

	logger.Info().
		Bool("playground", cfg.GraphQL.Playground).
		Bool("introspection", cfg.GraphQL.Introspection).
		Int("complexity_limit", cfg.GraphQL.ComplexityLimit).
		Int("depth_limit", cfg.GraphQL.DepthLimit).
		Msg("GraphQL server configured")

	return srv
}

func setupHTTPServer(cfg *config.Config, graphqlServer *handler.Server, logger zerolog.Logger) *http.Server {
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(cfg.Server.WriteTimeout))

	// Add custom middleware
	router.Use(middleware.NewRequestLogger(logger))
	
	if cfg.Auth.Enabled {
		router.Use(middleware.NewAuthMiddleware(cfg.Auth, logger))
	}

	// CORS configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure appropriately for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"shipments-graphql-api","timestamp":"%s"}`, 
			time.Now().UTC().Format(time.RFC3339))
	})

	// Ready check endpoint
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity and other dependencies
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"shipments-graphql-api","timestamp":"%s"}`, 
			time.Now().UTC().Format(time.RFC3339))
	})

	// Metrics endpoint (basic)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "# Human Resources GraphQL API Metrics\n")
		fmt.Fprintf(w, "hr_api_uptime{service=\"shipments-graphql-api\"} %d\n", 
			int64(time.Since(startTime).Seconds()))
	})

	// GraphQL endpoints
	router.Handle("/graphql", graphqlServer)
	
	if cfg.GraphQL.Playground {
		router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
		logger.Info().Msg("GraphQL Playground available at /playground")
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return server
}

func runServer(server *http.Server, logger zerolog.Logger) error {
	// Create error group for concurrent operations
	g, ctx := errgroup.WithContext(context.Background())

	// Start HTTP server
	g.Go(func() error {
		logger.Info().
			Str("addr", server.Addr).
			Msg("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
		return nil
	})

	// Handle shutdown signals
	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		case <-ctx.Done():
			return ctx.Err()
		}

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Info().Msg("Shutting down HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("HTTP server shutdown failed")
			return err
		}

		logger.Info().Msg("HTTP server shutdown complete")
		return nil
	})

	return g.Wait()
}

func getVersion() string {
	// In a real application, this would be set during build
	return "0.0.1-SNAPSHOT"
}

var startTime = time.Now()