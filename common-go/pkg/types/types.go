// Package types provides shared entity definitions and interfaces for the ERP microservices system.
// It includes base entity types, common enums, status types, and business domain interfaces.
package types

import (
	"fmt"
	"time"

	"github.com/erpmicroservices/common-go/pkg/audit"
	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// Entity represents the base interface for all ERP entities.
type Entity interface {
	GetID() uuid.UUID
	SetID(uuid.UUID)
	GetAuditFields() *audit.AuditFields
	SetAuditFields(audit.AuditFields)
}

// BaseEntity provides common fields and methods for all entities.
type BaseEntity struct {
	ID    uuid.UUID         `json:"id" db:"id"`
	Audit audit.AuditFields `json:"audit" db:",inline"`
}

// GetID returns the entity ID.
func (e *BaseEntity) GetID() uuid.UUID {
	return e.ID
}

// SetID sets the entity ID.
func (e *BaseEntity) SetID(id uuid.UUID) {
	e.ID = id
}

// GetAuditFields returns the audit fields.
func (e *BaseEntity) GetAuditFields() *audit.AuditFields {
	return &e.Audit
}

// SetAuditFields sets the audit fields.
func (e *BaseEntity) SetAuditFields(fields audit.AuditFields) {
	e.Audit = fields
}

// NewBaseEntity creates a new base entity with generated ID and audit fields.
func NewBaseEntity(userID uuid.UUID) BaseEntity {
	return BaseEntity{
		ID:    uuid.New(),
		Audit: audit.NewAuditFields(userID),
	}
}

// TypeEntity represents a type/category entity in the ERP system.
type TypeEntity struct {
	BaseEntity
	Code        string `json:"code" db:"code"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IsActive    bool   `json:"isActive" db:"is_active"`
	SortOrder   int    `json:"sortOrder" db:"sort_order"`
}

// NewTypeEntity creates a new type entity.
func NewTypeEntity(code, name, description string, userID uuid.UUID) TypeEntity {
	return TypeEntity{
		BaseEntity:  NewBaseEntity(userID),
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    true,
		SortOrder:   0,
	}
}

// Status represents common status values across the ERP system.
type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusInactive  Status = "INACTIVE"
	StatusPending   Status = "PENDING"
	StatusCancelled Status = "CANCELLED"
	StatusCompleted Status = "COMPLETED"
	StatusDraft     Status = "DRAFT"
	StatusArchived  Status = "ARCHIVED"
)

// IsValid returns true if the status is valid.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusPending, StatusCancelled,
		StatusCompleted, StatusDraft, StatusArchived:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Priority represents priority levels in the ERP system.
type Priority string

const (
	PriorityLow      Priority = "LOW"
	PriorityMedium   Priority = "MEDIUM"
	PriorityHigh     Priority = "HIGH"
	PriorityCritical Priority = "CRITICAL"
)

// IsValid returns true if the priority is valid.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

// String returns the string representation of the priority.
func (p Priority) String() string {
	return string(p)
}

// GetNumericValue returns a numeric representation of the priority for sorting.
func (p Priority) GetNumericValue() int {
	switch p {
	case PriorityLow:
		return 1
	case PriorityMedium:
		return 2
	case PriorityHigh:
		return 3
	case PriorityCritical:
		return 4
	default:
		return 0
	}
}

// Address represents a physical address.
type Address struct {
	Street1    string `json:"street1" db:"street1"`
	Street2    string `json:"street2" db:"street2"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	PostalCode string `json:"postalCode" db:"postal_code"`
	Country    string `json:"country" db:"country"`
}

// IsEmpty returns true if the address is empty.
func (a Address) IsEmpty() bool {
	return a.Street1 == "" && a.City == "" && a.State == "" && a.PostalCode == "" && a.Country == ""
}

// GetFullAddress returns the formatted full address.
func (a Address) GetFullAddress() string {
	parts := []string{}

	if a.Street1 != "" {
		parts = append(parts, a.Street1)
	}
	if a.Street2 != "" {
		parts = append(parts, a.Street2)
	}

	cityStateZip := ""
	if a.City != "" {
		cityStateZip = a.City
	}
	if a.State != "" {
		if cityStateZip != "" {
			cityStateZip += ", " + a.State
		} else {
			cityStateZip = a.State
		}
	}
	if a.PostalCode != "" {
		if cityStateZip != "" {
			cityStateZip += " " + a.PostalCode
		} else {
			cityStateZip = a.PostalCode
		}
	}
	if cityStateZip != "" {
		parts = append(parts, cityStateZip)
	}

	if a.Country != "" {
		parts = append(parts, a.Country)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}

	return result
}

// ContactMethod represents different methods of contact.
type ContactMethod string

const (
	ContactMethodEmail ContactMethod = "EMAIL"
	ContactMethodPhone ContactMethod = "PHONE"
	ContactMethodMail  ContactMethod = "MAIL"
	ContactMethodFax   ContactMethod = "FAX"
	ContactMethodSMS   ContactMethod = "SMS"
	ContactMethodWeb   ContactMethod = "WEB"
)

// IsValid returns true if the contact method is valid.
func (cm ContactMethod) IsValid() bool {
	switch cm {
	case ContactMethodEmail, ContactMethodPhone, ContactMethodMail,
		ContactMethodFax, ContactMethodSMS, ContactMethodWeb:
		return true
	default:
		return false
	}
}

// String returns the string representation of the contact method.
func (cm ContactMethod) String() string {
	return string(cm)
}

// ContactInfo represents contact information.
type ContactInfo struct {
	Method    ContactMethod `json:"method" db:"method"`
	Value     string        `json:"value" db:"value"`
	Label     string        `json:"label" db:"label"`
	IsPrimary bool          `json:"isPrimary" db:"is_primary"`
}

// Money represents a monetary amount with currency.
type Money struct {
	Amount   float64 `json:"amount" db:"amount"`
	Currency string  `json:"currency" db:"currency"`
}

// NewMoney creates a new Money instance.
func NewMoney(amount float64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// String returns a formatted string representation.
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}

// DateRange represents a date range.
type DateRange struct {
	StartDate time.Time  `json:"startDate" db:"start_date"`
	EndDate   *time.Time `json:"endDate" db:"end_date"`
}

// NewDateRange creates a new date range.
func NewDateRange(start time.Time, end *time.Time) DateRange {
	return DateRange{
		StartDate: start,
		EndDate:   end,
	}
}

// IsActive returns true if the date range is currently active.
func (dr DateRange) IsActive() bool {
	now := time.Now()
	if now.Before(dr.StartDate) {
		return false
	}
	if dr.EndDate != nil && now.After(*dr.EndDate) {
		return false
	}
	return true
}

// Contains returns true if the given date falls within the range.
func (dr DateRange) Contains(date time.Time) bool {
	if date.Before(dr.StartDate) {
		return false
	}
	if dr.EndDate != nil && date.After(*dr.EndDate) {
		return false
	}
	return true
}

// Duration returns the duration of the date range.
func (dr DateRange) Duration() time.Duration {
	if dr.EndDate == nil {
		return time.Since(dr.StartDate)
	}
	return dr.EndDate.Sub(dr.StartDate)
}

// Relationship types for entities

// RelationshipType represents different types of relationships between entities.
type RelationshipType string

const (
	RelationshipTypeParent     RelationshipType = "PARENT"
	RelationshipTypeChild      RelationshipType = "CHILD"
	RelationshipTypeSibling    RelationshipType = "SIBLING"
	RelationshipTypeAssociate  RelationshipType = "ASSOCIATE"
	RelationshipTypeDependency RelationshipType = "DEPENDENCY"
)

// IsValid returns true if the relationship type is valid.
func (rt RelationshipType) IsValid() bool {
	switch rt {
	case RelationshipTypeParent, RelationshipTypeChild, RelationshipTypeSibling,
		RelationshipTypeAssociate, RelationshipTypeDependency:
		return true
	default:
		return false
	}
}

// String returns the string representation of the relationship type.
func (rt RelationshipType) String() string {
	return string(rt)
}

// Relationship represents a relationship between two entities.
type Relationship struct {
	FromEntityID     uuid.UUID        `json:"fromEntityId" db:"from_entity_id"`
	ToEntityID       uuid.UUID        `json:"toEntityId" db:"to_entity_id"`
	RelationshipType RelationshipType `json:"relationshipType" db:"relationship_type"`
	DateRange        DateRange        `json:"dateRange" db:",inline"`
	Description      string           `json:"description" db:"description"`
}

// NewRelationship creates a new relationship.
func NewRelationship(from, to uuid.UUID, relType RelationshipType, start time.Time) Relationship {
	return Relationship{
		FromEntityID:     from,
		ToEntityID:       to,
		RelationshipType: relType,
		DateRange:        NewDateRange(start, nil),
	}
}

// IsActive returns true if the relationship is currently active.
func (r Relationship) IsActive() bool {
	return r.DateRange.IsActive()
}

// Interfaces for domain-specific functionality

// Timestampable represents entities that track timestamps.
type Timestampable interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

// Activatable represents entities that can be activated/deactivated.
type Activatable interface {
	IsActive() bool
	SetActive(bool)
	GetStatus() Status
	SetStatus(Status)
}

// Sortable represents entities that can be sorted.
type Sortable interface {
	GetSortOrder() int
	SetSortOrder(int)
}

// Categorizable represents entities that belong to categories.
type Categorizable interface {
	GetCategory() string
	SetCategory(string)
	GetCategoryID() uuid.UUID
	SetCategoryID(uuid.UUID)
}

// Taggable represents entities that can have tags.
type Taggable interface {
	GetTags() []string
	SetTags([]string)
	AddTag(string)
	RemoveTag(string)
	HasTag(string) bool
}

// Addressable represents entities that have addresses.
type Addressable interface {
	GetAddress() Address
	SetAddress(Address)
	GetAddresses() []Address
	AddAddress(Address)
}

// Contactable represents entities that have contact information.
type Contactable interface {
	GetContactInfo() []ContactInfo
	SetContactInfo([]ContactInfo)
	AddContactInfo(ContactInfo)
	GetPrimaryContact(ContactMethod) *ContactInfo
}

// Common enumerations and constants

// Gender represents gender options.
type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderOther   Gender = "OTHER"
	GenderUnknown Gender = "UNKNOWN"
)

// IsValid returns true if the gender is valid.
func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale, GenderOther, GenderUnknown:
		return true
	default:
		return false
	}
}

// MaritalStatus represents marital status options.
type MaritalStatus string

const (
	MaritalStatusSingle    MaritalStatus = "SINGLE"
	MaritalStatusMarried   MaritalStatus = "MARRIED"
	MaritalStatusDivorced  MaritalStatus = "DIVORCED"
	MaritalStatusWidowed   MaritalStatus = "WIDOWED"
	MaritalStatusSeparated MaritalStatus = "SEPARATED"
	MaritalStatusUnknown   MaritalStatus = "UNKNOWN"
)

// IsValid returns true if the marital status is valid.
func (ms MaritalStatus) IsValid() bool {
	switch ms {
	case MaritalStatusSingle, MaritalStatusMarried, MaritalStatusDivorced,
		MaritalStatusWidowed, MaritalStatusSeparated, MaritalStatusUnknown:
		return true
	default:
		return false
	}
}

// Currency represents currency codes.
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyCAD Currency = "CAD"
	CurrencyAUD Currency = "AUD"
	CurrencyCHF Currency = "CHF"
	CurrencyCNY Currency = "CNY"
)

// IsValid returns true if the currency is valid.
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyJPY,
		CurrencyCAD, CurrencyAUD, CurrencyCHF, CurrencyCNY:
		return true
	default:
		return false
	}
}

// GetSymbol returns the currency symbol.
func (c Currency) GetSymbol() string {
	switch c {
	case CurrencyUSD, CurrencyCAD, CurrencyAUD:
		return "$"
	case CurrencyEUR:
		return "€"
	case CurrencyGBP:
		return "£"
	case CurrencyJPY, CurrencyCNY:
		return "¥"
	case CurrencyCHF:
		return "₣"
	default:
		return string(c)
	}
}

// TimeZone represents time zone identifiers.
type TimeZone string

const (
	TimeZoneUTC  TimeZone = "UTC"
	TimeZoneEST  TimeZone = "America/New_York"
	TimeZoneCST  TimeZone = "America/Chicago"
	TimeZoneMST  TimeZone = "America/Denver"
	TimeZonePST  TimeZone = "America/Los_Angeles"
	TimeZoneCET  TimeZone = "Europe/Paris"
	TimeZoneGMT  TimeZone = "Europe/London"
	TimeZoneJST  TimeZone = "Asia/Tokyo"
	TimeZoneAEST TimeZone = "Australia/Sydney"
)

// Language represents language codes.
type Language string

const (
	LanguageEnglish    Language = "en"
	LanguageSpanish    Language = "es"
	LanguageFrench     Language = "fr"
	LanguageGerman     Language = "de"
	LanguageItalian    Language = "it"
	LanguagePortuguese Language = "pt"
	LanguageChinese    Language = "zh"
	LanguageJapanese   Language = "ja"
	LanguageRussian    Language = "ru"
	LanguageArabic     Language = "ar"
)

// IsValid returns true if the language code is valid.
func (l Language) IsValid() bool {
	switch l {
	case LanguageEnglish, LanguageSpanish, LanguageFrench, LanguageGerman,
		LanguageItalian, LanguagePortuguese, LanguageChinese, LanguageJapanese,
		LanguageRussian, LanguageArabic:
		return true
	default:
		return false
	}
}

// Country represents country codes (ISO 3166-1 alpha-2).
type Country string

const (
	CountryUS Country = "US"
	CountryCA Country = "CA"
	CountryGB Country = "GB"
	CountryDE Country = "DE"
	CountryFR Country = "FR"
	CountryIT Country = "IT"
	CountryES Country = "ES"
	CountryAU Country = "AU"
	CountryJP Country = "JP"
	CountryCN Country = "CN"
)

// IsValid returns true if the country code is valid.
func (c Country) IsValid() bool {
	switch c {
	case CountryUS, CountryCA, CountryGB, CountryDE, CountryFR,
		CountryIT, CountryES, CountryAU, CountryJP, CountryCN:
		return true
	default:
		return false
	}
}
