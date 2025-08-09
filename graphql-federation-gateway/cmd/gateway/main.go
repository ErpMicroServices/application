package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/graphql-federation-gateway/internal/config"
	"github.com/erpmicroservices/graphql-federation-gateway/internal/federation"
	"github.com/erpmicroservices/graphql-federation-gateway/pkg/gateway"
)

func main() {
	// Initialize logging
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	log.Info().Msg("Starting GraphQL Federation Gateway")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("port", cfg.Server.Port).
		Str("environment", cfg.Environment).
		Msg("Configuration loaded")

	// Initialize federation gateway
	fedGateway, err := federation.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize federation gateway")
	}

	// Create gateway server
	server, err := gateway.NewServer(cfg, fedGateway)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create gateway server")
	}

	// Start server
	go func() {
		log.Info().Str("address", fmt.Sprintf(":%s", cfg.Server.Port)).Msg("Starting GraphQL server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}