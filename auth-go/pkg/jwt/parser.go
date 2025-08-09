package jwt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// Parser handles JWT token parsing and validation
type Parser struct {
	issuer         string
	audience       string
	signingKey     []byte
	signingMethod  jwt.SigningMethod
}

// NewParser creates a new JWT parser
func NewParser(issuer, audience, signingKey string) *Parser {
	return &Parser{
		issuer:        issuer,
		audience:      audience,
		signingKey:    []byte(signingKey),
		signingMethod: jwt.SigningMethodHS256, // Default to HMAC-SHA256
	}
}

// NewParserWithSigningMethod creates a new JWT parser with a specific signing method
func NewParserWithSigningMethod(issuer, audience, signingKey string, method jwt.SigningMethod) *Parser {
	return &Parser{
		issuer:        issuer,
		audience:      audience,
		signingKey:    []byte(signingKey),
		signingMethod: method,
	}
}

// Parse parses a JWT token string and returns the parsed token
func (p *Parser) Parse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method != p.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.signingKey, nil
	})

	if err != nil {
		log.Error().Err(err).Str("token", tokenString[:min(50, len(tokenString))]).Msg("Failed to parse JWT")
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return token, nil
}

// ParseWithClaims parses a JWT token and extracts custom claims
func (p *Parser) ParseWithClaims(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method != p.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.signingKey, nil
	})

	if err != nil {
		log.Error().Err(err).Str("token", tokenString[:min(50, len(tokenString))]).Msg("Failed to parse JWT with claims")
		return nil, fmt.Errorf("failed to parse JWT with claims: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return token, nil
}

// ExtractClaims extracts claims from a JWT token without validation
func (p *Parser) ExtractClaims(tokenString string) (*Claims, error) {
	// Split the token to get the payload
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (claims) without validation
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// Validate validates a JWT token and returns the claims
func (p *Parser) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := p.ParseWithClaims(tokenString, claims)
	if err != nil {
		return nil, err
	}

	// Additional validation
	if err := p.validateClaims(token, claims); err != nil {
		return nil, fmt.Errorf("claims validation failed: %w", err)
	}

	log.Debug().
		Str("subject", claims.Subject).
		Str("issuer", claims.Issuer).
		Time("expires", claims.ExpiresAt.Time).
		Msg("JWT token validated successfully")

	return claims, nil
}

// validateClaims performs additional claims validation
func (p *Parser) validateClaims(token *jwt.Token, claims *Claims) error {
	now := time.Now()

	// Validate issuer
	if p.issuer != "" && claims.Issuer != p.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", p.issuer, claims.Issuer)
	}

	// Validate audience
	if p.audience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == p.audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience: %v does not contain %s", claims.Audience, p.audience)
		}
	}

	// Validate expiration
	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time) {
		return fmt.Errorf("token has expired at %v", claims.ExpiresAt.Time)
	}

	// Validate not before
	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time) {
		return fmt.Errorf("token is not valid before %v", claims.NotBefore.Time)
	}

	// Validate issued at
	if claims.IssuedAt != nil && now.Before(claims.IssuedAt.Time.Add(-time.Minute)) {
		// Allow 1 minute clock skew
		return fmt.Errorf("token issued in the future at %v", claims.IssuedAt.Time)
	}

	return nil
}

// CreateToken creates a new JWT token with the specified claims
func (p *Parser) CreateToken(claims *Claims) (string, error) {
	// Set default claims if not provided
	now := time.Now()
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(now)
	}
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Hour))
	}
	if claims.Issuer == "" {
		claims.Issuer = p.issuer
	}
	if len(claims.Audience) == 0 && p.audience != "" {
		claims.Audience = []string{p.audience}
	}

	token := jwt.NewWithClaims(p.signingMethod, claims)
	tokenString, err := token.SignedString(p.signingKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create JWT token")
		return "", fmt.Errorf("failed to create JWT token: %w", err)
	}

	log.Debug().
		Str("subject", claims.Subject).
		Str("issuer", claims.Issuer).
		Time("expires", claims.ExpiresAt.Time).
		Msg("JWT token created successfully")

	return tokenString, nil
}

// RefreshToken creates a new token with extended expiration based on an existing token
func (p *Parser) RefreshToken(tokenString string, newExpiry time.Duration) (string, error) {
	claims, err := p.ExtractClaims(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to extract claims from token: %w", err)
	}

	// Update the expiration and issued at times
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(newExpiry))

	return p.CreateToken(claims)
}

// GetTokenInfo returns basic information about a token without validation
func (p *Parser) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	claims, err := p.ExtractClaims(tokenString)
	if err != nil {
		return nil, err
	}

	info := &TokenInfo{
		Subject:   claims.Subject,
		Issuer:    claims.Issuer,
		Audience:  claims.Audience,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
		NotBefore: claims.NotBefore.Time,
		Roles:     claims.Roles,
		Scopes:    claims.Scopes,
	}

	if claims.NotBefore != nil {
		info.NotBefore = claims.NotBefore.Time
	}

	return info, nil
}

// IsExpired checks if a token is expired without full validation
func (p *Parser) IsExpired(tokenString string) (bool, error) {
	info, err := p.GetTokenInfo(tokenString)
	if err != nil {
		return true, err
	}

	return time.Now().After(info.ExpiresAt), nil
}

// TokenInfo represents basic token information
type TokenInfo struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	Audience  []string  `json:"audience"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	NotBefore time.Time `json:"not_before"`
	Roles     []string  `json:"roles"`
	Scopes    []string  `json:"scopes"`
}

// MarshalJSON implements json.Marshaler for TokenInfo
func (ti *TokenInfo) MarshalJSON() ([]byte, error) {
	type Alias TokenInfo
	return json.Marshal(&struct {
		IssuedAt  int64 `json:"issued_at"`
		ExpiresAt int64 `json:"expires_at"`
		NotBefore int64 `json:"not_before"`
		*Alias
	}{
		IssuedAt:  ti.IssuedAt.Unix(),
		ExpiresAt: ti.ExpiresAt.Unix(),
		NotBefore: ti.NotBefore.Unix(),
		Alias:     (*Alias)(ti),
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}