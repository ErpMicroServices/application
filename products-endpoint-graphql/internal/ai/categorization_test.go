package ai

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erpmicroservices/products-endpoint-graphql/pkg/models"
)

func TestCategorizationService(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	service := NewCategorizationService(logger, "test-model-v1.0", "test-api-key")

	t.Run("CategorizeFromImage", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		imageURL := "https://example.com/laptop-image.jpg"

		result, err := service.CategorizeFromImage(ctx, imageURL, productID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ImageID)
		assert.Equal(t, "test-model-v1.0", result.ModelVersion)
		assert.True(t, result.QualityScore > 0.8)
		assert.NotEmpty(t, result.DetectedObjects)
		assert.NotEmpty(t, result.DominantColors)
		assert.NotEmpty(t, result.EstimatedCategories)
		
		// Verify first category prediction
		firstCategory := result.EstimatedCategories[0]
		assert.True(t, firstCategory.Confidence > 0.8)
		assert.NotEmpty(t, firstCategory.CategoryName)
		assert.NotEmpty(t, firstCategory.Reasoning)
	})

	t.Run("CategorizeFromDescription", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		description := "Apple MacBook Pro 16-inch with M2 Max chip, 32GB RAM, 1TB SSD"

		predictions, err := service.CategorizeFromDescription(ctx, description, productID)

		require.NoError(t, err)
		assert.NotEmpty(t, predictions)
		
		// Should categorize as laptop
		foundLaptopCategory := false
		for _, pred := range predictions {
			if pred.CategoryName == "Electronics/Computers/Laptops" {
				foundLaptopCategory = true
				assert.True(t, pred.Confidence > 0.8)
				assert.NotEmpty(t, pred.Reasoning)
			}
		}
		assert.True(t, foundLaptopCategory, "Should identify laptop category from description")
	})

	t.Run("BatchCategorize", func(t *testing.T) {
		ctx := context.Background()
		productIDs := []uuid.UUID{
			uuid.New(),
			uuid.New(),
			uuid.New(),
		}

		startTime := time.Now()
		results, err := service.BatchCategorize(ctx, productIDs)
		duration := time.Since(startTime)

		require.NoError(t, err)
		assert.Len(t, results, len(productIDs))
		
		// Verify each product has predictions
		for _, productID := range productIDs {
			predictions, exists := results[productID]
			assert.True(t, exists, "Should have predictions for product %s", productID)
			assert.NotEmpty(t, predictions)
			assert.True(t, predictions[0].Confidence >= 0.5)
		}

		// Performance check - should complete within reasonable time
		assert.True(t, duration < 2*time.Second, "Batch processing should complete quickly")
	})

	t.Run("DetectNewCategories", func(t *testing.T) {
		ctx := context.Background()
		
		// Create mock products with patterns that suggest new categories
		products := []models.Product{
			{
				ID:          uuid.New(),
				Name:        "Smart Home Hub with IoT Control",
				Description: "IoT device for smart home automation",
			},
			{
				ID:          uuid.New(),
				Name:        "Eco-Friendly Bamboo Laptop Stand",
				Description: "Sustainable laptop stand made from bamboo",
			},
		}

		suggestions, err := service.DetectNewCategories(ctx, products)

		require.NoError(t, err)
		assert.NotEmpty(t, suggestions)
		
		// Should suggest IoT and eco-friendly categories
		categoryNames := make([]string, len(suggestions))
		for i, suggestion := range suggestions {
			categoryNames[i] = suggestion.SuggestedName
			assert.True(t, suggestion.Confidence > 0.5)
			assert.True(t, suggestion.ProductCount > 0)
			assert.NotEmpty(t, suggestion.Evidence)
			assert.NotEmpty(t, suggestion.RecommendedPath)
		}
		
		// Should include smart home and sustainable categories
		assert.Contains(t, categoryNames, "Smart Home/IoT Devices")
		assert.Contains(t, categoryNames, "Sustainable/Eco-Friendly")
	})

	t.Run("CorrectCategorization", func(t *testing.T) {
		ctx := context.Background()
		
		correction := CategoryCorrection{
			ProductID:   uuid.New(),
			OldCategory: "Electronics/Audio/Headphones",
			NewCategory: "Electronics/Audio/Speakers",
			CorrectedBy: uuid.New(),
			CorrectedAt: time.Now(),
			Reason:      "Product is actually a speaker, not headphones",
		}

		err := service.CorrectCategorization(ctx, correction)

		require.NoError(t, err)
		// In a real implementation, this would verify the correction was stored
	})
}

func TestCategorizationPerformance(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	service := NewCategorizationService(logger, "performance-test-v1.0", "test-key")

	t.Run("SingleCategorizationPerformance", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		imageURL := "https://example.com/test-image.jpg"

		startTime := time.Now()
		_, err := service.CategorizeFromImage(ctx, imageURL, productID)
		duration := time.Since(startTime)

		require.NoError(t, err)
		assert.True(t, duration < 5*time.Second, "Single categorization should complete within 5 seconds")
	})

	t.Run("BatchCategorizationPerformance", func(t *testing.T) {
		ctx := context.Background()
		
		// Generate 100 product IDs for performance test
		productIDs := make([]uuid.UUID, 100)
		for i := range productIDs {
			productIDs[i] = uuid.New()
		}

		startTime := time.Now()
		results, err := service.BatchCategorize(ctx, productIDs)
		duration := time.Since(startTime)

		require.NoError(t, err)
		assert.Len(t, results, 100)
		assert.True(t, duration < 2*time.Minute, "Batch categorization of 100 products should complete within 2 minutes")
	})
}

func TestCategorizationEdgeCases(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))
	service := NewCategorizationService(logger, "edge-case-test-v1.0", "test-key")

	t.Run("EmptyDescription", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		description := ""

		predictions, err := service.CategorizeFromDescription(ctx, description, productID)

		require.NoError(t, err)
		assert.NotEmpty(t, predictions)
		
		// Should fall back to uncategorized
		assert.Equal(t, "General/Uncategorized", predictions[0].CategoryName)
		assert.True(t, predictions[0].Confidence <= 0.6) // Low confidence expected
	})

	t.Run("AmbiguousDescription", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		description := "Universal remote control device thingy"

		predictions, err := service.CategorizeFromDescription(ctx, description, productID)

		require.NoError(t, err)
		assert.NotEmpty(t, predictions)
		
		// Should provide multiple suggestions or low confidence
		if len(predictions) == 1 {
			assert.True(t, predictions[0].Confidence < 0.8, "Ambiguous description should have low confidence")
		}
	})

	t.Run("VeryLongDescription", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		
		// Create very long description
		longDescription := ""
		for i := 0; i < 1000; i++ {
			longDescription += "laptop computer electronic device "
		}

		predictions, err := service.CategorizeFromDescription(ctx, longDescription, productID)

		require.NoError(t, err)
		assert.NotEmpty(t, predictions)
		
		// Should still identify key category despite length
		foundCategory := false
		for _, pred := range predictions {
			if pred.CategoryName == "Electronics/Computers/Laptops" {
				foundCategory = true
				break
			}
		}
		assert.True(t, foundCategory, "Should handle long descriptions")
	})

	t.Run("EmptyBatchCategorization", func(t *testing.T) {
		ctx := context.Background()
		productIDs := []uuid.UUID{}

		results, err := service.BatchCategorize(ctx, productIDs)

		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// Benchmark tests for performance validation
func BenchmarkCategorizeFromImage(b *testing.B) {
	logger := zerolog.Nop()
	service := NewCategorizationService(logger, "benchmark-v1.0", "test-key")
	ctx := context.Background()
	productID := uuid.New()
	imageURL := "https://example.com/benchmark-image.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CategorizeFromImage(ctx, imageURL, productID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCategorizeFromDescription(b *testing.B) {
	logger := zerolog.Nop()
	service := NewCategorizationService(logger, "benchmark-v1.0", "test-key")
	ctx := context.Background()
	productID := uuid.New()
	description := "Apple MacBook Pro 16-inch with M2 Max chip, 32GB RAM, 1TB SSD"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CategorizeFromDescription(ctx, description, productID)
		if err != nil {
			b.Fatal(err)
		}
	}
}