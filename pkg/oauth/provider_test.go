package oauth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockProvider implements Provider for testing.
type mockProvider struct {
	nameVal            string
	requiredEnvVars    []string
	authURLVal         string
	exchangeCodeResp   *TokenResponse
	exchangeCodeErr    error
	getUserInfoResp    *UserInfo
	getUserInfoErr     error
}

func (m *mockProvider) Name() string                        { return m.nameVal }
func (m *mockProvider) RequiredEnvVars() []string           { return m.requiredEnvVars }
func (m *mockProvider) AuthURL(state string) string         { return m.authURLVal }
func (m *mockProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	return m.exchangeCodeResp, m.exchangeCodeErr
}
func (m *mockProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	return m.getUserInfoResp, m.getUserInfoErr
}

func TestGetProvider_NotFound(t *testing.T) {
	// Clear registry first by creating a fresh one
	registry := &providerRegistry{
		providers:    make(map[string]Provider),
		providerInfo: make(map[string]ProviderInfo),
	}
	_ = registry // unused but shows intent

	_, err := GetProvider("nonexistent")
	assert.True(t, errors.Is(err, ErrProviderNotFound))
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetProvider_Success(t *testing.T) {
	// Set the required env var for the test provider
	t.Setenv("TEST_VAR", "test-value")

	provider := &mockProvider{
		nameVal:         "test",
		requiredEnvVars: []string{"TEST_VAR"},
		authURLVal:      "https://example.com/auth",
	}

	// Register our mock
	err := Register(provider)
	assert.NoError(t, err)

	// Retrieve and verify
	retrieved, err := GetProvider("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", retrieved.Name())
}

func TestGetProviders(t *testing.T) {
	// Set the required env var for the test provider
	t.Setenv("TEST_PROVIDER_VAR", "test-value")

	// Register a provider first
	provider := &mockProvider{
		nameVal:         "testprovider",
		requiredEnvVars: []string{"TEST_PROVIDER_VAR"},
		authURLVal:      "https://example.com/auth",
	}
	_ = Register(provider)

	providers := GetProviders()
	assert.NotEmpty(t, providers)

	// Check that our test provider is in the list
	found := false
	for _, p := range providers {
		if p.ID == "testprovider" {
			found = true
			assert.Equal(t, "Testprovider", p.Name)
			assert.Equal(t, "/assets/icons/testprovider.svg", p.IconURL)
			assert.Equal(t, "/api/auth/oauth/testprovider/callback", p.CallbackPath)
		}
	}
	assert.True(t, found, "testprovider should be in the providers list")
}

func TestOAuthState_Encode(t *testing.T) {
	state := &OAuthState{
		Nonce:    "test-nonce-123",
		Redirect: "/dashboard",
		Purpose:  "login",
	}

	encoded := state.Encode()
	assert.NotEmpty(t, encoded)
	assert.NotEqual(t, "test-nonce-123", encoded) // Should be base64 encoded
}

func TestDecodeOAuthState(t *testing.T) {
	state := &OAuthState{
		Nonce:    "test-nonce-123",
		Redirect: "/dashboard",
		Purpose:  "login",
	}

	encoded := state.Encode()
	decoded, err := DecodeOAuthState(encoded)

	assert.NoError(t, err)
	assert.Equal(t, state.Nonce, decoded.Nonce)
	assert.Equal(t, state.Redirect, decoded.Redirect)
	assert.Equal(t, state.Purpose, decoded.Purpose)
}

func TestDecodeOAuthState_InvalidEncoding(t *testing.T) {
	_, err := DecodeOAuthState("not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state encoding")
}

func TestDecodeOAuthState_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	encoded := "bm90LWpzb24=" // "not-json" in base64
	_, err := DecodeOAuthState(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state JSON")
}

func TestRegister_MissingEnvVar(t *testing.T) {
	// Create a provider that requires a non-existent env var
	provider := &mockProvider{
		nameVal:         "missingenv",
		requiredEnvVars: []string{"NON_EXISTENT_ENV_VAR_12345"},
		authURLVal:      "https://example.com/auth",
	}

	err := Register(provider)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrProviderConfigMissing))
	assert.Contains(t, err.Error(), "NON_EXISTENT_ENV_VAR_12345")
}
