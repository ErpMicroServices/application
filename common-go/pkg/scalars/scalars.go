// Package scalars provides custom GraphQL scalar types for the ERP microservices system.
// It includes DateTime, UUID, Money, Date, and other domain-specific scalar types.
package scalars

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/erpmicroservices/common-go/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/vektah/gqlparser/v2/ast"
)

// DateTime represents a date and time in RFC3339 format with timezone information.
// This scalar type handles serialization/deserialization of time.Time values.
func MarshalDateTime(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
	})
}

func UnmarshalDateTime(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(string); ok {
		return time.Parse(time.RFC3339, tmpStr)
	}
	return time.Time{}, fmt.Errorf("time should be RFC3339 formatted string")
}

// Date represents a date-only value in YYYY-MM-DD format.
type Date time.Time

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) Time() time.Time {
	return time.Time(d)
}

func (d Date) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.String())
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" || str == "null" {
		*d = Date(time.Time{})
		return nil
	}

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	*d = Date(t)
	return nil
}

func MarshalDate(d Date) graphql.Marshaler {
	if time.Time(d).IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(d.String()))
	})
}

func UnmarshalDate(v interface{}) (Date, error) {
	if tmpStr, ok := v.(string); ok {
		t, err := time.Parse("2006-01-02", tmpStr)
		return Date(t), err
	}
	return Date{}, fmt.Errorf("date should be YYYY-MM-DD formatted string")
}

// NewDate creates a new Date from year, month, and day.
func NewDate(year int, month time.Month, day int) Date {
	return Date(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

// Today returns the current date.
func Today() Date {
	now := time.Now()
	return Date(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
}

// UUID scalar marshaling functions
func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	if id.IsNil() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(id.String()))
	})
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	if tmpStr, ok := v.(string); ok {
		return uuid.NewFromString(tmpStr)
	}
	return uuid.UUID{}, fmt.Errorf("UUID should be a valid UUID string")
}

// Money represents a monetary value with precision.
type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// NewMoney creates a new Money instance.
func NewMoney(amount decimal.Decimal, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: strings.ToUpper(currency),
	}
}

// NewMoneyFromFloat creates a new Money instance from a float64.
func NewMoneyFromFloat(amount float64, currency string) Money {
	return Money{
		Amount:   decimal.NewFromFloat(amount),
		Currency: strings.ToUpper(currency),
	}
}

// NewMoneyFromString creates a new Money instance from a string amount.
func NewMoneyFromString(amount, currency string) (Money, error) {
	dec, err := decimal.NewFromString(amount)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount format: %w", err)
	}

	return Money{
		Amount:   dec,
		Currency: strings.ToUpper(currency),
	}, nil
}

// String returns a string representation of the money value.
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.String(), m.Currency)
}

// IsZero returns true if the money amount is zero.
func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

// Add adds another money value to this one.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}

	return Money{
		Amount:   m.Amount.Add(other.Amount),
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts another money value from this one.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}

	return Money{
		Amount:   m.Amount.Sub(other.Amount),
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies the money amount by a factor.
func (m Money) Multiply(factor decimal.Decimal) Money {
	return Money{
		Amount:   m.Amount.Mul(factor),
		Currency: m.Currency,
	}
}

// MultiplyFloat multiplies the money amount by a float factor.
func (m Money) MultiplyFloat(factor float64) Money {
	return m.Multiply(decimal.NewFromFloat(factor))
}

// Divide divides the money amount by a factor.
func (m Money) Divide(factor decimal.Decimal) Money {
	if factor.IsZero() {
		return m // Avoid division by zero, return original value
	}

	return Money{
		Amount:   m.Amount.Div(factor),
		Currency: m.Currency,
	}
}

// Equal returns true if two money values are equal (same currency and amount).
func (m Money) Equal(other Money) bool {
	return m.Currency == other.Currency && m.Amount.Equal(other.Amount)
}

// GreaterThan returns true if this money value is greater than the other.
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		return false // Cannot compare different currencies
	}
	return m.Amount.GreaterThan(other.Amount)
}

// LessThan returns true if this money value is less than the other.
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		return false // Cannot compare different currencies
	}
	return m.Amount.LessThan(other.Amount)
}

// MarshalJSON implements json.Marshaler.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	}{
		Amount:   m.Amount.String(),
		Currency: m.Currency,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Money) UnmarshalJSON(data []byte) error {
	var temp struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	amount, err := decimal.NewFromString(temp.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount format: %w", err)
	}

	m.Amount = amount
	m.Currency = strings.ToUpper(temp.Currency)
	return nil
}

func MarshalMoney(m Money) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		data, _ := json.Marshal(m)
		w.Write(data)
	})
}

func UnmarshalMoney(v interface{}) (Money, error) {
	switch v := v.(type) {
	case string:
		// Parse from string format like "100.50 USD"
		parts := strings.Fields(v)
		if len(parts) != 2 {
			return Money{}, fmt.Errorf("invalid money format, expected 'amount currency'")
		}
		return NewMoneyFromString(parts[0], parts[1])
	case map[string]interface{}:
		// Parse from object format
		data, err := json.Marshal(v)
		if err != nil {
			return Money{}, err
		}
		var money Money
		err = json.Unmarshal(data, &money)
		return money, err
	default:
		return Money{}, fmt.Errorf("money should be a string or object")
	}
}

// Percentage represents a percentage value.
type Percentage decimal.Decimal

// NewPercentage creates a new percentage from a decimal value.
func NewPercentage(value decimal.Decimal) Percentage {
	return Percentage(value)
}

// NewPercentageFromFloat creates a new percentage from a float64.
func NewPercentageFromFloat(value float64) Percentage {
	return Percentage(decimal.NewFromFloat(value))
}

// String returns the percentage as a string with % symbol.
func (p Percentage) String() string {
	return decimal.Decimal(p).String() + "%"
}

// Decimal returns the underlying decimal value.
func (p Percentage) Decimal() decimal.Decimal {
	return decimal.Decimal(p)
}

// ToRatio converts percentage to a ratio (e.g., 50% -> 0.5).
func (p Percentage) ToRatio() decimal.Decimal {
	return decimal.Decimal(p).Div(decimal.NewFromInt(100))
}

func MarshalPercentage(p Percentage) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(p.String()))
	})
}

func UnmarshalPercentage(v interface{}) (Percentage, error) {
	switch v := v.(type) {
	case string:
		// Remove % symbol if present
		str := strings.TrimSuffix(v, "%")
		dec, err := decimal.NewFromString(str)
		return Percentage(dec), err
	case float64:
		return NewPercentageFromFloat(v), nil
	case int:
		return NewPercentageFromFloat(float64(v)), nil
	default:
		return Percentage{}, fmt.Errorf("percentage should be a string with %% or a number")
	}
}

// JSON scalar type for handling arbitrary JSON data
type JSON map[string]interface{}

func MarshalJSON(j JSON) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		data, _ := json.Marshal(j)
		w.Write(data)
	})
}

func UnmarshalJSON(v interface{}) (JSON, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		return JSON(v), nil
	case string:
		var result JSON
		err := json.Unmarshal([]byte(v), &result)
		return result, err
	default:
		return nil, fmt.Errorf("JSON should be an object or JSON string")
	}
}

// Email scalar type for email addresses
type Email string

func (e Email) String() string {
	return string(e)
}

func (e Email) IsValid() bool {
	// Basic email validation
	str := string(e)
	return strings.Contains(str, "@") && strings.Contains(str, ".")
}

func MarshalEmail(e Email) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(string(e)))
	})
}

func UnmarshalEmail(v interface{}) (Email, error) {
	if str, ok := v.(string); ok {
		email := Email(str)
		if !email.IsValid() {
			return "", fmt.Errorf("invalid email format")
		}
		return email, nil
	}
	return "", fmt.Errorf("email should be a string")
}

// URL scalar type for URLs
type URL string

func (u URL) String() string {
	return string(u)
}

func MarshalURL(u URL) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(string(u)))
	})
}

func UnmarshalURL(v interface{}) (URL, error) {
	if str, ok := v.(string); ok {
		if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
			return "", fmt.Errorf("invalid URL format, must start with http:// or https://")
		}
		return URL(str), nil
	}
	return "", fmt.Errorf("URL should be a string")
}

// PhoneNumber scalar type for phone numbers
type PhoneNumber string

func (p PhoneNumber) String() string {
	return string(p)
}

func MarshalPhoneNumber(p PhoneNumber) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(string(p)))
	})
}

func UnmarshalPhoneNumber(v interface{}) (PhoneNumber, error) {
	if str, ok := v.(string); ok {
		// Basic phone number validation (contains only digits, spaces, +, -, (, ))
		for _, char := range str {
			if !strings.ContainsRune("0123456789 +-()", char) {
				return "", fmt.Errorf("invalid phone number format")
			}
		}
		return PhoneNumber(str), nil
	}
	return "", fmt.Errorf("phone number should be a string")
}

// Upload scalar type for file uploads
type Upload struct {
	File     graphql.Upload
	Filename string
	Size     int64
}

func MarshalUpload(u Upload) graphql.Marshaler {
	// Uploads are input-only, so this should not be called
	return graphql.Null
}

func UnmarshalUpload(v interface{}) (Upload, error) {
	upload, ok := v.(graphql.Upload)
	if !ok {
		return Upload{}, fmt.Errorf("upload should be a file upload")
	}

	return Upload{
		File:     upload,
		Filename: upload.Filename,
		Size:     upload.Size,
	}, nil
}

// Void scalar type for mutations that don't return data
type Void struct{}

func MarshalVoid(v Void) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "true")
	})
}

func UnmarshalVoid(v interface{}) (Void, error) {
	return Void{}, nil
}

// Helper functions for scalar registration

// RegisterScalars registers all custom scalars with a GraphQL schema.
func RegisterScalars() map[string]interface{} {
	return map[string]interface{}{
		"DateTime":    MarshalDateTime,
		"Date":        MarshalDate,
		"UUID":        MarshalUUID,
		"Money":       MarshalMoney,
		"Percentage":  MarshalPercentage,
		"JSON":        MarshalJSON,
		"Email":       MarshalEmail,
		"URL":         MarshalURL,
		"PhoneNumber": MarshalPhoneNumber,
		"Upload":      MarshalUpload,
		"Void":        MarshalVoid,
	}
}

// DirectiveConfig provides configuration for GraphQL directives related to scalars.
type DirectiveConfig struct {
	Name        string
	Description string
	Locations   []ast.DirectiveLocation
}

// GetScalarDirectives returns directive configurations for scalar validation.
func GetScalarDirectives() []DirectiveConfig {
	return []DirectiveConfig{
		{
			Name:        "currency",
			Description: "Validates that a Money value uses the specified currency",
			Locations:   []ast.DirectiveLocation{ast.LocationFieldDefinition, ast.LocationInputFieldDefinition},
		},
		{
			Name:        "range",
			Description: "Validates that a numeric value is within the specified range",
			Locations:   []ast.DirectiveLocation{ast.LocationFieldDefinition, ast.LocationInputFieldDefinition},
		},
		{
			Name:        "length",
			Description: "Validates that a string value has the specified length constraints",
			Locations:   []ast.DirectiveLocation{ast.LocationFieldDefinition, ast.LocationInputFieldDefinition},
		},
		{
			Name:        "format",
			Description: "Validates that a string value matches the specified format",
			Locations:   []ast.DirectiveLocation{ast.LocationFieldDefinition, ast.LocationInputFieldDefinition},
		},
	}
}
