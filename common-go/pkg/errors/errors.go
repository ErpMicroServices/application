// Package errors provides structured error handling with context and categorization
// for the ERP microservices system. It includes custom error types, error wrapping,
// and integration with GraphQL error handling.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/erpmicroservices/common-go/pkg/uuid"
	"go.uber.org/multierr"
)

// ErrorCode represents a standardized error code for the ERP system.
type ErrorCode string

const (
	// Generic error codes
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeTimeout      ErrorCode = "TIMEOUT"

	// Business logic error codes
	CodeBusinessRule     ErrorCode = "BUSINESS_RULE_VIOLATION"
	CodeInvalidState     ErrorCode = "INVALID_STATE"
	CodeInsufficientData ErrorCode = "INSUFFICIENT_DATA"
	CodeDuplicateEntry   ErrorCode = "DUPLICATE_ENTRY"

	// External service error codes
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeNetworkError       ErrorCode = "NETWORK_ERROR"
	CodeRateLimited        ErrorCode = "RATE_LIMITED"
)

// Severity represents the severity level of an error.
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ErrorCategory represents the category of an error for classification.
type ErrorCategory string

const (
	CategoryInput    ErrorCategory = "INPUT"
	CategoryBusiness ErrorCategory = "BUSINESS"
	CategorySystem   ErrorCategory = "SYSTEM"
	CategoryExternal ErrorCategory = "EXTERNAL"
	CategorySecurity ErrorCategory = "SECURITY"
)

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// ERPError represents a structured error with context and metadata.
type ERPError struct {
	ID            uuid.UUID              `json:"id"`
	Code          ErrorCode              `json:"code"`
	Message       string                 `json:"message"`
	Details       string                 `json:"details,omitempty"`
	Category      ErrorCategory          `json:"category"`
	Severity      Severity               `json:"severity"`
	Timestamp     time.Time              `json:"timestamp"`
	HTTPStatus    int                    `json:"httpStatus,omitempty"`
	UserMessage   string                 `json:"userMessage,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	Stack         []StackFrame           `json:"stack,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Cause         error                  `json:"cause,omitempty"`
}

// Error implements the error interface.
func (e *ERPError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *ERPError) Unwrap() error {
	return e.Cause
}

// WithCause adds a cause to the error.
func (e *ERPError) WithCause(cause error) *ERPError {
	e.Cause = cause
	return e
}

// WithDetails adds additional details to the error.
func (e *ERPError) WithDetails(details string) *ERPError {
	e.Details = details
	return e
}

// WithUserMessage adds a user-friendly message to the error.
func (e *ERPError) WithUserMessage(message string) *ERPError {
	e.UserMessage = message
	return e
}

// WithCorrelationID adds a correlation ID to the error.
func (e *ERPError) WithCorrelationID(id string) *ERPError {
	e.CorrelationID = id
	return e
}

// WithMetadata adds metadata to the error.
func (e *ERPError) WithMetadata(key string, value interface{}) *ERPError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithStack captures the current stack trace.
func (e *ERPError) WithStack() *ERPError {
	e.Stack = captureStack(2) // Skip this method and the caller
	return e
}

// IsTemporary returns true if the error is temporary and might succeed on retry.
func (e *ERPError) IsTemporary() bool {
	return e.Code == CodeTimeout ||
		e.Code == CodeServiceUnavailable ||
		e.Code == CodeNetworkError ||
		e.Code == CodeRateLimited
}

// IsRetryable returns true if the error indicates a retryable operation.
func (e *ERPError) IsRetryable() bool {
	return e.IsTemporary()
}

// GetHTTPStatus returns the appropriate HTTP status code for the error.
func (e *ERPError) GetHTTPStatus() int {
	if e.HTTPStatus != 0 {
		return e.HTTPStatus
	}

	switch e.Code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeConflict, CodeDuplicateEntry:
		return http.StatusConflict
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeBusinessRule, CodeInvalidState:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// MarshalJSON implements json.Marshaler for custom JSON serialization.
func (e *ERPError) MarshalJSON() ([]byte, error) {
	type Alias ERPError
	return json.Marshal(&struct {
		*Alias
		CauseMessage string `json:"causeMessage,omitempty"`
	}{
		Alias:        (*Alias)(e),
		CauseMessage: formatCause(e.Cause),
	})
}

// captureStack captures the current stack trace.
func captureStack(skip int) []StackFrame {
	const maxFrames = 50
	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip, pcs)
	frames := make([]StackFrame, 0, n)

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			File:     file,
			Function: fn.Name(),
			Line:     line,
		})
	}

	return frames
}

// formatCause formats the cause error for JSON serialization.
func formatCause(cause error) string {
	if cause == nil {
		return ""
	}
	return cause.Error()
}

// NewERPError creates a new ERP error with the specified parameters.
func NewERPError(code ErrorCode, message string, category ErrorCategory, severity Severity) *ERPError {
	return &ERPError{
		ID:        uuid.New(),
		Code:      code,
		Message:   message,
		Category:  category,
		Severity:  severity,
		Timestamp: time.Now().UTC(),
		Metadata:  make(map[string]interface{}),
	}
}

// Validation creates a validation error.
func Validation(message string) *ERPError {
	return NewERPError(CodeValidation, message, CategoryInput, SeverityMedium)
}

// ValidationWithField creates a validation error for a specific field.
func ValidationWithField(field, message string) *ERPError {
	return Validation(message).WithMetadata("field", field)
}

// NotFound creates a not found error.
func NotFound(resource string) *ERPError {
	return NewERPError(CodeNotFound, fmt.Sprintf("%s not found", resource), CategoryInput, SeverityLow)
}

// NotFoundWithID creates a not found error with a specific ID.
func NotFoundWithID(resource string, id uuid.UUID) *ERPError {
	return NotFound(resource).WithMetadata("id", id.String())
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *ERPError {
	return NewERPError(CodeUnauthorized, message, CategorySecurity, SeverityHigh)
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *ERPError {
	return NewERPError(CodeForbidden, message, CategorySecurity, SeverityHigh)
}

// BusinessRule creates a business rule violation error.
func BusinessRule(message string) *ERPError {
	return NewERPError(CodeBusinessRule, message, CategoryBusiness, SeverityMedium)
}

// InvalidState creates an invalid state error.
func InvalidState(message string) *ERPError {
	return NewERPError(CodeInvalidState, message, CategoryBusiness, SeverityMedium)
}

// Internal creates an internal server error.
func Internal(message string) *ERPError {
	return NewERPError(CodeInternal, message, CategorySystem, SeverityCritical).WithStack()
}

// InternalWithCause creates an internal server error with a cause.
func InternalWithCause(message string, cause error) *ERPError {
	return Internal(message).WithCause(cause)
}

// Conflict creates a conflict error.
func Conflict(message string) *ERPError {
	return NewERPError(CodeConflict, message, CategoryBusiness, SeverityMedium)
}

// DuplicateEntry creates a duplicate entry error.
func DuplicateEntry(resource string) *ERPError {
	return NewERPError(CodeDuplicateEntry, fmt.Sprintf("Duplicate %s", resource), CategoryBusiness, SeverityMedium)
}

// ServiceUnavailable creates a service unavailable error.
func ServiceUnavailable(service string) *ERPError {
	return NewERPError(CodeServiceUnavailable, fmt.Sprintf("Service %s is unavailable", service), CategoryExternal, SeverityHigh)
}

// DatabaseError creates a database error.
func DatabaseError(operation string, cause error) *ERPError {
	return NewERPError(CodeDatabaseError, fmt.Sprintf("Database %s failed", operation), CategorySystem, SeverityHigh).
		WithCause(cause).
		WithStack()
}

// NetworkError creates a network error.
func NetworkError(message string, cause error) *ERPError {
	return NewERPError(CodeNetworkError, message, CategoryExternal, SeverityMedium).WithCause(cause)
}

// Timeout creates a timeout error.
func Timeout(operation string) *ERPError {
	return NewERPError(CodeTimeout, fmt.Sprintf("Operation %s timed out", operation), CategorySystem, SeverityMedium)
}

// RateLimited creates a rate limited error.
func RateLimited(message string) *ERPError {
	return NewERPError(CodeRateLimited, message, CategoryExternal, SeverityLow)
}

// ErrorList represents a list of errors that can be accumulated.
type ErrorList struct {
	errors []error
}

// NewErrorList creates a new error list.
func NewErrorList() *ErrorList {
	return &ErrorList{
		errors: make([]error, 0),
	}
}

// Add adds an error to the list.
func (el *ErrorList) Add(err error) {
	if err != nil {
		el.errors = append(el.errors, err)
	}
}

// AddValidation adds a validation error to the list.
func (el *ErrorList) AddValidation(field, message string) {
	el.Add(ValidationWithField(field, message))
}

// HasErrors returns true if the list contains any errors.
func (el *ErrorList) HasErrors() bool {
	return len(el.errors) > 0
}

// Count returns the number of errors in the list.
func (el *ErrorList) Count() int {
	return len(el.errors)
}

// ToError converts the error list to a single error using multierr.
func (el *ErrorList) ToError() error {
	if len(el.errors) == 0 {
		return nil
	}
	return multierr.Combine(el.errors...)
}

// ToERPError converts the error list to a single ERP error.
func (el *ErrorList) ToERPError() *ERPError {
	if len(el.errors) == 0 {
		return nil
	}

	if len(el.errors) == 1 {
		if erpErr, ok := el.errors[0].(*ERPError); ok {
			return erpErr
		}
	}

	messages := make([]string, len(el.errors))
	for i, err := range el.errors {
		messages[i] = err.Error()
	}

	return Validation("Multiple validation errors").
		WithDetails(strings.Join(messages, "; ")).
		WithMetadata("errorCount", len(el.errors))
}

// GetErrors returns all errors in the list.
func (el *ErrorList) GetErrors() []error {
	return el.errors
}

// Clear removes all errors from the list.
func (el *ErrorList) Clear() {
	el.errors = el.errors[:0]
}

// IsERPError checks if an error is an ERP error.
func IsERPError(err error) bool {
	_, ok := err.(*ERPError)
	return ok
}

// AsERPError converts an error to an ERP error if possible.
func AsERPError(err error) (*ERPError, bool) {
	erpErr, ok := err.(*ERPError)
	return erpErr, ok
}

// WrapIfNotERP wraps an error as an internal ERP error if it's not already an ERP error.
func WrapIfNotERP(err error, message string) *ERPError {
	if erpErr, ok := AsERPError(err); ok {
		return erpErr
	}
	return InternalWithCause(message, err)
}

// ErrorHandler provides methods for handling and converting errors.
type ErrorHandler struct {
	defaultSeverity Severity
	includeStack    bool
}

// NewErrorHandler creates a new error handler with default configuration.
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		defaultSeverity: SeverityMedium,
		includeStack:    true,
	}
}

// WithDefaultSeverity sets the default severity for wrapped errors.
func (h *ErrorHandler) WithDefaultSeverity(severity Severity) *ErrorHandler {
	h.defaultSeverity = severity
	return h
}

// WithStackTrace enables or disables stack trace capture.
func (h *ErrorHandler) WithStackTrace(include bool) *ErrorHandler {
	h.includeStack = include
	return h
}

// Handle processes an error and returns an appropriate ERP error.
func (h *ErrorHandler) Handle(err error, message string) *ERPError {
	if err == nil {
		return nil
	}

	if erpErr, ok := AsERPError(err); ok {
		return erpErr
	}

	erpErr := NewERPError(CodeInternal, message, CategorySystem, h.defaultSeverity).WithCause(err)

	if h.includeStack {
		erpErr.WithStack()
	}

	return erpErr
}
