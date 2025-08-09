package oauth2

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"github.com/erpmicroservices/auth-go/pkg/cache"
)

// Validator handles token validation using both introspection and JWT verification
type Validator struct {
	client      *Client
	jwksCache   cache.Cache
	jwksURL     string
	httpClient  *http.Client
	audience    string
	issuer      string
}

// NewValidator creates a new token validator
func NewValidator(client *Client, jwksURL, issuer, audience string) *Validator {
	return &Validator{
		client:      client,
		jwksCache:   cache.NewInMemoryCache(1*time.Hour, 2*time.Hour),
		jwksURL:     jwksURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		audience:    audience,
		issuer:      issuer,
	}
}

// ValidateTokenWithIntrospection validates a token using OAuth2 introspection
func (v *Validator) ValidateTokenWithIntrospection(ctx context.Context, token string) (*TokenIntrospection, error) {
	return v.client.ValidateToken(ctx, token)
}

// ValidateJWT validates a JWT token using public key verification
func (v *Validator) ValidateJWT(ctx context.Context, tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no key ID found in token header")
		}

		// Get the public key for verification
		return v.getPublicKey(ctx, kid)
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JWT token")
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("JWT token is not valid")
	}

	// Validate claims
	if err := v.validateClaims(token); err != nil {
		return nil, fmt.Errorf("JWT claims validation failed: %w", err)
	}

	return token, nil
}

// ValidateToken validates a token using both introspection and JWT verification (if JWT)
func (v *Validator) ValidateToken(ctx context.Context, tokenString string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid: false,
		Token: tokenString,
	}

	// First try JWT validation if it looks like a JWT
	if v.isJWT(tokenString) {
		jwtToken, err := v.ValidateJWT(ctx, tokenString)
		if err == nil {
			result.Valid = true
			result.JWT = jwtToken
			result.Claims = v.extractClaims(jwtToken)
			log.Debug().Msg("Token validated successfully using JWT verification")
			return result, nil
		}
		
		log.Debug().Err(err).Msg("JWT validation failed, falling back to introspection")
	}

	// Fall back to introspection
	introspection, err := v.ValidateTokenWithIntrospection(ctx, tokenString)
	if err != nil {
		log.Error().Err(err).Msg("Token validation failed")
		return result, fmt.Errorf("token validation failed: %w", err)
	}

	result.Valid = introspection.Active && !introspection.IsExpired()
	result.Introspection = introspection

	if result.Valid {
		log.Debug().Msg("Token validated successfully using introspection")
	}

	return result, nil
}

// getPublicKey retrieves the public key for JWT verification from JWKS
func (v *Validator) getPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("jwks_key_%s", keyID)
	if cachedKey, found := v.jwksCache.Get(cacheKey); found {
		if key, ok := cachedKey.(*rsa.PublicKey); ok {
			return key, nil
		}
	}

	// Fetch JWKS
	jwks, err := v.fetchJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Find the key
	for _, key := range jwks.Keys {
		if key.Kid == keyID {
			publicKey, err := v.parseRSAPublicKey(key)
			if err != nil {
				return nil, err
			}

			// Cache the key
			v.jwksCache.Set(cacheKey, publicKey)
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("public key with ID %s not found", keyID)
}

// fetchJWKS fetches the JSON Web Key Set from the authorization server
func (v *Validator) fetchJWKS(ctx context.Context) (*JWKS, error) {
	// Check cache first
	const jwksCacheKey = "jwks"
	if cachedJWKS, found := v.jwksCache.Get(jwksCacheKey); found {
		if jwks, ok := cachedJWKS.(*JWKS); ok {
			return jwks, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", v.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWKS: %w", err)
	}

	// Cache the JWKS
	v.jwksCache.Set(jwksCacheKey, &jwks)

	return &jwks, nil
}

// parseRSAPublicKey parses an RSA public key from JWK format
func (v *Validator) parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	// Parse the RSA public key from JWK format
	// This is a simplified implementation - in production, you might want to use a library like go-jose
	// that handles all the JWK format details properly
	return nil, fmt.Errorf("RSA public key parsing not implemented - use a JWK library like go-jose")
}

// validateClaims validates JWT claims
func (v *Validator) validateClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims type")
	}

	// Validate issuer
	if v.issuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != v.issuer {
			return fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, iss)
		}
	}

	// Validate audience
	if v.audience != "" {
		if aud, ok := claims["aud"]; ok {
			var audienceValid bool
			switch audVal := aud.(type) {
			case string:
				audienceValid = audVal == v.audience
			case []interface{}:
				for _, a := range audVal {
					if audStr, ok := a.(string); ok && audStr == v.audience {
						audienceValid = true
						break
					}
				}
			}
			if !audienceValid {
				return fmt.Errorf("invalid audience")
			}
		}
	}

	// Validate expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return fmt.Errorf("token has expired")
		}
	}

	// Validate not before
	if nbf, ok := claims["nbf"].(float64); ok {
		if time.Now().Unix() < int64(nbf) {
			return fmt.Errorf("token not yet valid")
		}
	}

	return nil
}

// extractClaims extracts user information from JWT claims
func (v *Validator) extractClaims(token *jwt.Token) *UserInfo {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	userInfo := &UserInfo{}

	if sub, ok := claims["sub"].(string); ok {
		userInfo.Subject = sub
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		userInfo.PreferredUsername = preferredUsername
	}

	// Extract roles and authorities
	if roles, ok := claims["roles"].([]interface{}); ok {
		userInfo.Roles = make([]string, 0, len(roles))
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				userInfo.Roles = append(userInfo.Roles, roleStr)
			}
		}
	}

	if authorities, ok := claims["authorities"].([]interface{}); ok {
		userInfo.Authorities = make([]string, 0, len(authorities))
		for _, authority := range authorities {
			if authStr, ok := authority.(string); ok {
				userInfo.Authorities = append(userInfo.Authorities, authStr)
			}
		}
	}

	return userInfo
}

// isJWT checks if a token string looks like a JWT (has three base64 segments separated by dots)
func (v *Validator) isJWT(tokenString string) bool {
	parts := strings.Split(tokenString, ".")
	return len(parts) == 3
}

// ValidationResult represents the result of token validation
type ValidationResult struct {
	Valid         bool
	Token         string
	JWT           *jwt.Token
	Claims        *UserInfo
	Introspection *TokenIntrospection
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c,omitempty"`
	X5t string   `json:"x5t,omitempty"`
}