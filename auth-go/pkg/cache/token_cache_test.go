package cache

import (
	"testing"
	"time"
)

func TestInMemoryTokenCache(t *testing.T) {
	cache := NewInMemoryTokenCache(1*time.Hour, 10*time.Minute)

	// Test setting and getting a token
	token := &Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	cache.Set("test-key", token)

	// Retrieve the token
	retrieved, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find cached token")
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("Expected access token %s, got %s", token.AccessToken, retrieved.AccessToken)
	}

	// Test size
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Test delete
	cache.Delete("test-key")
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after delete, got %d", cache.Size())
	}

	// Test that deleted key is not found
	_, found = cache.Get("test-key")
	if found {
		t.Error("Expected not to find deleted token")
	}
}

func TestTokenValidity(t *testing.T) {
	// Valid token
	validToken := &Token{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	if !validToken.IsValid() {
		t.Error("Expected token to be valid")
	}

	if validToken.IsExpired() {
		t.Error("Expected token not to be expired")
	}

	// Expired token
	expiredToken := &Token{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}

	if expiredToken.IsValid() {
		t.Error("Expected expired token to be invalid")
	}

	if !expiredToken.IsExpired() {
		t.Error("Expected token to be expired")
	}
}

func TestInMemoryCache(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Test basic operations
	cache.Set("key1", "value1")
	
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find cached value")
	}

	if str, ok := value.(string); !ok || str != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Test size
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Test clear
	cache.Clear()
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}