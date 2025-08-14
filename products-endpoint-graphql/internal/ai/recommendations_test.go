package ai

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendationEngine(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	engine := NewRecommendationEngine(logger, "test-model-v1.0")

	t.Run("CollaborativeFiltering", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()
		limit := 5

		recommendations, err := engine.CollaborativeFiltering(ctx, customerID, limit)

		require.NoError(t, err)
		assert.Len(t, recommendations, 3) // Based on mock data
		
		// Verify recommendations are sorted by score (descending)
		for i := 1; i < len(recommendations); i++ {
			assert.True(t, recommendations[i-1].Score >= recommendations[i].Score,
				"Recommendations should be sorted by score")
		}
		
		// Verify recommendation quality
		for _, rec := range recommendations {
			assert.True(t, rec.Score > 0 && rec.Score <= 1.0, "Score should be between 0 and 1")
			assert.True(t, rec.ConfidenceLevel > 0 && rec.ConfidenceLevel <= 1.0, "Confidence should be between 0 and 1")
			assert.NotEmpty(t, rec.ReasonCode)
			assert.NotEmpty(t, rec.Explanation)
			assert.NotNil(t, rec.ProductID)
		}
		
		// Verify collaborative filtering specific reason codes
		foundCollaborativeReason := false
		for _, rec := range recommendations {
			if rec.ReasonCode == "COLLABORATIVE_SIMILAR_USERS" ||
			   rec.ReasonCode == "COLLABORATIVE_FREQUENT_TOGETHER" ||
			   rec.ReasonCode == "COLLABORATIVE_CATEGORY_AFFINITY" {
				foundCollaborativeReason = true
				break
			}
		}
		assert.True(t, foundCollaborativeReason, "Should have collaborative filtering reason codes")
	})

	t.Run("ContentBasedFiltering", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		limit := 3

		recommendations, err := engine.ContentBasedFiltering(ctx, productID, limit)

		require.NoError(t, err)
		assert.Len(t, recommendations, 3)
		
		// Verify content-based specific reason codes
		foundContentReason := false
		for _, rec := range recommendations {
			if rec.ReasonCode == "CONTENT_SIMILAR_FEATURES" ||
			   rec.ReasonCode == "CONTENT_SAME_CATEGORY" ||
			   rec.ReasonCode == "CONTENT_PRICE_RANGE" {
				foundContentReason = true
			}
			
			// Verify basic recommendation structure
			assert.True(t, rec.Score > 0.7, "Content-based recommendations should have high scores")
			assert.NotEmpty(t, rec.Explanation)
		}
		assert.True(t, foundContentReason, "Should have content-based reason codes")
	})

	t.Run("HybridRecommendations", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()
		productID := uuid.New()
		limit := 5

		recommendations, err := engine.HybridRecommendations(ctx, &customerID, &productID, limit)

		require.NoError(t, err)
		assert.True(t, len(recommendations) <= limit, "Should respect limit")
		assert.True(t, len(recommendations) > 0, "Should return recommendations")
		
		// Should include diverse recommendation types
		reasonCodes := make(map[string]bool)
		for _, rec := range recommendations {
			reasonCodes[rec.ReasonCode] = true
			assert.True(t, rec.Score > 0, "All recommendations should have positive scores")
		}
		
		// Should have blend of different recommendation types
		assert.True(t, len(reasonCodes) > 1, "Hybrid should combine multiple recommendation types")
	})

	t.Run("PersonalizedRecommendations", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()
		preferences := CustomerPreferences{
			EcoFriendly:      true,
			MaxPrice:        500.0,
			PreferredBrands: []string{"Apple", "Samsung"},
			Categories:      []string{"Electronics", "Books"},
		}

		recommendations, err := engine.PersonalizedRecommendations(ctx, customerID, preferences)

		require.NoError(t, err)
		assert.NotEmpty(t, recommendations)
		
		// Should include personalization-specific reasons
		foundPersonalizationReasons := make(map[string]bool)
		for _, rec := range recommendations {
			switch rec.ReasonCode {
			case "PERSONALIZED_ECO_PREFERENCE":
				foundPersonalizationReasons["eco"] = true
			case "PERSONALIZED_BUDGET_MATCH":
				foundPersonalizationReasons["budget"] = true
			case "PERSONALIZED_BRAND_PREFERENCE":
				foundPersonalizationReasons["brand"] = true
			}
		}
		
		// Should reflect eco-friendly preference
		assert.True(t, foundPersonalizationReasons["eco"], "Should include eco-friendly recommendations")
		assert.True(t, foundPersonalizationReasons["budget"], "Should include budget-conscious recommendations")
		assert.True(t, foundPersonalizationReasons["brand"], "Should include brand preferences")
	})

	t.Run("SeasonalRecommendations", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()
		
		// Get base recommendations
		baseRecs, err := engine.CollaborativeFiltering(ctx, customerID, 3)
		require.NoError(t, err)
		
		// Apply seasonal adjustments
		holidayRecs := engine.SeasonalRecommendations(ctx, baseRecs, "holiday")
		
		// Holiday season should boost scores
		for i, rec := range holidayRecs {
			assert.True(t, rec.Score >= baseRecs[i].Score, "Holiday season should boost or maintain scores")
			
			if rec.Score > baseRecs[i].Score {
				assert.Contains(t, rec.Explanation, "Seasonal boost", "Boosted recommendations should mention seasonal adjustment")
			}
		}
		
		// Test with unknown season (should not change scores)
		unknownRecs := engine.SeasonalRecommendations(ctx, baseRecs, "unknown")
		for i, rec := range unknownRecs {
			assert.Equal(t, baseRecs[i].Score, rec.Score, "Unknown season should not change scores")
		}
	})

	t.Run("RealTimeUpdate", func(t *testing.T) {
		ctx := context.Background()
		sessionData := SessionData{
			SessionID:  "test-session-123",
			CustomerID: &uuid.UUID{},
			ViewedProducts: []ViewedProduct{
				{
					ProductID:   uuid.New(),
					Category:    "Electronics/Laptops",
					ViewedAt:    time.Now().Add(-5 * time.Minute),
					DurationSec: 120,
				},
				{
					ProductID:   uuid.New(),
					Category:    "Electronics/Laptops",
					ViewedAt:    time.Now().Add(-3 * time.Minute),
					DurationSec: 90,
				},
				{
					ProductID:   uuid.New(),
					Category:    "Electronics/Accessories",
					ViewedAt:    time.Now().Add(-1 * time.Minute),
					DurationSec: 60,
				},
			},
			CartItems:     []uuid.UUID{uuid.New()},
			SearchQueries: []string{"laptop", "macbook"},
			StartTime:     time.Now().Add(-30 * time.Minute),
		}

		recommendations, err := engine.RealTimeUpdate(ctx, sessionData)

		require.NoError(t, err)
		
		// Should generate recommendations based on session interest
		if len(recommendations) > 0 {
			// Verify real-time recommendation characteristics
			for _, rec := range recommendations {
				assert.Equal(t, "REALTIME_SESSION_INTEREST", rec.ReasonCode)
				assert.Contains(t, rec.Explanation, "current interest")
				assert.True(t, rec.Score > 0.5, "Real-time recommendations should have decent scores")
			}
		}
	})
}

func TestRecommendationEnginePerformance(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	engine := NewRecommendationEngine(logger, "performance-test-v1.0")

	t.Run("RecommendationLatency", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()

		// Test collaborative filtering performance
		start := time.Now()
		_, err := engine.CollaborativeFiltering(ctx, customerID, 10)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.True(t, duration < 200*time.Millisecond, 
			"Collaborative filtering should complete within 200ms, took %v", duration)
		
		// Test content-based filtering performance
		productID := uuid.New()
		start = time.Now()
		_, err = engine.ContentBasedFiltering(ctx, productID, 10)
		duration = time.Since(start)

		require.NoError(t, err)
		assert.True(t, duration < 200*time.Millisecond,
			"Content-based filtering should complete within 200ms, took %v", duration)
	})

	t.Run("HybridRecommendationPerformance", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()
		productID := uuid.New()

		start := time.Now()
		recommendations, err := engine.HybridRecommendations(ctx, &customerID, &productID, 20)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotEmpty(t, recommendations)
		assert.True(t, duration < 500*time.Millisecond,
			"Hybrid recommendations should complete within 500ms, took %v", duration)
	})
}

func TestRecommendationEngineEdgeCases(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	engine := NewRecommendationEngine(logger, "edge-case-test-v1.0")

	t.Run("ZeroLimit", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()

		recommendations, err := engine.CollaborativeFiltering(ctx, customerID, 0)

		require.NoError(t, err)
		// Should return available recommendations even with 0 limit
		assert.NotEmpty(t, recommendations)
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		ctx := context.Background()
		customerID := uuid.New()

		recommendations, err := engine.CollaborativeFiltering(ctx, customerID, -5)

		require.NoError(t, err)
		// Should handle negative limit gracefully
		assert.NotEmpty(t, recommendations)
	})

	t.Run("HybridWithNilParameters", func(t *testing.T) {
		ctx := context.Background()

		// Test with nil customer and product IDs
		recommendations, err := engine.HybridRecommendations(ctx, nil, nil, 5)

		require.NoError(t, err)
		// Should still return trending recommendations
		assert.NotEmpty(t, recommendations)
		
		// Verify trending recommendations are returned
		foundTrending := false
		for _, rec := range recommendations {
			if rec.ReasonCode == "TRENDING_POPULAR" || rec.ReasonCode == "TRENDING_NEW_ARRIVAL" {
				foundTrending = true
				break
			}
		}
		assert.True(t, foundTrending, "Should return trending recommendations when no customer/product context")
	})

	t.Run("EmptySessionData", func(t *testing.T) {
		ctx := context.Background()
		sessionData := SessionData{
			SessionID:      "empty-session",
			ViewedProducts: []ViewedProduct{},
			CartItems:      []uuid.UUID{},
			SearchQueries:  []string{},
			StartTime:      time.Now(),
		}

		recommendations, err := engine.RealTimeUpdate(ctx, sessionData)

		require.NoError(t, err)
		// Empty session should not generate recommendations
		assert.Empty(t, recommendations)
	})
}

// Test helper methods
func TestRecommendationEngineHelpers(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	engine := NewRecommendationEngine(logger, "helper-test-v1.0")

	t.Run("DeduplicateAndBlend", func(t *testing.T) {
		// Create recommendations with duplicate product IDs
		recommendations := []models.RecommendationScore{
			{
				ProductID:       uuid.New(),
				Score:          0.8,
				ReasonCode:     "TEST_REASON_1",
				Explanation:    "First explanation",
				ConfidenceLevel: 0.75,
			},
			{
				ProductID:       uuid.New(), // Different product
				Score:          0.7,
				ReasonCode:     "TEST_REASON_2", 
				Explanation:    "Second explanation",
				ConfidenceLevel: 0.65,
			},
		}
		
		// Add duplicate of first product with different score
		duplicate := recommendations[0]
		duplicate.Score = 0.9
		duplicate.ReasonCode = "TEST_REASON_3"
		duplicate.Explanation = "Third explanation"
		duplicate.ConfidenceLevel = 0.85
		recommendations = append(recommendations, duplicate)

		result := engine.deduplicateAndBlend(recommendations)

		// Should have 2 unique products (duplicate was blended)
		assert.Len(t, result, 2)
		
		// Find the blended recommendation
		var blendedRec *models.RecommendationScore
		for i := range result {
			if result[i].ReasonCode == "HYBRID_BLENDED" {
				blendedRec = &result[i]
				break
			}
		}
		
		require.NotNil(t, blendedRec, "Should have a blended recommendation")
		assert.Contains(t, blendedRec.Explanation, "Combined:")
		
		// Blended score should be average of 0.8 and 0.9 = 0.85
		assert.InDelta(t, 0.85, blendedRec.Score, 0.01)
	})

	t.Run("AnalyzeCategoryInterest", func(t *testing.T) {
		viewedProducts := []ViewedProduct{
			{Category: "Electronics", ViewedAt: time.Now()},
			{Category: "Electronics", ViewedAt: time.Now()},
			{Category: "Books", ViewedAt: time.Now()},
		}

		interest := engine.analyzeCategoryInterest(viewedProducts)

		// Electronics: 2/3 = 0.67, Books: 1/3 = 0.33
		assert.InDelta(t, 0.67, interest["Electronics"], 0.01)
		assert.InDelta(t, 0.33, interest["Books"], 0.01)
	})
}

// Benchmark tests
func BenchmarkCollaborativeFiltering(b *testing.B) {
	logger := zerolog.Nop()
	engine := NewRecommendationEngine(logger, "benchmark-v1.0")
	ctx := context.Background()
	customerID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CollaborativeFiltering(ctx, customerID, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHybridRecommendations(b *testing.B) {
	logger := zerolog.Nop()
	engine := NewRecommendationEngine(logger, "benchmark-v1.0")
	ctx := context.Background()
	customerID := uuid.New()
	productID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.HybridRecommendations(ctx, &customerID, &productID, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}