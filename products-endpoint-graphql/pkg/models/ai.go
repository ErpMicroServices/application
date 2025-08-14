package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RecommendationScore represents AI-generated product recommendations
type RecommendationScore struct {
	ProductID        uuid.UUID `json:"productId" db:"product_id"`
	Score            float64   `json:"score" db:"score"`
	ReasonCode       string    `json:"reasonCode" db:"reason_code"`
	Explanation      string    `json:"explanation" db:"explanation"`
	ConfidenceLevel  float64   `json:"confidenceLevel" db:"confidence_level"`
	GeneratedAt      time.Time `json:"generatedAt" db:"generated_at"`
}

// CategoryPrediction represents AI category predictions
type CategoryPrediction struct {
	CategoryID      uuid.UUID `json:"categoryId" db:"category_id"`
	CategoryName    string    `json:"categoryName" db:"category_name"`
	Confidence      float64   `json:"confidence" db:"confidence"`
	Reasoning       []string  `json:"reasoning" db:"reasoning"`
	ModelVersion    string    `json:"modelVersion" db:"model_version"`
	PredictedAt     time.Time `json:"predictedAt" db:"predicted_at"`
}

// ImageAnalysisResult contains results from computer vision analysis
type ImageAnalysisResult struct {
	ImageID             uuid.UUID                `json:"imageId" db:"image_id"`
	DetectedObjects     []DetectedObject         `json:"detectedObjects" db:"detected_objects"`
	DominantColors      []ColorInfo              `json:"dominantColors" db:"dominant_colors"`
	EstimatedCategories []CategoryPrediction     `json:"estimatedCategories" db:"estimated_categories"`
	QualityScore        float64                  `json:"qualityScore" db:"quality_score"`
	TechnicalSpecs      *TechnicalSpecsFromImage `json:"technicalSpecs,omitempty" db:"technical_specs"`
	TextRecognition     []RecognizedText         `json:"textRecognition" db:"text_recognition"`
	BrandDetection      *BrandInfo               `json:"brandDetection,omitempty" db:"brand_detection"`
	StyleTags           []string                 `json:"styleTags" db:"style_tags"`
	AnalyzedAt          time.Time                `json:"analyzedAt" db:"analyzed_at"`
	ModelVersion        string                   `json:"modelVersion" db:"model_version"`
}

// DetectedObject represents objects detected in product images
type DetectedObject struct {
	Label       string            `json:"label" db:"label"`
	Confidence  float64           `json:"confidence" db:"confidence"`
	BoundingBox *BoundingBox      `json:"boundingBox,omitempty" db:"bounding_box"`
	Attributes  map[string]string `json:"attributes" db:"attributes"`
}

// BoundingBox represents object location in image
type BoundingBox struct {
	X      float64 `json:"x" db:"x"`
	Y      float64 `json:"y" db:"y"`
	Width  float64 `json:"width" db:"width"`
	Height float64 `json:"height" db:"height"`
}

// ColorInfo represents color analysis results
type ColorInfo struct {
	HexCode    string  `json:"hexCode" db:"hex_code"`
	Percentage float64 `json:"percentage" db:"percentage"`
	ColorName  string  `json:"colorName" db:"color_name"`
	Prominence float64 `json:"prominence" db:"prominence"`
}

// TechnicalSpecsFromImage represents technical specifications extracted from images
type TechnicalSpecsFromImage struct {
	EstimatedDimensions *Dimensions           `json:"estimatedDimensions,omitempty" db:"estimated_dimensions"`
	MaterialGuess       []string              `json:"materialGuess" db:"material_guess"`
	ConditionAssessment string                `json:"conditionAssessment" db:"condition_assessment"`
	Features            []string              `json:"features" db:"features"`
	Defects             []string              `json:"defects" db:"defects"`
	ExtractedSpecs      map[string]string     `json:"extractedSpecs" db:"extracted_specs"`
}

// Dimensions represents estimated product dimensions
type Dimensions struct {
	Length decimal.Decimal `json:"length" db:"length"`
	Width  decimal.Decimal `json:"width" db:"width"`
	Height decimal.Decimal `json:"height" db:"height"`
	Unit   string          `json:"unit" db:"unit"`
}

// RecognizedText represents OCR results from product images
type RecognizedText struct {
	Text       string       `json:"text" db:"text"`
	Confidence float64      `json:"confidence" db:"confidence"`
	Language   string       `json:"language" db:"language"`
	Location   *BoundingBox `json:"location,omitempty" db:"location"`
	TextType   string       `json:"textType" db:"text_type"` // "label", "specification", "brand", etc.
}

// BrandInfo represents brand detection results
type BrandInfo struct {
	BrandName   string    `json:"brandName" db:"brand_name"`
	Confidence  float64   `json:"confidence" db:"confidence"`
	LogoFound   bool      `json:"logoFound" db:"logo_found"`
	TextBased   bool      `json:"textBased" db:"text_based"`
	Location    *BoundingBox `json:"location,omitempty" db:"location"`
}

// PricingSuggestion represents AI-generated pricing recommendations
type PricingSuggestion struct {
	SuggestionID     uuid.UUID       `json:"suggestionId" db:"suggestion_id"`
	ProductID        uuid.UUID       `json:"productId" db:"product_id"`
	SuggestedPrice   decimal.Decimal `json:"suggestedPrice" db:"suggested_price"`
	Currency         string          `json:"currency" db:"currency"`
	PriceRange       *PriceRange     `json:"priceRange,omitempty" db:"price_range"`
	ReasonCode       string          `json:"reasonCode" db:"reason_code"`
	Factors          []PricingFactor `json:"factors" db:"factors"`
	MarketPosition   string          `json:"marketPosition" db:"market_position"` // "premium", "competitive", "budget"
	ConfidenceScore  float64         `json:"confidenceScore" db:"confidence_score"`
	ValidUntil       time.Time       `json:"validUntil" db:"valid_until"`
	GeneratedAt      time.Time       `json:"generatedAt" db:"generated_at"`
	ModelVersion     string          `json:"modelVersion" db:"model_version"`
}

// PriceRange represents a price range suggestion
type PriceRange struct {
	MinPrice decimal.Decimal `json:"minPrice" db:"min_price"`
	MaxPrice decimal.Decimal `json:"maxPrice" db:"max_price"`
	OptimalPrice decimal.Decimal `json:"optimalPrice" db:"optimal_price"`
}

// PricingFactor represents factors influencing pricing suggestions
type PricingFactor struct {
	Factor      string  `json:"factor" db:"factor"`
	Impact      string  `json:"impact" db:"impact"` // "positive", "negative", "neutral"
	Weight      float64 `json:"weight" db:"weight"`
	Description string  `json:"description" db:"description"`
}

// DemandForecast represents AI-generated demand predictions
type DemandForecast struct {
	ProductID           uuid.UUID           `json:"productId" db:"product_id"`
	ForecastPeriod      string              `json:"forecastPeriod" db:"forecast_period"` // "week", "month", "quarter"
	PredictedDemand     int64               `json:"predictedDemand" db:"predicted_demand"`
	ConfidenceInterval  *ConfidenceInterval `json:"confidenceInterval,omitempty" db:"confidence_interval"`
	SeasonalFactors     []SeasonalFactor    `json:"seasonalFactors" db:"seasonal_factors"`
	TrendDirection      string              `json:"trendDirection" db:"trend_direction"` // "up", "down", "stable"
	InfluencingFactors  []DemandFactor      `json:"influencingFactors" db:"influencing_factors"`
	AccuracyScore       float64             `json:"accuracyScore" db:"accuracy_score"`
	LastUpdated         time.Time           `json:"lastUpdated" db:"last_updated"`
	ModelVersion        string              `json:"modelVersion" db:"model_version"`
}

// ConfidenceInterval represents statistical confidence interval
type ConfidenceInterval struct {
	LowerBound int64   `json:"lowerBound" db:"lower_bound"`
	UpperBound int64   `json:"upperBound" db:"upper_bound"`
	Level      float64 `json:"level" db:"level"` // e.g., 0.95 for 95% confidence
}

// SeasonalFactor represents seasonal demand patterns
type SeasonalFactor struct {
	Period       string  `json:"period" db:"period"` // "Q1", "summer", "holiday", etc.
	Multiplier   float64 `json:"multiplier" db:"multiplier"`
	Historical   bool    `json:"historical" db:"historical"`
	Description  string  `json:"description" db:"description"`
}

// DemandFactor represents factors influencing demand
type DemandFactor struct {
	Factor      string  `json:"factor" db:"factor"`
	Impact      string  `json:"impact" db:"impact"` // "positive", "negative", "neutral"
	Strength    float64 `json:"strength" db:"strength"`
	Description string  `json:"description" db:"description"`
}

// CompetitorAnalysis represents competitive analysis data
type CompetitorAnalysis struct {
	ProductID              uuid.UUID             `json:"productId" db:"product_id"`
	CompetitorProducts     []CompetitorProduct   `json:"competitorProducts" db:"competitor_products"`
	MarketPosition         string                `json:"marketPosition" db:"market_position"`
	CompetitiveAdvantages  []string              `json:"competitiveAdvantages" db:"competitive_advantages"`
	CompetitiveWeaknesses  []string              `json:"competitiveWeaknesses" db:"competitive_weaknesses"`
	RecommendedActions     []RecommendedAction   `json:"recommendedActions" db:"recommended_actions"`
	MarketShare           *decimal.Decimal       `json:"marketShare,omitempty" db:"market_share"`
	PricePositioning      string                `json:"pricePositioning" db:"price_positioning"`
	AnalyzedAt            time.Time             `json:"analyzedAt" db:"analyzed_at"`
	DataSources           []string              `json:"dataSources" db:"data_sources"`
}

// CompetitorProduct represents a competing product
type CompetitorProduct struct {
	ProductName     string           `json:"productName" db:"product_name"`
	Brand          string           `json:"brand" db:"brand"`
	Price          *decimal.Decimal `json:"price,omitempty" db:"price"`
	Currency       string           `json:"currency" db:"currency"`
	Features       []string         `json:"features" db:"features"`
	Rating         *float64         `json:"rating,omitempty" db:"rating"`
	ReviewCount    *int64           `json:"reviewCount,omitempty" db:"review_count"`
	Availability   string           `json:"availability" db:"availability"`
	URL            *string          `json:"url,omitempty" db:"url"`
	SimilarityScore float64         `json:"similarityScore" db:"similarity_score"`
}

// RecommendedAction represents recommended competitive actions
type RecommendedAction struct {
	Action      string    `json:"action" db:"action"`
	Priority    string    `json:"priority" db:"priority"` // "high", "medium", "low"
	Timeline    string    `json:"timeline" db:"timeline"`
	Description string    `json:"description" db:"description"`
	Impact      string    `json:"impact" db:"impact"`
}

// InventoryAI represents AI-driven inventory optimization
type InventoryAI struct {
	ProductID               uuid.UUID   `json:"productId" db:"product_id"`
	OptimalStockLevel       int64       `json:"optimalStockLevel" db:"optimal_stock_level"`
	RecommendedReorderPoint int64       `json:"recommendedReorderPoint" db:"recommended_reorder_point"`
	RecommendedOrderQuantity int64      `json:"recommendedOrderQuantity" db:"recommended_order_quantity"`
	TurnoverPrediction      float64     `json:"turnoverPrediction" db:"turnover_prediction"`
	SeasonalAdjustments     []int64     `json:"seasonalAdjustments" db:"seasonal_adjustments"`
	StockoutRisk            float64     `json:"stockoutRisk" db:"stockout_risk"`
	OverstockRisk           float64     `json:"overstockRisk" db:"overstock_risk"`
	OptimizationReasons     []string    `json:"optimizationReasons" db:"optimization_reasons"`
	LastOptimized          time.Time    `json:"lastOptimized" db:"last_optimized"`
	ModelVersion           string       `json:"modelVersion" db:"model_version"`
}

// AIPricingData represents AI-enhanced pricing information
type AIPricingData struct {
	PriceOptimizationScore float64           `json:"priceOptimizationScore" db:"price_optimization_score"`
	ElasticityEstimate     float64           `json:"elasticityEstimate" db:"elasticity_estimate"`
	CompetitiveDifferential decimal.Decimal `json:"competitiveDifferential" db:"competitive_differential"`
	SuggestedAdjustments   []PriceAdjustment `json:"suggestedAdjustments" db:"suggested_adjustments"`
	MarketSentiment        string            `json:"marketSentiment" db:"market_sentiment"`
	RevenueImpactForecast  decimal.Decimal   `json:"revenueImpactForecast" db:"revenue_impact_forecast"`
	LastAnalyzed          time.Time          `json:"lastAnalyzed" db:"last_analyzed"`
}

// PriceAdjustment represents suggested price adjustments
type PriceAdjustment struct {
	AdjustmentType string          `json:"adjustmentType" db:"adjustment_type"` // "increase", "decrease", "seasonal"
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	Percentage     float64         `json:"percentage" db:"percentage"`
	Reason         string          `json:"reason" db:"reason"`
	Timeline       string          `json:"timeline" db:"timeline"`
	ImpactScore    float64         `json:"impactScore" db:"impact_score"`
}

// CategoryRecommendation represents AI category recommendations
type CategoryRecommendation struct {
	CategoryID      uuid.UUID `json:"categoryId" db:"category_id"`
	RecommendationType string `json:"recommendationType" db:"recommendation_type"` // "trending", "cross_sell", "upsell"
	Score           float64   `json:"score" db:"score"`
	Reason          string    `json:"reason" db:"reason"`
	ValidUntil      time.Time `json:"validUntil" db:"valid_until"`
}

// CategoryMarketInsight represents market insights for categories
type CategoryMarketInsight struct {
	CategoryID         uuid.UUID `json:"categoryId" db:"category_id"`
	Growth            float64   `json:"growth" db:"growth"`
	Seasonality       []float64 `json:"seasonality" db:"seasonality"`
	CompetitiveIndex  float64   `json:"competitiveIndex" db:"competitive_index"`
	ProfitMarginAvg   float64   `json:"profitMarginAvg" db:"profit_margin_avg"`
	TrendDirection    string    `json:"trendDirection" db:"trend_direction"`
	EmergingKeywords  []string  `json:"emergingKeywords" db:"emerging_keywords"`
	InsightDate       time.Time `json:"insightDate" db:"insight_date"`
}