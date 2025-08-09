// Package validation provides common validation functions for business logic
// in the ERP microservices system. It includes validators for email, phone numbers,
// addresses, and other domain-specific data types.
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/erpmicroservices/common-go/pkg/errors"
	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// ValidationRule represents a validation rule that can be applied to a value.
type ValidationRule interface {
	Validate(value interface{}) error
}

// Validator provides validation functionality for various data types.
type Validator struct {
	rules map[string][]ValidationRule
}

// NewValidator creates a new validator instance.
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule for a specific field.
func (v *Validator) AddRule(field string, rule ValidationRule) *Validator {
	v.rules[field] = append(v.rules[field], rule)
	return v
}

// Validate validates a map of field values against the configured rules.
func (v *Validator) Validate(values map[string]interface{}) error {
	errorList := errors.NewErrorList()

	for field, rules := range v.rules {
		value, exists := values[field]

		for _, rule := range rules {
			if !exists || value == nil {
				if _, ok := rule.(RequiredRule); ok {
					errorList.AddValidation(field, "field is required")
				}
				continue
			}

			if err := rule.Validate(value); err != nil {
				var message string
				if validationErr, ok := err.(*errors.ERPError); ok {
					message = validationErr.Message
				} else {
					message = err.Error()
				}
				errorList.AddValidation(field, message)
			}
		}
	}

	return errorList.ToError()
}

// Built-in validation rules

// RequiredRule validates that a field is present and not empty.
type RequiredRule struct{}

func (r RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return errors.Validation("field is required")
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return errors.Validation("field cannot be empty")
		}
	case []interface{}:
		if len(v) == 0 {
			return errors.Validation("field cannot be empty")
		}
	}

	return nil
}

// Required creates a required validation rule.
func Required() RequiredRule {
	return RequiredRule{}
}

// LengthRule validates string length constraints.
type LengthRule struct {
	Min int
	Max int
}

func (r LengthRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.Validation("value must be a string")
	}

	length := len(strings.TrimSpace(str))

	if r.Min > 0 && length < r.Min {
		return errors.Validation(fmt.Sprintf("minimum length is %d characters", r.Min))
	}

	if r.Max > 0 && length > r.Max {
		return errors.Validation(fmt.Sprintf("maximum length is %d characters", r.Max))
	}

	return nil
}

// Length creates a length validation rule.
func Length(min, max int) LengthRule {
	return LengthRule{Min: min, Max: max}
}

// MinLength creates a minimum length validation rule.
func MinLength(min int) LengthRule {
	return LengthRule{Min: min}
}

// MaxLength creates a maximum length validation rule.
func MaxLength(max int) LengthRule {
	return LengthRule{Max: max}
}

// EmailRule validates email addresses.
type EmailRule struct {
	Strict bool
}

var (
	emailRegexBasic  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	emailRegexStrict = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
)

func (r EmailRule) Validate(value interface{}) error {
	email, ok := value.(string)
	if !ok {
		return errors.Validation("email must be a string")
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return errors.Validation("email cannot be empty")
	}

	regex := emailRegexBasic
	if r.Strict {
		regex = emailRegexStrict
	}

	if !regex.MatchString(email) {
		return errors.Validation("invalid email format")
	}

	// Additional checks for strict validation
	if r.Strict {
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return errors.Validation("invalid email format")
		}

		localPart := parts[0]
		domain := parts[1]

		// Check local part length (64 characters max)
		if len(localPart) > 64 {
			return errors.Validation("email local part too long")
		}

		// Check domain length (253 characters max)
		if len(domain) > 253 {
			return errors.Validation("email domain too long")
		}

		// Check for consecutive dots
		if strings.Contains(email, "..") {
			return errors.Validation("email cannot contain consecutive dots")
		}
	}

	return nil
}

// Email creates an email validation rule.
func Email() EmailRule {
	return EmailRule{Strict: false}
}

// EmailStrict creates a strict email validation rule.
func EmailStrict() EmailRule {
	return EmailRule{Strict: true}
}

// PhoneRule validates phone numbers.
type PhoneRule struct {
	AllowInternational bool
	RequireCountryCode bool
}

var phoneRegex = regexp.MustCompile(`^[\+]?[1-9][\d\s\-\(\)\.]{7,15}$`)

func (r PhoneRule) Validate(value interface{}) error {
	phone, ok := value.(string)
	if !ok {
		return errors.Validation("phone number must be a string")
	}

	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.Validation("phone number cannot be empty")
	}

	// Remove common separators for validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ".", "")

	if !phoneRegex.MatchString(phone) {
		return errors.Validation("invalid phone number format")
	}

	// Check international format requirements
	if r.RequireCountryCode && !strings.HasPrefix(cleanPhone, "+") {
		return errors.Validation("phone number must include country code")
	}

	if !r.AllowInternational && strings.HasPrefix(cleanPhone, "+") {
		return errors.Validation("international phone numbers not allowed")
	}

	// Check length after cleaning
	digitCount := 0
	for _, char := range cleanPhone {
		if unicode.IsDigit(char) {
			digitCount++
		}
	}

	if digitCount < 7 || digitCount > 15 {
		return errors.Validation("phone number must contain 7-15 digits")
	}

	return nil
}

// Phone creates a phone number validation rule.
func Phone() PhoneRule {
	return PhoneRule{AllowInternational: true, RequireCountryCode: false}
}

// PhoneStrict creates a strict phone number validation rule.
func PhoneStrict() PhoneRule {
	return PhoneRule{AllowInternational: true, RequireCountryCode: true}
}

// PhoneUS creates a US phone number validation rule.
func PhoneUS() PhoneRule {
	return PhoneRule{AllowInternational: false, RequireCountryCode: false}
}

// UUIDRule validates UUID strings.
type UUIDRule struct{}

func (r UUIDRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.Validation("UUID must be a string")
	}

	if !uuid.IsValid(str) {
		return errors.Validation("invalid UUID format")
	}

	return nil
}

// UUID creates a UUID validation rule.
func UUID() UUIDRule {
	return UUIDRule{}
}

// RangeRule validates numeric ranges.
type RangeRule struct {
	Min *float64
	Max *float64
}

func (r RangeRule) Validate(value interface{}) error {
	var num float64
	var ok bool

	switch v := value.(type) {
	case int:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case float32:
		num = float64(v)
		ok = true
	case float64:
		num = v
		ok = true
	}

	if !ok {
		return errors.Validation("value must be numeric")
	}

	if r.Min != nil && num < *r.Min {
		return errors.Validation(fmt.Sprintf("value must be at least %g", *r.Min))
	}

	if r.Max != nil && num > *r.Max {
		return errors.Validation(fmt.Sprintf("value must be at most %g", *r.Max))
	}

	return nil
}

// Range creates a numeric range validation rule.
func Range(min, max *float64) RangeRule {
	return RangeRule{Min: min, Max: max}
}

// Min creates a minimum value validation rule.
func Min(min float64) RangeRule {
	return RangeRule{Min: &min}
}

// Max creates a maximum value validation rule.
func Max(max float64) RangeRule {
	return RangeRule{Max: &max}
}

// PatternRule validates strings against regular expressions.
type PatternRule struct {
	Pattern *regexp.Regexp
	Message string
}

func (r PatternRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.Validation("value must be a string")
	}

	if !r.Pattern.MatchString(str) {
		if r.Message != "" {
			return errors.Validation(r.Message)
		}
		return errors.Validation("value does not match required pattern")
	}

	return nil
}

// Pattern creates a pattern validation rule.
func Pattern(pattern *regexp.Regexp, message string) PatternRule {
	return PatternRule{Pattern: pattern, Message: message}
}

// MustPattern creates a pattern validation rule with a string pattern.
func MustPattern(pattern, message string) PatternRule {
	regex := regexp.MustCompile(pattern)
	return PatternRule{Pattern: regex, Message: message}
}

// DateRule validates date values and ranges.
type DateRule struct {
	After  *time.Time
	Before *time.Time
}

func (r DateRule) Validate(value interface{}) error {
	var date time.Time
	var ok bool

	switch v := value.(type) {
	case time.Time:
		date = v
		ok = true
	case string:
		var err error
		date, err = time.Parse(time.RFC3339, v)
		if err != nil {
			date, err = time.Parse("2006-01-02", v)
			if err != nil {
				return errors.Validation("invalid date format")
			}
		}
		ok = true
	}

	if !ok {
		return errors.Validation("value must be a date")
	}

	if r.After != nil && date.Before(*r.After) {
		return errors.Validation(fmt.Sprintf("date must be after %s", r.After.Format("2006-01-02")))
	}

	if r.Before != nil && date.After(*r.Before) {
		return errors.Validation(fmt.Sprintf("date must be before %s", r.Before.Format("2006-01-02")))
	}

	return nil
}

// DateRange creates a date range validation rule.
func DateRange(after, before *time.Time) DateRule {
	return DateRule{After: after, Before: before}
}

// DateAfter creates an "after date" validation rule.
func DateAfter(after time.Time) DateRule {
	return DateRule{After: &after}
}

// DateBefore creates a "before date" validation rule.
func DateBefore(before time.Time) DateRule {
	return DateRule{Before: &before}
}

// InRule validates that a value is in a set of allowed values.
type InRule struct {
	Values []interface{}
}

func (r InRule) Validate(value interface{}) error {
	for _, allowedValue := range r.Values {
		if value == allowedValue {
			return nil
		}
	}

	return errors.Validation("value is not in the allowed set")
}

// In creates an "in set" validation rule.
func In(values ...interface{}) InRule {
	return InRule{Values: values}
}

// Custom validation functions

// ValidateStruct validates a struct using field tags or custom rules.
func ValidateStruct(s interface{}) error {
	// This would require reflection to inspect struct tags
	// For now, return nil - implement based on specific requirements
	return nil
}

// ValidatePassword validates password strength.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.Validation("password must be at least 8 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.Validation("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.Validation("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.Validation("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.Validation("password must contain at least one special character")
	}

	return nil
}

// ValidateURL validates URL format.
func ValidateURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return errors.Validation("URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.Validation("URL must start with http:// or https://")
	}

	// Basic URL structure validation
	if len(url) < 10 || !strings.Contains(url, ".") {
		return errors.Validation("invalid URL format")
	}

	return nil
}

// ValidatePostalCode validates postal/zip codes.
func ValidatePostalCode(code, country string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	country = strings.TrimSpace(strings.ToUpper(country))

	if code == "" {
		return errors.Validation("postal code cannot be empty")
	}

	// Country-specific validation patterns
	patterns := map[string]*regexp.Regexp{
		"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),                    // US ZIP codes
		"CA": regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`),          // Canadian postal codes
		"GB": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`), // UK postal codes
		"DE": regexp.MustCompile(`^\d{5}$`),                             // German postal codes
		"FR": regexp.MustCompile(`^\d{5}$`),                             // French postal codes
		"AU": regexp.MustCompile(`^\d{4}$`),                             // Australian postal codes
	}

	if pattern, exists := patterns[country]; exists {
		if !pattern.MatchString(code) {
			return errors.Validation(fmt.Sprintf("invalid postal code format for %s", country))
		}
	} else {
		// Generic validation for unknown countries
		if len(code) < 2 || len(code) > 10 {
			return errors.Validation("postal code must be 2-10 characters long")
		}
	}

	return nil
}

// ValidationBuilder provides a fluent interface for building validators.
type ValidationBuilder struct {
	validator *Validator
}

// NewValidationBuilder creates a new validation builder.
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{
		validator: NewValidator(),
	}
}

// Field starts validation rules for a specific field.
func (b *ValidationBuilder) Field(name string) *FieldValidationBuilder {
	return &FieldValidationBuilder{
		builder: b,
		field:   name,
	}
}

// Build returns the configured validator.
func (b *ValidationBuilder) Build() *Validator {
	return b.validator
}

// FieldValidationBuilder provides a fluent interface for field validation rules.
type FieldValidationBuilder struct {
	builder *ValidationBuilder
	field   string
}

// Required adds a required validation rule.
func (fb *FieldValidationBuilder) Required() *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, Required())
	return fb
}

// Length adds a length validation rule.
func (fb *FieldValidationBuilder) Length(min, max int) *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, Length(min, max))
	return fb
}

// Email adds an email validation rule.
func (fb *FieldValidationBuilder) Email() *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, Email())
	return fb
}

// Phone adds a phone validation rule.
func (fb *FieldValidationBuilder) Phone() *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, Phone())
	return fb
}

// UUID adds a UUID validation rule.
func (fb *FieldValidationBuilder) UUID() *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, UUID())
	return fb
}

// Pattern adds a pattern validation rule.
func (fb *FieldValidationBuilder) Pattern(pattern, message string) *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, MustPattern(pattern, message))
	return fb
}

// Range adds a numeric range validation rule.
func (fb *FieldValidationBuilder) Range(min, max *float64) *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, Range(min, max))
	return fb
}

// In adds an "in set" validation rule.
func (fb *FieldValidationBuilder) In(values ...interface{}) *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, In(values...))
	return fb
}

// Custom adds a custom validation rule.
func (fb *FieldValidationBuilder) Custom(rule ValidationRule) *FieldValidationBuilder {
	fb.builder.validator.AddRule(fb.field, rule)
	return fb
}

// Field ends the current field validation and starts a new one.
func (fb *FieldValidationBuilder) Field(name string) *FieldValidationBuilder {
	return fb.builder.Field(name)
}

// Build returns the configured validator.
func (fb *FieldValidationBuilder) Build() *Validator {
	return fb.builder.Build()
}
