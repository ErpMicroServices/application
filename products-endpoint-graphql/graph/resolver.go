package graph

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/erpmicroservices/products-endpoint-graphql/internal/ai"
)

//go:generate go run github.com/99designs/gqlgen generate

// Resolver is the root GraphQL resolver
type Resolver struct {
	DB         *sqlx.DB
	AIServices *AIServices
	Logger     zerolog.Logger
}

// AIServices holds all AI service instances
type AIServices struct {
	Categorization *ai.CategorizationService
	Recommendation *ai.RecommendationEngine
}