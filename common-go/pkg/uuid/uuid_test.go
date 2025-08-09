package uuid

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	u := New()
	if u.IsNil() {
		t.Error("New() should not return nil UUID")
	}
	if u.Version() != 4 {
		t.Errorf("New() should return UUID v4, got v%d", u.Version())
	}
}

func TestNewFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			input:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID with whitespace",
			input:   "  550e8400-e29b-41d4-a716-446655440000  ",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "nil UUID string",
			input:   "00000000-0000-0000-0000-000000000000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u.IsNil() && tt.input != "00000000-0000-0000-0000-000000000000" {
				t.Error("NewFromString() returned nil UUID for valid input")
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	u := MustParse(validUUID)
	if u.String() != validUUID {
		t.Errorf("MustParse() = %v, want %v", u.String(), validUUID)
	}

	// Test panic behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse() should panic for invalid UUID")
		}
	}()
	MustParse("invalid-uuid")
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"  550e8400-e29b-41d4-a716-446655440000  ", true},
		{"", false},
		{"invalid-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
	}

	for _, tt := range tests {
		if got := IsValid(tt.input); got != tt.want {
			t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestUUID_IsNil(t *testing.T) {
	nilUUID := Nil()
	if !nilUUID.IsNil() {
		t.Error("Nil UUID should report IsNil() = true")
	}

	nonNilUUID := New()
	if nonNilUUID.IsNil() {
		t.Error("Non-nil UUID should report IsNil() = false")
	}
}

func TestUUID_String(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
		want string
	}{
		{
			name: "nil UUID",
			uuid: Nil(),
			want: "",
		},
		{
			name: "valid UUID",
			uuid: MustParse("550e8400-e29b-41d4-a716-446655440000"),
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uuid.String(); got != tt.want {
				t.Errorf("UUID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUUID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
		want string
	}{
		{
			name: "nil UUID",
			uuid: Nil(),
			want: "null",
		},
		{
			name: "valid UUID",
			uuid: MustParse("550e8400-e29b-41d4-a716-446655440000"),
			want: `"550e8400-e29b-41d4-a716-446655440000"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.uuid.MarshalJSON()
			if err != nil {
				t.Errorf("UUID.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("UUID.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestUUID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    UUID
		wantErr bool
	}{
		{
			name: "null value",
			data: "null",
			want: Nil(),
		},
		{
			name: "empty string",
			data: `""`,
			want: Nil(),
		},
		{
			name: "valid UUID",
			data: `"550e8400-e29b-41d4-a716-446655440000"`,
			want: MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:    "invalid UUID",
			data:    `"invalid-uuid"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UUID
			err := u.UnmarshalJSON([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("UUID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u != tt.want {
				t.Errorf("UUID.UnmarshalJSON() = %v, want %v", u, tt.want)
			}
		})
	}
}

func TestUUID_Value(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
		want driver.Value
	}{
		{
			name: "nil UUID",
			uuid: Nil(),
			want: nil,
		},
		{
			name: "valid UUID",
			uuid: MustParse("550e8400-e29b-41d4-a716-446655440000"),
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.uuid.Value()
			if err != nil {
				t.Errorf("UUID.Value() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UUID.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUUID_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    UUID
		wantErr bool
	}{
		{
			name:  "nil value",
			value: nil,
			want:  Nil(),
		},
		{
			name:  "string value",
			value: "550e8400-e29b-41d4-a716-446655440000",
			want:  MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:  "byte slice value",
			value: []byte("550e8400-e29b-41d4-a716-446655440000"),
			want:  MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:    "invalid string",
			value:   "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UUID
			err := u.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("UUID.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u != tt.want {
				t.Errorf("UUID.Scan() = %v, want %v", u, tt.want)
			}
		})
	}
}

func TestUUID_Equal(t *testing.T) {
	u1 := MustParse("550e8400-e29b-41d4-a716-446655440000")
	u2 := MustParse("550e8400-e29b-41d4-a716-446655440000")
	u3 := New()

	if !u1.Equal(u2) {
		t.Error("Equal UUIDs should be equal")
	}
	if u1.Equal(u3) {
		t.Error("Different UUIDs should not be equal")
	}
}

func TestNewSlice(t *testing.T) {
	tests := []struct {
		name    string
		strs    []string
		wantErr bool
		wantLen int
	}{
		{
			name:    "empty slice",
			strs:    []string{},
			wantLen: 0,
		},
		{
			name:    "nil slice",
			strs:    nil,
			wantLen: 0,
		},
		{
			name:    "valid UUIDs",
			strs:    []string{"550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
			wantLen: 2,
		},
		{
			name:    "invalid UUID in slice",
			strs:    []string{"550e8400-e29b-41d4-a716-446655440000", "invalid-uuid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSlice(tt.strs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("NewSlice() length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		uuids []UUID
		want  int
	}{
		{
			name:  "empty slice",
			uuids: []UUID{},
			want:  0,
		},
		{
			name:  "nil slice",
			uuids: nil,
			want:  0,
		},
		{
			name:  "valid UUIDs",
			uuids: []UUID{New(), New()},
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringSlice(tt.uuids)
			if len(got) != tt.want {
				t.Errorf("StringSlice() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestJSON_RoundTrip(t *testing.T) {
	original := New()

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled UUID
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !original.Equal(unmarshaled) {
		t.Errorf("JSON round trip failed: original = %v, unmarshaled = %v", original, unmarshaled)
	}
}
