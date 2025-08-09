package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for the ERP microservices system
type Claims struct {
	jwt.RegisteredClaims
	
	// User Information
	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	
	// Authorization
	Roles       []string `json:"roles,omitempty"`
	Authorities []string `json:"authorities,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	
	// ERP-specific claims
	OrganizationID     string                 `json:"organization_id,omitempty"`
	DepartmentID       string                 `json:"department_id,omitempty"`
	EmployeeID         string                 `json:"employee_id,omitempty"`
	TenantID           string                 `json:"tenant_id,omitempty"`
	
	// Session Information
	SessionID          string                 `json:"session_id,omitempty"`
	ClientID           string                 `json:"client_id,omitempty"`
	TokenType          string                 `json:"token_type,omitempty"`
	
	// Additional custom claims
	CustomClaims       map[string]interface{} `json:"custom_claims,omitempty"`
}

// NewClaims creates a new Claims instance with default values
func NewClaims() *Claims {
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Roles:           []string{},
		Authorities:     []string{},
		Scopes:          []string{},
		CustomClaims:    make(map[string]interface{}),
	}
}

// NewUserClaims creates claims for a user token
func NewUserClaims(subject, email, name string, roles, authorities []string) *Claims {
	claims := NewClaims()
	claims.Subject = subject
	claims.Email = email
	claims.Name = name
	claims.Roles = roles
	claims.Authorities = authorities
	return claims
}

// NewServiceClaims creates claims for a service-to-service token
func NewServiceClaims(clientID string, scopes []string) *Claims {
	claims := NewClaims()
	claims.Subject = clientID
	claims.ClientID = clientID
	claims.TokenType = "client_credentials"
	claims.Scopes = scopes
	claims.Authorities = []string{"SERVICE"}
	return claims
}

// HasRole checks if the token contains a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the token contains any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the token contains all of the specified roles
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// HasAuthority checks if the token contains a specific authority
func (c *Claims) HasAuthority(authority string) bool {
	for _, a := range c.Authorities {
		if a == authority {
			return true
		}
	}
	return false
}

// HasAnyAuthority checks if the token contains any of the specified authorities
func (c *Claims) HasAnyAuthority(authorities ...string) bool {
	for _, authority := range authorities {
		if c.HasAuthority(authority) {
			return true
		}
	}
	return false
}

// HasAllAuthorities checks if the token contains all of the specified authorities
func (c *Claims) HasAllAuthorities(authorities ...string) bool {
	for _, authority := range authorities {
		if !c.HasAuthority(authority) {
			return false
		}
	}
	return true
}

// HasScope checks if the token contains a specific scope
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the token contains any of the specified scopes
func (c *Claims) HasAnyScope(scopes ...string) bool {
	for _, scope := range scopes {
		if c.HasScope(scope) {
			return true
		}
	}
	return false
}

// IsServiceToken returns true if this is a service-to-service token
func (c *Claims) IsServiceToken() bool {
	return c.TokenType == "client_credentials" || c.HasAuthority("SERVICE")
}

// IsUserToken returns true if this is a user token
func (c *Claims) IsUserToken() bool {
	return !c.IsServiceToken() && c.Subject != c.ClientID
}

// GetDisplayName returns a suitable display name for the user
func (c *Claims) GetDisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	if c.PreferredUsername != "" {
		return c.PreferredUsername
	}
	if c.Email != "" {
		return c.Email
	}
	return c.Subject
}

// GetFullName returns the full name of the user
func (c *Claims) GetFullName() string {
	if c.Name != "" {
		return c.Name
	}
	if c.GivenName != "" || c.FamilyName != "" {
		return c.GivenName + " " + c.FamilyName
	}
	return c.GetDisplayName()
}

// SetCustomClaim sets a custom claim value
func (c *Claims) SetCustomClaim(key string, value interface{}) {
	if c.CustomClaims == nil {
		c.CustomClaims = make(map[string]interface{})
	}
	c.CustomClaims[key] = value
}

// GetCustomClaim gets a custom claim value
func (c *Claims) GetCustomClaim(key string) (interface{}, bool) {
	if c.CustomClaims == nil {
		return nil, false
	}
	value, exists := c.CustomClaims[key]
	return value, exists
}

// GetCustomClaimString gets a custom claim as a string
func (c *Claims) GetCustomClaimString(key string) (string, bool) {
	if value, exists := c.GetCustomClaim(key); exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// AddRole adds a role to the token claims
func (c *Claims) AddRole(role string) {
	if !c.HasRole(role) {
		c.Roles = append(c.Roles, role)
	}
}

// RemoveRole removes a role from the token claims
func (c *Claims) RemoveRole(role string) {
	for i, r := range c.Roles {
		if r == role {
			c.Roles = append(c.Roles[:i], c.Roles[i+1:]...)
			return
		}
	}
}

// AddAuthority adds an authority to the token claims
func (c *Claims) AddAuthority(authority string) {
	if !c.HasAuthority(authority) {
		c.Authorities = append(c.Authorities, authority)
	}
}

// RemoveAuthority removes an authority from the token claims
func (c *Claims) RemoveAuthority(authority string) {
	for i, a := range c.Authorities {
		if a == authority {
			c.Authorities = append(c.Authorities[:i], c.Authorities[i+1:]...)
			return
		}
	}
}

// AddScope adds a scope to the token claims
func (c *Claims) AddScope(scope string) {
	if !c.HasScope(scope) {
		c.Scopes = append(c.Scopes, scope)
	}
}

// RemoveScope removes a scope from the token claims
func (c *Claims) RemoveScope(scope string) {
	for i, s := range c.Scopes {
		if s == scope {
			c.Scopes = append(c.Scopes[:i], c.Scopes[i+1:]...)
			return
		}
	}
}

// ToMap converts claims to a map for easy marshaling
func (c *Claims) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Marshal to JSON and back to get all fields
	data, _ := json.Marshal(c)
	json.Unmarshal(data, &result)
	
	return result
}

// FromMap populates claims from a map
func (c *Claims) FromMap(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, c)
}

// Clone creates a deep copy of the claims
func (c *Claims) Clone() *Claims {
	clone := &Claims{
		RegisteredClaims:  c.RegisteredClaims,
		Name:              c.Name,
		Email:             c.Email,
		EmailVerified:     c.EmailVerified,
		PreferredUsername: c.PreferredUsername,
		GivenName:         c.GivenName,
		FamilyName:        c.FamilyName,
		OrganizationID:    c.OrganizationID,
		DepartmentID:      c.DepartmentID,
		EmployeeID:        c.EmployeeID,
		TenantID:          c.TenantID,
		SessionID:         c.SessionID,
		ClientID:          c.ClientID,
		TokenType:         c.TokenType,
	}
	
	// Deep copy slices
	if c.Roles != nil {
		clone.Roles = make([]string, len(c.Roles))
		copy(clone.Roles, c.Roles)
	}
	
	if c.Authorities != nil {
		clone.Authorities = make([]string, len(c.Authorities))
		copy(clone.Authorities, c.Authorities)
	}
	
	if c.Scopes != nil {
		clone.Scopes = make([]string, len(c.Scopes))
		copy(clone.Scopes, c.Scopes)
	}
	
	// Deep copy custom claims
	if c.CustomClaims != nil {
		clone.CustomClaims = make(map[string]interface{})
		for k, v := range c.CustomClaims {
			clone.CustomClaims[k] = v
		}
	}
	
	return clone
}

// Validate performs basic validation on the claims
func (c *Claims) Validate() error {
	if c.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	
	// Validate registered claims
	now := time.Now()
	
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Time) {
		return fmt.Errorf("token has expired")
	}
	
	if c.NotBefore != nil && now.Before(c.NotBefore.Time) {
		return fmt.Errorf("token not yet valid")
	}
	
	return nil
}