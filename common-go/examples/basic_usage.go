// Package main demonstrates basic usage of the ERP common Go library.
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/erpmicroservices/common-go/pkg/audit"
	"github.com/erpmicroservices/common-go/pkg/database"
	"github.com/erpmicroservices/common-go/pkg/errors"
	"github.com/erpmicroservices/common-go/pkg/logging"
	"github.com/erpmicroservices/common-go/pkg/middleware"
	"github.com/erpmicroservices/common-go/pkg/pagination"
	"github.com/erpmicroservices/common-go/pkg/scalars"
	"github.com/erpmicroservices/common-go/pkg/types"
	"github.com/erpmicroservices/common-go/pkg/uuid"
	"github.com/erpmicroservices/common-go/pkg/validation"
)

func main() {
	fmt.Println("ERP Common Go Library - Basic Usage Examples")
	fmt.Println("===========================================")

	// UUID Examples
	demonstrateUUID()

	// Logging Examples
	demonstrateLogging()

	// Validation Examples
	demonstrateValidation()

	// Error Handling Examples
	demonstrateErrorHandling()

	// Audit Examples
	demonstrateAudit()

	// Pagination Examples
	demonstratePagination()

	// Scalars Examples
	demonstrateScalars()

	// Types Examples
	demonstrateTypes()

	// Database Examples
	demonstrateDatabase()

	// Middleware Examples
	demonstrateMiddleware()
}

func demonstrateUUID() {
	fmt.Println("\n1. UUID Examples:")
	fmt.Println("-----------------")

	// Generate new UUID
	id := uuid.New()
	fmt.Printf("Generated UUID: %s\n", id.String())

	// Parse UUID from string
	parsed, err := uuid.NewFromString("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		fmt.Printf("Error parsing UUID: %v\n", err)
	} else {
		fmt.Printf("Parsed UUID: %s\n", parsed.String())
	}

	// Validate UUID
	valid := uuid.IsValid("550e8400-e29b-41d4-a716-446655440000")
	fmt.Printf("UUID is valid: %t\n", valid)

	// JSON marshaling
	data, _ := id.MarshalJSON()
	fmt.Printf("UUID as JSON: %s\n", data)
}

func demonstrateLogging() {
	fmt.Println("\n2. Logging Examples:")
	fmt.Println("-------------------")

	// Create logger
	logger := logging.NewLogger("demo-service")

	// Basic logging
	logger.Info().Msg("Application started")
	logger.Debug().Str("component", "demo").Msg("Debug information")

	// Structured logging with context
	userID := uuid.New()
	contextLogger := logger.WithUserID(userID)
	contextLogger.Info().Msg("User action performed")

	// Performance logging
	timer := logger.StartTimer("database_query")
	time.Sleep(100 * time.Millisecond) // Simulate work
	timer.End("Query completed")
}

func demonstrateValidation() {
	fmt.Println("\n3. Validation Examples:")
	fmt.Println("----------------------")

	// Build validator
	validator := validation.NewValidationBuilder().
		Field("email").Required().Email().
		Field("name").Required().Length(2, 50).
		Field("age").Range(floatPtr(18), floatPtr(120)).
		Build()

	// Valid data
	validData := map[string]interface{}{
		"email": "user@example.com",
		"name":  "John Doe",
		"age":   30,
	}

	err := validator.Validate(validData)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Validation passed!")
	}

	// Invalid data
	invalidData := map[string]interface{}{
		"email": "invalid-email",
		"name":  "X",
		"age":   15,
	}

	err = validator.Validate(invalidData)
	if err != nil {
		fmt.Printf("Validation failed (expected): %v\n", err)
	}

	// Individual validations
	err = validation.ValidatePassword("weakpass")
	if err != nil {
		fmt.Printf("Password validation failed: %v\n", err)
	}
}

func demonstrateErrorHandling() {
	fmt.Println("\n4. Error Handling Examples:")
	fmt.Println("--------------------------")

	// Create different types of errors
	validationErr := errors.Validation("Email is required")
	notFoundErr := errors.NotFound("User")
	businessErr := errors.BusinessRule("Cannot delete user with active orders")
	internalErr := errors.Internal("Database connection failed")

	fmt.Printf("Validation Error: %s (Status: %d)\n", validationErr.Error(), validationErr.GetHTTPStatus())
	fmt.Printf("Not Found Error: %s (Status: %d)\n", notFoundErr.Error(), notFoundErr.GetHTTPStatus())
	fmt.Printf("Business Error: %s (Status: %d)\n", businessErr.Error(), businessErr.GetHTTPStatus())
	fmt.Printf("Internal Error: %s (Status: %d)\n", internalErr.Error(), internalErr.GetHTTPStatus())

	// Error with metadata
	errWithMeta := errors.ValidationWithField("email", "Invalid format").
		WithUserMessage("Please provide a valid email address").
		WithCorrelationID("12345")

	fmt.Printf("Error with metadata: %s\n", errWithMeta.Error())

	// Error list
	errorList := errors.NewErrorList()
	errorList.AddValidation("email", "Required field")
	errorList.AddValidation("name", "Too short")

	if errorList.HasErrors() {
		fmt.Printf("Multiple errors: %v\n", errorList.ToError())
	}
}

func demonstrateAudit() {
	fmt.Println("\n5. Audit Examples:")
	fmt.Println("-----------------")

	userID := uuid.New()

	// Create audit fields
	auditFields := audit.NewAuditFields(userID)
	fmt.Printf("Created audit fields: Created=%v, By=%s\n",
		auditFields.CreatedAt.Format(time.RFC3339), auditFields.CreatedBy.String())

	// Update audit fields
	time.Sleep(1 * time.Millisecond) // Small delay to show time difference
	auditFields.UpdateAuditFields(userID)
	fmt.Printf("Updated audit fields: Updated=%v, Modified=%t\n",
		auditFields.UpdatedAt.Format(time.RFC3339), auditFields.IsModified())

	// Audit context
	ctx := audit.WithUserID(context.Background(), userID)
	ctx = audit.WithCorrelationID(ctx, "corr-123")

	auditInfo := audit.GetAuditInfoFromContext(ctx)
	fmt.Printf("Audit info from context: UserID=%s, CorrelationID=%s\n",
		auditInfo.UserID.String(), auditInfo.CorrelationID)
}

func demonstratePagination() {
	fmt.Println("\n6. Pagination Examples:")
	fmt.Println("----------------------")

	// Cursor-based pagination
	connection := pagination.NewConnection[string]()
	connection.AddEdge(pagination.EncodeIDCursor(uuid.New()), "Item 1")
	connection.AddEdge(pagination.EncodeIDCursor(uuid.New()), "Item 2")
	connection.AddEdge(pagination.EncodeIDCursor(uuid.New()), "Item 3")

	args := pagination.CursorArgs{
		First: intPtr(10),
	}

	connection.UpdatePageInfo(args, 100, true)
	fmt.Printf("Connection has %d edges, total count: %d\n",
		len(connection.Edges), connection.TotalCount)

	// Offset-based pagination
	data := []string{"A", "B", "C", "D", "E"}
	page := pagination.NewOffsetPage(data, 50, 1, 5)
	fmt.Printf("Offset page: %d items, page %d of %d\n",
		len(page.Data), page.Page, page.TotalPages)

	// Pagination builder
	limit, offset, orderBy, orderDir, filters := pagination.NewPaginationBuilder().
		WithLimit(20).
		WithOffset(40).
		WithOrderBy("created_at").
		WithOrderDirection("DESC").
		WithFilter("status", "active").
		Build()

	fmt.Printf("Builder result: limit=%d, offset=%d, orderBy=%s %s, filters=%v\n",
		limit, offset, orderBy, orderDir, filters)
}

func demonstrateScalars() {
	fmt.Println("\n7. Scalars Examples:")
	fmt.Println("-------------------")

	// Money
	money := scalars.NewMoneyFromFloat(99.99, "USD")
	fmt.Printf("Money: %s\n", money.String())

	tax := scalars.NewMoneyFromFloat(8.50, "USD")
	total, _ := money.Add(tax)
	fmt.Printf("Total with tax: %s\n", total.String())

	// Percentage
	discount := scalars.NewPercentageFromFloat(15.0)
	fmt.Printf("Discount: %s\n", discount.String())

	// Date
	today := scalars.Today()
	fmt.Printf("Today: %s\n", today.String())

	// Email
	email := scalars.Email("user@example.com")
	fmt.Printf("Email valid: %t\n", email.IsValid())
}

func demonstrateTypes() {
	fmt.Println("\n8. Types Examples:")
	fmt.Println("-----------------")

	userID := uuid.New()

	// Base entity
	entity := types.NewBaseEntity(userID)
	fmt.Printf("Entity ID: %s\n", entity.GetID().String())

	// Type entity
	statusType := types.NewTypeEntity("ACTIVE", "Active Status", "Active status description", userID)
	fmt.Printf("Type entity: %s - %s\n", statusType.Code, statusType.Name)

	// Status enum
	status := types.StatusActive
	fmt.Printf("Status is valid: %t\n", status.IsValid())

	// Priority enum
	priority := types.PriorityHigh
	fmt.Printf("Priority numeric value: %d\n", priority.GetNumericValue())

	// Address
	address := types.Address{
		Street1:    "123 Main St",
		City:       "Anytown",
		State:      "CA",
		PostalCode: "12345",
		Country:    "US",
	}
	fmt.Printf("Full address: %s\n", address.GetFullAddress())

	// Date range
	dateRange := types.NewDateRange(time.Now().AddDate(0, -1, 0), nil)
	fmt.Printf("Date range is active: %t\n", dateRange.IsActive())
}

func demonstrateDatabase() {
	fmt.Println("\n9. Database Examples:")
	fmt.Println("--------------------")

	// Database configuration
	config := database.DefaultConfig()
	config.Database = "demo_db"
	fmt.Printf("Database DSN: %s\n", config.DSN())

	// Null types
	nullStr := database.NewNullString(stringPtr("Hello"))
	fmt.Printf("Null string valid: %t, value: %s\n", nullStr.Valid, nullStr.String)

	nullInt := database.NewNullInt64(int64Ptr(42))
	fmt.Printf("Null int64 valid: %t, value: %d\n", nullInt.Valid, nullInt.Int64)

	// Note: Actual database connection would require a real database
	fmt.Println("Database connection examples require a real database connection")
}

func demonstrateMiddleware() {
	fmt.Println("\n10. Middleware Examples:")
	fmt.Println("-----------------------")

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	// CORS middleware
	corsConfig := middleware.DefaultCORSConfig()
	corsMiddleware := middleware.CORS(corsConfig)

	// Security headers middleware
	securityMiddleware := middleware.SecurityHeaders()

	// Correlation ID middleware
	correlationMiddleware := middleware.CorrelationID()

	// Chain middlewares
	chainedHandler := middleware.Chain(
		corsMiddleware,
		securityMiddleware,
		correlationMiddleware,
	)(handler)

	fmt.Println("Created chained middleware handler")
	fmt.Printf("Middleware chain includes: CORS, Security Headers, Correlation ID\n")

	// Rate limiting
	rateLimitConfig := middleware.DefaultRateLimitConfig()
	rateLimitMiddleware := middleware.RateLimit(rateLimitConfig)

	finalHandler := rateLimitMiddleware(chainedHandler)
	fmt.Println("Added rate limiting middleware")

	// Note: To actually use these, you would start an HTTP server
	// http.ListenAndServe(":8080", finalHandler)
	_ = finalHandler // Avoid unused variable warning
}

// Helper functions for creating pointers
func stringPtr(s string) *string  { return &s }
func intPtr(i int) *int           { return &i }
func int64Ptr(i int64) *int64     { return &i }
func floatPtr(f float64) *float64 { return &f }
