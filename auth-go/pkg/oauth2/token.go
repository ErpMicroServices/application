package oauth2

import (
	"encoding/json"
	"time"
)

// Token represents an OAuth2 access token
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// IsValid checks if the token is still valid (not expired)
func (t *Token) IsValid() bool {
	if t.AccessToken == "" {
		return false
	}
	
	// Add 30 second buffer to account for clock skew and processing time
	return time.Now().Add(30 * time.Second).Before(t.ExpiresAt)
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return !t.IsValid()
}

// ExpiresIn returns the duration until the token expires
func (t *Token) ExpiresIn() time.Duration {
	return time.Until(t.ExpiresAt)
}

// MarshalJSON implements the json.Marshaler interface
func (t *Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(&struct {
		ExpiresAt int64 `json:"expires_at"`
		*Alias
	}{
		ExpiresAt: t.ExpiresAt.Unix(),
		Alias:     (*Alias)(t),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *Token) UnmarshalJSON(data []byte) error {
	type Alias Token
	aux := &struct {
		ExpiresAt int64 `json:"expires_at"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	t.ExpiresAt = time.Unix(aux.ExpiresAt, 0)
	return nil
}

// UserInfo represents user information from the OAuth2 provider
type UserInfo struct {
	Subject           string   `json:"sub"`
	Name             string   `json:"name,omitempty"`
	GivenName        string   `json:"given_name,omitempty"`
	FamilyName       string   `json:"family_name,omitempty"`
	MiddleName       string   `json:"middle_name,omitempty"`
	Nickname         string   `json:"nickname,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Profile          string   `json:"profile,omitempty"`
	Picture          string   `json:"picture,omitempty"`
	Website          string   `json:"website,omitempty"`
	Email            string   `json:"email,omitempty"`
	EmailVerified    bool     `json:"email_verified,omitempty"`
	Gender           string   `json:"gender,omitempty"`
	Birthdate        string   `json:"birthdate,omitempty"`
	ZoneInfo         string   `json:"zoneinfo,omitempty"`
	Locale           string   `json:"locale,omitempty"`
	PhoneNumber      string   `json:"phone_number,omitempty"`
	PhoneVerified    bool     `json:"phone_number_verified,omitempty"`
	Address          *Address `json:"address,omitempty"`
	UpdatedAt        int64    `json:"updated_at,omitempty"`
	
	// ERP-specific fields
	Roles           []string `json:"roles,omitempty"`
	Authorities     []string `json:"authorities,omitempty"`
	OrganizationID  string   `json:"organization_id,omitempty"`
	DepartmentID    string   `json:"department_id,omitempty"`
}

// Address represents a user's address information
type Address struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

// HasRole checks if the user has a specific role
func (u *UserInfo) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAuthority checks if the user has a specific authority
func (u *UserInfo) HasAuthority(authority string) bool {
	for _, a := range u.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *UserInfo) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAnyAuthority checks if the user has any of the specified authorities
func (u *UserInfo) HasAnyAuthority(authorities ...string) bool {
	for _, authority := range authorities {
		if u.HasAuthority(authority) {
			return true
		}
	}
	return false
}

// TokenIntrospection represents the response from a token introspection request
type TokenIntrospection struct {
	Active    bool       `json:"active"`
	Scope     string     `json:"scope,omitempty"`
	ClientID  string     `json:"client_id,omitempty"`
	Username  string     `json:"username,omitempty"`
	TokenType string     `json:"token_type,omitempty"`
	ExpiresAt *time.Time `json:"exp,omitempty"`
	IssuedAt  *time.Time `json:"iat,omitempty"`
	Subject   string     `json:"sub,omitempty"`
	Audience  []string   `json:"aud,omitempty"`
	Issuer    string     `json:"iss,omitempty"`
	
	// ERP-specific fields that might be included
	Roles       []string `json:"roles,omitempty"`
	Authorities []string `json:"authorities,omitempty"`
}

// IsExpired checks if the token introspection indicates the token is expired
func (ti *TokenIntrospection) IsExpired() bool {
	if !ti.Active {
		return true
	}
	
	if ti.ExpiresAt == nil {
		return false
	}
	
	return time.Now().After(*ti.ExpiresAt)
}

// HasRole checks if the introspected token has a specific role
func (ti *TokenIntrospection) HasRole(role string) bool {
	for _, r := range ti.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAuthority checks if the introspected token has a specific authority
func (ti *TokenIntrospection) HasAuthority(authority string) bool {
	for _, a := range ti.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}

// TokenResponse represents a token response from the authorization server
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// ToToken converts a TokenResponse to a Token
func (tr *TokenResponse) ToToken() *Token {
	expiresAt := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	
	return &Token{
		AccessToken:  tr.AccessToken,
		TokenType:    tr.TokenType,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    expiresAt,
		Scope:        tr.Scope,
	}
}

// TokenError represents an OAuth2 token error response
type TokenError struct {
	ErrorType        string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// Error implements the error interface
func (te *TokenError) Error() string {
	if te.ErrorDescription != "" {
		return te.ErrorType + ": " + te.ErrorDescription
	}
	return te.ErrorType
}