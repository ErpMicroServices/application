package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/erpmicroservices/auth-go/internal/config"
	"github.com/erpmicroservices/auth-go/pkg/cache"
)

// Client represents an OAuth2 client for the ERP microservices
type Client struct {
	config              *config.OAuth2Config
	oauth2Config        *oauth2.Config
	clientCredsConfig   *clientcredentials.Config
	httpClient          *http.Client
	cache               cache.TokenCache
	userInfoCache       cache.Cache
	introspectionCache  cache.Cache
}

// NewClient creates a new OAuth2 client
func NewClient(cfg *config.OAuth2Config, tokenCache cache.TokenCache) *Client {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthorizeURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
		Scopes:      cfg.Scopes,
	}

	clientCredsConfig := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       cfg.Scopes,
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Create cache instances with appropriate TTLs
	userInfoCache := cache.NewInMemoryCache(5*time.Minute, 10*time.Minute)
	introspectionCache := cache.NewInMemoryCache(2*time.Minute, 5*time.Minute)

	return &Client{
		config:              cfg,
		oauth2Config:        oauth2Config,
		clientCredsConfig:   clientCredsConfig,
		httpClient:          httpClient,
		cache:               tokenCache,
		userInfoCache:       userInfoCache,
		introspectionCache:  introspectionCache,
	}
}

// GetAuthorizationURL returns the authorization URL for the OAuth2 flow
func (c *Client) GetAuthorizationURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (c *Client) ExchangeCodeForToken(ctx context.Context, code string) (*Token, error) {
	oauth2Token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange code for token")
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	token := &Token{
		AccessToken:  oauth2Token.AccessToken,
		TokenType:    oauth2Token.TokenType,
		RefreshToken: oauth2Token.RefreshToken,
		ExpiresAt:    oauth2Token.Expiry,
	}

	// Store token in cache
	c.cache.Set(token.AccessToken, tokenToCache(token))

	log.Info().Msg("Successfully exchanged code for token")
	return token, nil
}

// GetClientCredentialsToken gets a token using client credentials flow
func (c *Client) GetClientCredentialsToken(ctx context.Context) (*Token, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("client_credentials_%s", c.config.ClientID)
	if cachedToken, found := c.cache.Get(cacheKey); found {
		if cachedToken.IsValid() {
			token := tokenFromCache(cachedToken)
			log.Debug().Msg("Using cached client credentials token")
			return token, nil
		}
	}

	oauth2Token, err := c.clientCredsConfig.Token(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get client credentials token")
		return nil, fmt.Errorf("failed to get client credentials token: %w", err)
	}

	token := &Token{
		AccessToken: oauth2Token.AccessToken,
		TokenType:   oauth2Token.TokenType,
		ExpiresAt:   oauth2Token.Expiry,
	}

	// Store token in cache
	c.cache.Set(cacheKey, tokenToCache(token))

	log.Info().Msg("Successfully obtained client credentials token")
	return token, nil
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	tokenSource := c.oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	oauth2Token, err := tokenSource.Token()
	if err != nil {
		log.Error().Err(err).Msg("Failed to refresh token")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	token := &Token{
		AccessToken:  oauth2Token.AccessToken,
		TokenType:    oauth2Token.TokenType,
		RefreshToken: oauth2Token.RefreshToken,
		ExpiresAt:    oauth2Token.Expiry,
	}

	// Update cache
	c.cache.Set(token.AccessToken, tokenToCache(token))

	log.Info().Msg("Successfully refreshed token")
	return token, nil
}

// GetUserInfo retrieves user information using an access token
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// Check cache first
	if cachedUserInfo, found := c.userInfoCache.Get(accessToken); found {
		if userInfo, ok := cachedUserInfo.(*UserInfo); ok {
			log.Debug().Msg("Using cached user info")
			return userInfo, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("User info request failed")
		return nil, fmt.Errorf("user info request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	// Cache user info
	c.userInfoCache.Set(accessToken, &userInfo)

	log.Info().Str("subject", userInfo.Subject).Msg("Successfully retrieved user info")
	return &userInfo, nil
}

// IntrospectToken introspects a token to check its validity and get metadata
func (c *Client) IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	// Check cache first
	if cachedIntrospection, found := c.introspectionCache.Get(token); found {
		if introspection, ok := cachedIntrospection.(*TokenIntrospection); ok {
			log.Debug().Msg("Using cached token introspection")
			return introspection, nil
		}
	}

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.IntrospectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to introspect token")
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Token introspection failed")
		return nil, fmt.Errorf("token introspection failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read introspection response: %w", err)
	}

	var introspection TokenIntrospection
	if err := json.Unmarshal(body, &introspection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal introspection response: %w", err)
	}

	// Cache introspection result
	c.introspectionCache.Set(token, &introspection)

	log.Debug().Bool("active", introspection.Active).Msg("Token introspection completed")
	return &introspection, nil
}

// ValidateToken validates a token by introspecting it
func (c *Client) ValidateToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	introspection, err := c.IntrospectToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !introspection.Active {
		return nil, fmt.Errorf("token is not active")
	}

	if introspection.ExpiresAt != nil && time.Now().After(*introspection.ExpiresAt) {
		return nil, fmt.Errorf("token is expired")
	}

	return introspection, nil
}

// RevokeToken revokes an access or refresh token
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	// Build revoke URL (assuming Spring Authorization Server pattern)
	revokeURL := strings.TrimSuffix(c.config.AuthorizationServerURL, "/") + "/oauth2/revoke"

	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke token")
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Token revocation failed")
		return fmt.Errorf("token revocation failed with status %d", resp.StatusCode)
	}

	// Remove from cache
	c.cache.Delete(token)
	c.userInfoCache.Delete(token)
	c.introspectionCache.Delete(token)

	log.Info().Msg("Successfully revoked token")
	return nil
}

// ClearCaches clears all internal caches
func (c *Client) ClearCaches() {
	c.userInfoCache.Clear()
	c.introspectionCache.Clear()
	log.Info().Msg("Cleared OAuth2 client caches")
}

// tokenToCache converts an OAuth2 Token to a cache Token
func tokenToCache(token *Token) *cache.Token {
	if token == nil {
		return nil
	}
	return &cache.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
		Scope:        token.Scope,
	}
}

// tokenFromCache converts a cache Token to an OAuth2 Token
func tokenFromCache(cacheToken *cache.Token) *Token {
	if cacheToken == nil {
		return nil
	}
	return &Token{
		AccessToken:  cacheToken.AccessToken,
		TokenType:    cacheToken.TokenType,
		RefreshToken: cacheToken.RefreshToken,
		ExpiresAt:    cacheToken.ExpiresAt,
		Scope:        cacheToken.Scope,
	}
}