package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the ERP system with AI enhancements
type Product struct {
	ID                         uuid.UUID               `json:"id" db:"id"`
	SKU                        string                  `json:"sku" db:"sku"`
	Name                       string                  `json:"name" db:"name"`
	Description                string                  `json:"description" db:"description"`
	IntroductionDate          time.Time               `json:"introductionDate" db:"introduction_date"`
	SalesDiscontinuationDate  *time.Time              `json:"salesDiscontinuationDate,omitempty" db:"sales_discontinuation_date"`
	SupportDiscontinuationDate *time.Time             `json:"supportDiscontinuationDate,omitempty" db:"support_discontinuation_date"`
	Comment                   *string                 `json:"comment,omitempty" db:"comment"`
	ManufacturedByID          *uuid.UUID              `json:"manufacturedById,omitempty" db:"manufactured_by_id"`
	ProductTypeID             uuid.UUID               `json:"productTypeId" db:"product_type_id"`
	UnitOfMeasureID           *uuid.UUID              `json:"unitOfMeasureId,omitempty" db:"unit_of_measure_id"`
	
	// AI-Enhanced Fields
	AIMetadata                AIProductData           `json:"aiMetadata"`
	
	// Relationships
	ProductType               *ProductType            `json:"productType,omitempty"`
	UnitOfMeasure             *UnitOfMeasure          `json:"unitOfMeasure,omitempty"`
	Categories                []ProductCategory       `json:"categories,omitempty"`
	Images                    []ProductImage          `json:"images,omitempty"`
	Features                  []ProductFeature        `json:"features,omitempty"`
	Inventory                 *InventoryInfo          `json:"inventory,omitempty"`
	Pricing                   []ProductPricing        `json:"pricing,omitempty"`
	
	// Audit Fields
	CreatedAt                 time.Time               `json:"createdAt" db:"created_at"`
	UpdatedAt                 time.Time               `json:"updatedAt" db:"updated_at"`
	CreatedBy                 uuid.UUID               `json:"createdBy" db:"created_by"`
	UpdatedBy                 uuid.UUID               `json:"updatedBy" db:"updated_by"`
}

// AIProductData contains AI-generated metadata for products
type AIProductData struct {
	AutoCategory        string                   `json:"autoCategory" db:"auto_category"`
	Confidence          float64                  `json:"confidence" db:"confidence"`
	Tags                []string                 `json:"tags" db:"tags"`
	Recommendations     []RecommendationScore    `json:"recommendations" db:"recommendations"`
	CategoryPredictions []CategoryPrediction     `json:"categoryPredictions" db:"category_predictions"`
	ImageAnalysis       *ImageAnalysisResult     `json:"imageAnalysis,omitempty" db:"image_analysis"`
	PricingSuggestions  []PricingSuggestion      `json:"pricingSuggestions" db:"pricing_suggestions"`
	DemandForecast      *DemandForecast          `json:"demandForecast,omitempty" db:"demand_forecast"`
	CompetitorAnalysis  *CompetitorAnalysis      `json:"competitorAnalysis,omitempty" db:"competitor_analysis"`
}

// ProductType represents the type/classification of a product
type ProductType struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductCategory represents product categorization
type ProductCategory struct {
	ID                  uuid.UUID                       `json:"id" db:"id"`
	Description         string                          `json:"description" db:"description"`
	ParentCategoryID    *uuid.UUID                      `json:"parentCategoryId,omitempty" db:"parent_category_id"`
	Path                string                          `json:"path" db:"path"` // e.g., "Electronics/Computers/Laptops"
	Level               int                             `json:"level" db:"level"`
	AIRecommendations   []CategoryRecommendation        `json:"aiRecommendations,omitempty"`
	MarketInsights      *CategoryMarketInsight          `json:"marketInsights,omitempty"`
	CreatedAt           time.Time                       `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time                       `json:"updatedAt" db:"updated_at"`
}

// ProductImage represents product images with AI analysis
type ProductImage struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	ProductID       uuid.UUID            `json:"productId" db:"product_id"`
	URL             string               `json:"url" db:"url"`
	Alt             string               `json:"alt" db:"alt"`
	IsPrimary       bool                 `json:"isPrimary" db:"is_primary"`
	DisplayOrder    int                  `json:"displayOrder" db:"display_order"`
	ImageAnalysis   *ImageAnalysisResult `json:"imageAnalysis,omitempty"`
	CreatedAt       time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time            `json:"updatedAt" db:"updated_at"`
}

// InventoryInfo represents inventory information for a product
type InventoryInfo struct {
	ProductID            uuid.UUID       `json:"productId" db:"product_id"`
	QuantityOnHand       int64           `json:"quantityOnHand" db:"quantity_on_hand"`
	QuantityAvailable    int64           `json:"quantityAvailable" db:"quantity_available"`
	QuantityReserved     int64           `json:"quantityReserved" db:"quantity_reserved"`
	ReorderLevel         int64           `json:"reorderLevel" db:"reorder_level"`
	ReorderQuantity      int64           `json:"reorderQuantity" db:"reorder_quantity"`
	LastInventoryDate    *time.Time      `json:"lastInventoryDate,omitempty" db:"last_inventory_date"`
	AIOptimization       *InventoryAI    `json:"aiOptimization,omitempty"`
	UpdatedAt            time.Time       `json:"updatedAt" db:"updated_at"`
}

// ProductPricing represents product pricing information
type ProductPricing struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	ProductID         uuid.UUID        `json:"productId" db:"product_id"`
	Price             decimal.Decimal  `json:"price" db:"price"`
	Currency          string           `json:"currency" db:"currency"`
	PriceTypeID       uuid.UUID        `json:"priceTypeId" db:"price_type_id"`
	FromDate          time.Time        `json:"fromDate" db:"from_date"`
	ThruDate          *time.Time       `json:"thruDate,omitempty" db:"thru_date"`
	GeographicScope   *string          `json:"geographicScope,omitempty" db:"geographic_scope"`
	PartyScope        *uuid.UUID       `json:"partyScope,omitempty" db:"party_scope"`
	AIPricing         *AIPricingData   `json:"aiPricing,omitempty"`
	CreatedAt         time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time        `json:"updatedAt" db:"updated_at"`
}

// UnitOfMeasure represents measurement units
type UnitOfMeasure struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Description  string    `json:"description" db:"description"`
	Abbreviation *string   `json:"abbreviation,omitempty" db:"abbreviation"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// ProductFeature represents product features and characteristics
type ProductFeature struct {
	ID                  uuid.UUID    `json:"id" db:"id"`
	ProductID           uuid.UUID    `json:"productId" db:"product_id"`
	FeatureCategoryID   uuid.UUID    `json:"featureCategoryId" db:"feature_category_id"`
	Description         string       `json:"description" db:"description"`
	Value               *string      `json:"value,omitempty" db:"value"`
	NumericValue        *float64     `json:"numericValue,omitempty" db:"numeric_value"`
	UnitOfMeasureID     *uuid.UUID   `json:"unitOfMeasureId,omitempty" db:"unit_of_measure_id"`
	FromDate            time.Time    `json:"fromDate" db:"from_date"`
	ThruDate            *time.Time   `json:"thruDate,omitempty" db:"thru_date"`
	CreatedAt           time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time    `json:"updatedAt" db:"updated_at"`
}