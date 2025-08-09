package ai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/erpmicroservices/products-endpoint-graphql/pkg/models"
)

// RecommendationEngine provides AI-powered product recommendations
type RecommendationEngine struct {
	logger         zerolog.Logger
	modelVersion   string
	defaultLimit   int
	minConfidence  float64
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(logger zerolog.Logger, modelVersion string) *RecommendationEngine {
	return &RecommendationEngine{
		logger:         logger.With().Str("service", "ai-recommendations").Logger(),
		modelVersion:   modelVersion,
		defaultLimit:   10,
		minConfidence:  0.3,
	}
}

// CollaborativeFiltering generates recommendations based on user behavior patterns
func (r *RecommendationEngine) CollaborativeFiltering(ctx context.Context, customerID uuid.UUID, limit int) ([]models.RecommendationScore, error) {
	r.logger.Info().
		Str("customer_id", customerID.String()).
		Int("limit", limit).
		Msg("Generating collaborative filtering recommendations")

	// Simulate collaborative filtering algorithm
	recommendations := []models.RecommendationScore{
		{
			ProductID:       uuid.New(),
			Score:          0.89,
			ReasonCode:     "COLLABORATIVE_SIMILAR_USERS",
			Explanation:    "Customers with similar purchase history also bought this item",
			ConfidenceLevel: 0.85,
			GeneratedAt:    time.Now(),
		},
		{
			ProductID:       uuid.New(),
			Score:          0.82,
			ReasonCode:     "COLLABORATIVE_FREQUENT_TOGETHER",
			Explanation:    "Often purchased together with items in your cart",
			ConfidenceLevel: 0.78,
			GeneratedAt:    time.Now(),
		},
		{
			ProductID:       uuid.New(),
			Score:          0.75,
			ReasonCode:     "COLLABORATIVE_CATEGORY_AFFINITY",
			Explanation:    "Based on your category preferences",
			ConfidenceLevel: 0.71,
			GeneratedAt:    time.Now(),
		},
	}

	// Sort by score and limit results
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if limit > 0 && limit < len(recommendations) {
		recommendations = recommendations[:limit]
	}

	r.logger.Info().
		Str("customer_id", customerID.String()).
		Int("recommendations_count", len(recommendations)).
		Msg("Collaborative filtering completed")

	return recommendations, nil
}

// ContentBasedFiltering generates recommendations based on product features
func (r *RecommendationEngine) ContentBasedFiltering(ctx context.Context, productID uuid.UUID, limit int) ([]models.RecommendationScore, error) {
	r.logger.Info().
		Str("product_id", productID.String()).
		Int("limit", limit).
		Msg("Generating content-based recommendations")

	// Simulate content-based filtering
	recommendations := []models.RecommendationScore{
		{
			ProductID:       uuid.New(),
			Score:          0.92,
			ReasonCode:     "CONTENT_SIMILAR_FEATURES",
			Explanation:    "Similar features and specifications",
			ConfidenceLevel: 0.88,
			GeneratedAt:    time.Now(),
		},
		{
			ProductID:       uuid.New(),
			Score:          0.86,
			ReasonCode:     "CONTENT_SAME_CATEGORY",
			Explanation:    "Same category with higher ratings",
			ConfidenceLevel: 0.82,
			GeneratedAt:    time.Now(),
		},
		{
			ProductID:       uuid.New(),
			Score:          0.79,
			ReasonCode:     "CONTENT_PRICE_RANGE",
			Explanation:    "Similar price range and features",
			ConfidenceLevel: 0.74,
			GeneratedAt:    time.Now(),
		},
	}

	// Apply limit
	if limit > 0 && limit < len(recommendations) {
		recommendations = recommendations[:limit]
	}

	r.logger.Info().
		Str("product_id", productID.String()).
		Int("recommendations_count", len(recommendations)).
		Msg("Content-based filtering completed")

	return recommendations, nil
}

// HybridRecommendations combines multiple recommendation approaches
func (r *RecommendationEngine) HybridRecommendations(ctx context.Context, customerID *uuid.UUID, productID *uuid.UUID, limit int) ([]models.RecommendationScore, error) {
	r.logger.Info().
		Interface("customer_id", customerID).
		Interface("product_id", productID).
		Int("limit", limit).
		Msg("Generating hybrid recommendations")

	var allRecommendations []models.RecommendationScore

	// Get collaborative recommendations if customer ID provided
	if customerID != nil {
		collabRecs, err := r.CollaborativeFiltering(ctx, *customerID, limit*2)
		if err != nil {
			r.logger.Warn().Err(err).Msg("Failed to get collaborative recommendations")
		} else {
			allRecommendations = append(allRecommendations, collabRecs...)
		}
	}

	// Get content-based recommendations if product ID provided
	if productID != nil {
		contentRecs, err := r.ContentBasedFiltering(ctx, *productID, limit*2)
		if err != nil {
			r.logger.Warn().Err(err).Msg("Failed to get content-based recommendations")
		} else {
			allRecommendations = append(allRecommendations, contentRecs...)
		}
	}

	// Add trending/popular recommendations
	trendingRecs := r.getTrendingRecommendations(limit)
	allRecommendations = append(allRecommendations, trendingRecs...)

	// Deduplicate and blend scores
	uniqueRecs := r.deduplicateAndBlend(allRecommendations)

	// Sort by final score
	sort.Slice(uniqueRecs, func(i, j int) bool {
		return uniqueRecs[i].Score > uniqueRecs[j].Score
	})

	// Apply limit
	if limit > 0 && limit < len(uniqueRecs) {
		uniqueRecs = uniqueRecs[:limit]
	}

	r.logger.Info().
		Interface("customer_id", customerID).
		Interface("product_id", productID).
		Int("final_recommendations", len(uniqueRecs)).
		Msg("Hybrid recommendations completed")

	return uniqueRecs, nil
}

// PersonalizedRecommendations generates recommendations based on customer profile
func (r *RecommendationEngine) PersonalizedRecommendations(ctx context.Context, customerID uuid.UUID, preferences CustomerPreferences) ([]models.RecommendationScore, error) {
	r.logger.Info().
		Str("customer_id", customerID.String()).
		Interface("preferences", preferences).
		Msg("Generating personalized recommendations")

	// Simulate personalized recommendation generation
	recommendations := []models.RecommendationScore{}

	// Factor in eco-friendly preferences
	if preferences.EcoFriendly {
		recommendations = append(recommendations, models.RecommendationScore{
			ProductID:       uuid.New(),
			Score:          0.88,
			ReasonCode:     "PERSONALIZED_ECO_PREFERENCE",
			Explanation:    "Matches your eco-friendly preferences",
			ConfidenceLevel: 0.84,
			GeneratedAt:    time.Now(),
		})
	}

	// Factor in budget constraints
	if preferences.MaxPrice > 0 {
		recommendations = append(recommendations, models.RecommendationScore{
			ProductID:       uuid.New(),
			Score:          0.81,
			ReasonCode:     "PERSONALIZED_BUDGET_MATCH",
			Explanation:    fmt.Sprintf("Within your budget of $%.2f", preferences.MaxPrice),
			ConfidenceLevel: 0.79,
			GeneratedAt:    time.Now(),
		})
	}

	// Factor in brand preferences
	for _, brand := range preferences.PreferredBrands {
		recommendations = append(recommendations, models.RecommendationScore{
			ProductID:       uuid.New(),
			Score:          0.76,
			ReasonCode:     "PERSONALIZED_BRAND_PREFERENCE",
			Explanation:    fmt.Sprintf("From your preferred brand: %s", brand),
			ConfidenceLevel: 0.73,
			GeneratedAt:    time.Now(),
		})
	}

	r.logger.Info().
		Str("customer_id", customerID.String()).
		Int("personalized_recommendations", len(recommendations)).
		Msg("Personalized recommendations completed")

	return recommendations, nil
}

// SeasonalRecommendations adjusts recommendations based on seasonal patterns
func (r *RecommendationEngine) SeasonalRecommendations(ctx context.Context, baseRecommendations []models.RecommendationScore, season string) []models.RecommendationScore {
	r.logger.Info().
		Str("season", season).
		Int("base_count", len(baseRecommendations)).
		Msg("Applying seasonal adjustments")

	// Seasonal multipliers
	seasonalMultipliers := map[string]float64{
		"winter":  1.2,
		"spring":  1.1,
		"summer":  1.0,
		"fall":    1.15,
		"holiday": 1.3,
	}

	multiplier, exists := seasonalMultipliers[season]
	if !exists {
		multiplier = 1.0
	}

	// Apply seasonal adjustments
	adjustedRecommendations := make([]models.RecommendationScore, len(baseRecommendations))
	for i, rec := range baseRecommendations {
		adjustedRecommendations[i] = rec
		adjustedRecommendations[i].Score = math.Min(1.0, rec.Score*multiplier)
		
		if multiplier > 1.0 {
			adjustedRecommendations[i].Explanation = fmt.Sprintf("%s (Seasonal boost: %s)", 
				rec.Explanation, season)
			adjustedRecommendations[i].ReasonCode = fmt.Sprintf("%s_SEASONAL_%s", 
				rec.ReasonCode, strings.ToUpper(season))
		}
	}

	r.logger.Info().
		Str("season", season).
		Float64("multiplier", multiplier).
		Int("adjusted_count", len(adjustedRecommendations)).
		Msg("Seasonal adjustments applied")

	return adjustedRecommendations
}

// RealTimeUpdate updates recommendations based on current session activity
func (r *RecommendationEngine) RealTimeUpdate(ctx context.Context, sessionData SessionData) ([]models.RecommendationScore, error) {
	r.logger.Info().
		Str("session_id", sessionData.SessionID).
		Int("viewed_products", len(sessionData.ViewedProducts)).
		Msg("Updating recommendations for real-time session")

	// Analyze current session patterns
	categoryInterest := r.analyzeCategoryInterest(sessionData.ViewedProducts)
	recommendations := []models.RecommendationScore{}

	// Generate recommendations based on session data
	for category, interest := range categoryInterest {
		if interest > 0.5 { // Threshold for interest
			rec := models.RecommendationScore{
				ProductID:       uuid.New(), // Would be actual product from category
				Score:          interest * 0.9, // Slight discount for recency
				ReasonCode:     "REALTIME_SESSION_INTEREST",
				Explanation:    fmt.Sprintf("Based on your current interest in %s", category),
				ConfidenceLevel: interest * 0.85,
				GeneratedAt:    time.Now(),
			}
			recommendations = append(recommendations, rec)
		}
	}

	r.logger.Info().
		Str("session_id", sessionData.SessionID).
		Int("realtime_recommendations", len(recommendations)).
		Msg("Real-time recommendations updated")

	return recommendations, nil
}

// Helper methods

func (r *RecommendationEngine) getTrendingRecommendations(limit int) []models.RecommendationScore {
	// Simulate trending products
	trending := []models.RecommendationScore{
		{
			ProductID:       uuid.New(),
			Score:          0.72,
			ReasonCode:     "TRENDING_POPULAR",
			Explanation:    "Trending now - popular choice",
			ConfidenceLevel: 0.68,
			GeneratedAt:    time.Now(),
		},
		{
			ProductID:       uuid.New(),
			Score:          0.69,
			ReasonCode:     "TRENDING_NEW_ARRIVAL",
			Explanation:    "New arrival with high ratings",
			ConfidenceLevel: 0.65,
			GeneratedAt:    time.Now(),
		},
	}

	if limit > 0 && limit < len(trending) {
		return trending[:limit]
	}
	return trending
}

func (r *RecommendationEngine) deduplicateAndBlend(recommendations []models.RecommendationScore) []models.RecommendationScore {
	// Use map to deduplicate by product ID
	productScores := make(map[uuid.UUID]models.RecommendationScore)
	
	for _, rec := range recommendations {
		if existing, exists := productScores[rec.ProductID]; exists {
			// Blend scores using weighted average
			newScore := (existing.Score + rec.Score) / 2
			newConfidence := (existing.ConfidenceLevel + rec.ConfidenceLevel) / 2
			
			productScores[rec.ProductID] = models.RecommendationScore{
				ProductID:       rec.ProductID,
				Score:          newScore,
				ReasonCode:     "HYBRID_BLENDED",
				Explanation:    fmt.Sprintf("Combined: %s + %s", existing.Explanation, rec.Explanation),
				ConfidenceLevel: newConfidence,
				GeneratedAt:    time.Now(),
			}
		} else {
			productScores[rec.ProductID] = rec
		}
	}
	
	// Convert back to slice
	result := make([]models.RecommendationScore, 0, len(productScores))
	for _, rec := range productScores {
		result = append(result, rec)
	}
	
	return result
}

func (r *RecommendationEngine) analyzeCategoryInterest(viewedProducts []ViewedProduct) map[string]float64 {
	categoryViews := make(map[string]int)
	totalViews := len(viewedProducts)
	
	// Count views per category
	for _, product := range viewedProducts {
		categoryViews[product.Category]++
	}
	
	// Calculate interest scores
	categoryInterest := make(map[string]float64)
	for category, views := range categoryViews {
		interest := float64(views) / float64(totalViews)
		
		// Apply recency decay (more recent views are more important)
		// This is a simplified version - in reality you'd consider timestamps
		categoryInterest[category] = interest
	}
	
	return categoryInterest
}

// Supporting types

type CustomerPreferences struct {
	EcoFriendly      bool     `json:"ecoFriendly"`
	MaxPrice        float64   `json:"maxPrice"`
	PreferredBrands []string  `json:"preferredBrands"`
	Categories      []string  `json:"categories"`
	Attributes      []string  `json:"attributes"`
}

type SessionData struct {
	SessionID      string          `json:"sessionId"`
	CustomerID     *uuid.UUID      `json:"customerId,omitempty"`
	ViewedProducts []ViewedProduct `json:"viewedProducts"`
	CartItems      []uuid.UUID     `json:"cartItems"`
	SearchQueries  []string        `json:"searchQueries"`
	StartTime      time.Time       `json:"startTime"`
}

type ViewedProduct struct {
	ProductID   uuid.UUID `json:"productId"`
	Category    string    `json:"category"`
	ViewedAt    time.Time `json:"viewedAt"`
	DurationSec int       `json:"durationSec"`
}