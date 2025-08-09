package cache

import "time"

// Token represents an OAuth2 access token (duplicated to avoid circular imports)
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