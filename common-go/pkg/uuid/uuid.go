// Package uuid provides UUID utilities for the ERP microservices system.
// It wraps the google/uuid package with additional validation and conversion helpers.
package uuid

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UUID represents a UUID value that can be used throughout the ERP system.
// It provides additional validation and serialization methods beyond the standard
// google/uuid.UUID type.
type UUID struct {
	uuid.UUID
}

// New generates a new UUID v4.
// This is the primary method for creating new UUIDs in the ERP system.
func New() UUID {
	return UUID{UUID: uuid.New()}
}

// NewFromString parses a UUID from a string and returns it.
// Returns an error if the string is not a valid UUID format.
func NewFromString(s string) (UUID, error) {
	if s == "" {
		return UUID{}, fmt.Errorf("uuid: empty string")
	}

	parsed, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return UUID{}, fmt.Errorf("uuid: invalid format %q: %w", s, err)
	}

	return UUID{UUID: parsed}, nil
}

// MustParse is like NewFromString but panics if the string is invalid.
// Use this only when you're certain the input is valid, such as in tests or
// with hardcoded values.
func MustParse(s string) UUID {
	u, err := NewFromString(s)
	if err != nil {
		panic(err)
	}
	return u
}

// IsValid checks if a string is a valid UUID format.
func IsValid(s string) bool {
	_, err := uuid.Parse(strings.TrimSpace(s))
	return err == nil
}

// IsNil returns true if the UUID is the nil UUID (all zeros).
func (u UUID) IsNil() bool {
	return u.UUID == uuid.Nil
}

// String returns the string representation of the UUID.
// This implements the fmt.Stringer interface.
func (u UUID) String() string {
	if u.IsNil() {
		return ""
	}
	return u.UUID.String()
}

// MarshalJSON implements json.Marshaler interface.
func (u UUID) MarshalJSON() ([]byte, error) {
	if u.IsNil() {
		return []byte("null"), nil
	}
	return []byte(`"` + u.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (u *UUID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		*u = UUID{}
		return nil
	}

	// Remove quotes if present
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := NewFromString(s)
	if err != nil {
		return fmt.Errorf("uuid: unmarshal JSON %q: %w", string(data), err)
	}

	*u = parsed
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (u UUID) Value() (driver.Value, error) {
	if u.IsNil() {
		return nil, nil
	}
	return u.String(), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = UUID{}
		return nil
	}

	switch v := value.(type) {
	case string:
		parsed, err := NewFromString(v)
		if err != nil {
			return fmt.Errorf("uuid: scan string %q: %w", v, err)
		}
		*u = parsed
	case []byte:
		parsed, err := NewFromString(string(v))
		if err != nil {
			return fmt.Errorf("uuid: scan bytes %q: %w", v, err)
		}
		*u = parsed
	default:
		return fmt.Errorf("uuid: cannot scan %T into UUID", value)
	}

	return nil
}

// Nil returns the nil UUID (all zeros).
func Nil() UUID {
	return UUID{UUID: uuid.Nil}
}

// Equal returns true if two UUIDs are equal.
func (u UUID) Equal(other UUID) bool {
	return u.UUID == other.UUID
}

// Version returns the version of the UUID.
func (u UUID) Version() uuid.Version {
	return u.UUID.Version()
}

// Variant returns the variant of the UUID.
func (u UUID) Variant() uuid.Variant {
	return u.UUID.Variant()
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (u UUID) MarshalBinary() ([]byte, error) {
	return u.UUID.MarshalBinary()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (u *UUID) UnmarshalBinary(data []byte) error {
	return u.UUID.UnmarshalBinary(data)
}

// MarshalText implements encoding.TextMarshaler interface.
func (u UUID) MarshalText() ([]byte, error) {
	if u.IsNil() {
		return nil, nil
	}
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (u *UUID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*u = UUID{}
		return nil
	}

	parsed, err := NewFromString(string(text))
	if err != nil {
		return fmt.Errorf("uuid: unmarshal text %q: %w", text, err)
	}

	*u = parsed
	return nil
}

// NewSlice creates a slice of UUIDs from a slice of strings.
// Invalid UUIDs will cause an error to be returned.
func NewSlice(strs []string) ([]UUID, error) {
	if len(strs) == 0 {
		return nil, nil
	}

	uuids := make([]UUID, len(strs))
	for i, s := range strs {
		u, err := NewFromString(s)
		if err != nil {
			return nil, fmt.Errorf("uuid: invalid UUID at index %d: %w", i, err)
		}
		uuids[i] = u
	}

	return uuids, nil
}

// StringSlice converts a slice of UUIDs to a slice of strings.
func StringSlice(uuids []UUID) []string {
	if len(uuids) == 0 {
		return nil
	}

	strs := make([]string, len(uuids))
	for i, u := range uuids {
		strs[i] = u.String()
	}

	return strs
}
