// Package audit provides standardized audit fields and utilities for tracking
// entity creation and modification in the ERP microservices system.
package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// AuditFields contains the standard audit information that should be included
// in all entities within the ERP system. These fields track when and by whom
// an entity was created and last modified.
type AuditFields struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	CreatedBy uuid.UUID `json:"createdBy" db:"created_by"`
	UpdatedBy uuid.UUID `json:"updatedBy" db:"updated_by"`
}

// AuditInfo represents audit information extracted from context or request.
type AuditInfo struct {
	UserID        uuid.UUID
	CorrelationID string
	Timestamp     time.Time
	IPAddress     string
	UserAgent     string
}

// Auditable interface should be implemented by entities that support audit fields.
type Auditable interface {
	GetAuditFields() *AuditFields
	SetAuditFields(fields AuditFields)
}

// NewAuditFields creates new audit fields with the current timestamp and user ID.
func NewAuditFields(userID uuid.UUID) AuditFields {
	now := time.Now().UTC()
	return AuditFields{
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
}

// NewAuditFieldsWithTime creates new audit fields with the specified timestamp and user ID.
func NewAuditFieldsWithTime(userID uuid.UUID, timestamp time.Time) AuditFields {
	return AuditFields{
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
}

// UpdateAuditFields updates the UpdatedAt and UpdatedBy fields while preserving
// the original creation information.
func (af *AuditFields) UpdateAuditFields(userID uuid.UUID) {
	af.UpdatedAt = time.Now().UTC()
	af.UpdatedBy = userID
}

// UpdateAuditFieldsWithTime updates the UpdatedAt and UpdatedBy fields with
// the specified timestamp while preserving the original creation information.
func (af *AuditFields) UpdateAuditFieldsWithTime(userID uuid.UUID, timestamp time.Time) {
	af.UpdatedAt = timestamp
	af.UpdatedBy = userID
}

// IsValid checks if the audit fields contain valid data.
func (af AuditFields) IsValid() bool {
	return !af.CreatedAt.IsZero() &&
		!af.UpdatedAt.IsZero() &&
		!af.CreatedBy.IsNil() &&
		!af.UpdatedBy.IsNil() &&
		!af.UpdatedAt.Before(af.CreatedAt)
}

// GetAgeInDuration returns how long ago the entity was created.
func (af AuditFields) GetAgeInDuration() time.Duration {
	return time.Since(af.CreatedAt)
}

// GetLastModifiedDuration returns how long ago the entity was last modified.
func (af AuditFields) GetLastModifiedDuration() time.Duration {
	return time.Since(af.UpdatedAt)
}

// IsModified returns true if the entity has been modified since creation.
func (af AuditFields) IsModified() bool {
	return !af.CreatedAt.Equal(af.UpdatedAt) || !af.CreatedBy.Equal(af.UpdatedBy)
}

// Context keys for audit information
type contextKey string

const (
	UserIDKey        contextKey = "audit_user_id"
	CorrelationIDKey contextKey = "audit_correlation_id"
	IPAddressKey     contextKey = "audit_ip_address"
	UserAgentKey     contextKey = "audit_user_agent"
)

// WithUserID adds user ID to context for audit purposes.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithCorrelationID adds correlation ID to context for audit tracking.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithIPAddress adds IP address to context for audit logging.
func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ipAddress)
}

// WithUserAgent adds user agent to context for audit logging.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

// GetUserIDFromContext extracts user ID from context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetCorrelationIDFromContext extracts correlation ID from context.
func GetCorrelationIDFromContext(ctx context.Context) (string, bool) {
	correlationID, ok := ctx.Value(CorrelationIDKey).(string)
	return correlationID, ok
}

// GetIPAddressFromContext extracts IP address from context.
func GetIPAddressFromContext(ctx context.Context) (string, bool) {
	ipAddress, ok := ctx.Value(IPAddressKey).(string)
	return ipAddress, ok
}

// GetUserAgentFromContext extracts user agent from context.
func GetUserAgentFromContext(ctx context.Context) (string, bool) {
	userAgent, ok := ctx.Value(UserAgentKey).(string)
	return userAgent, ok
}

// GetAuditInfoFromContext extracts all audit information from context.
func GetAuditInfoFromContext(ctx context.Context) AuditInfo {
	info := AuditInfo{
		Timestamp: time.Now().UTC(),
	}

	if userID, ok := GetUserIDFromContext(ctx); ok {
		info.UserID = userID
	}

	if correlationID, ok := GetCorrelationIDFromContext(ctx); ok {
		info.CorrelationID = correlationID
	}

	if ipAddress, ok := GetIPAddressFromContext(ctx); ok {
		info.IPAddress = ipAddress
	}

	if userAgent, ok := GetUserAgentFromContext(ctx); ok {
		info.UserAgent = userAgent
	}

	return info
}

// ApplyAuditFields applies audit fields to an entity based on context information.
// For new entities, it sets both creation and update fields.
// For existing entities, it only updates the modification fields.
func ApplyAuditFields(ctx context.Context, entity Auditable, isNew bool) error {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok || userID.IsNil() {
		userID = uuid.Nil() // Allow system operations with nil user
	}

	auditFields := entity.GetAuditFields()
	if auditFields == nil {
		return fmt.Errorf("entity does not have audit fields")
	}

	if isNew {
		*auditFields = NewAuditFields(userID)
	} else {
		auditFields.UpdateAuditFields(userID)
	}

	return nil
}

// AuditEvent represents an audit event for logging purposes.
type AuditEvent struct {
	EntityType    string                 `json:"entityType"`
	EntityID      uuid.UUID              `json:"entityId"`
	Action        string                 `json:"action"`
	UserID        uuid.UUID              `json:"userId"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	IPAddress     string                 `json:"ipAddress,omitempty"`
	UserAgent     string                 `json:"userAgent,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Changes       map[string]Change      `json:"changes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Change represents a field change in an audit event.
type Change struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// AuditAction constants for common audit actions.
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionRead   = "READ"
)

// NewAuditEvent creates a new audit event.
func NewAuditEvent(entityType string, entityID uuid.UUID, action string, userID uuid.UUID) AuditEvent {
	return AuditEvent{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		UserID:     userID,
		Timestamp:  time.Now().UTC(),
		Changes:    make(map[string]Change),
		Metadata:   make(map[string]interface{}),
	}
}

// AddChange adds a field change to the audit event.
func (ae *AuditEvent) AddChange(field string, oldValue, newValue interface{}) {
	ae.Changes[field] = Change{
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
	}
}

// AddMetadata adds metadata to the audit event.
func (ae *AuditEvent) AddMetadata(key string, value interface{}) {
	ae.Metadata[key] = value
}

// WithAuditInfo enriches the audit event with audit info from context.
func (ae *AuditEvent) WithAuditInfo(info AuditInfo) *AuditEvent {
	ae.UserID = info.UserID
	ae.CorrelationID = info.CorrelationID
	ae.IPAddress = info.IPAddress
	ae.UserAgent = info.UserAgent
	ae.Timestamp = info.Timestamp
	return ae
}
