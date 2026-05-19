package handlers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOAuthProvidersHandler(t *testing.T) {
	// Set required environment variables for the mock provider
	os.Setenv("TEST_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_CLIENT_SECRET", "test-client-secret")
	os.Setenv("TEST_REDIRECT_URI", "http://localhost:8080/callback")
	defer os.Unsetenv("TEST_CLIENT_ID")
	defer os.Unsetenv("TEST_CLIENT_SECRET")
	defer os.Unsetenv("TEST_REDIRECT_URI")

	// Register a mock provider for testing
	mockProvider := &oauth.MockProvider{
		NameVal:            "test-provider",
		RequiredEnvVarsVal: []string{"TEST_CLIENT_ID", "TEST_CLIENT_SECRET", "TEST_REDIRECT_URI"},
		AuthURLVal:         "https://test-provider.example.com/auth",
	}
	err := oauth.Register(mockProvider)
	require.NoError(t, err)

	handler := handlers.ListOAuthProvidersHandler()

	// Call handler directly
	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Providers)

	// Find our test provider
	var found bool
	for _, p := range resp.Providers {
		if p.ID == "test-provider" {
			found = true
			assert.Equal(t, "Test-Provider", p.Name) // strings.Title capitalizes each word
			assert.Equal(t, "/assets/icons/test-provider.svg", p.IconURL)
			assert.Equal(t, "/api/auth/oauth/test-provider/callback", p.CallbackPath)
			break
		}
	}
	assert.True(t, found, "test-provider should be in the list")
}

// encodeOAuthState creates a properly encoded OAuthState string for testing
func encodeOAuthState(nonce, redirect, purpose string) string {
	state := dto.OAuthState{
		Nonce:    nonce,
		Redirect: redirect,
		Purpose:  purpose,
	}
	data, _ := json.Marshal(state)
	return base64.URLEncoding.EncodeToString(data)
}

func TestInitiateOAuthHandler_Success(t *testing.T) {
	// Set required environment variables for the mock provider
	os.Setenv("TEST_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_CLIENT_SECRET", "test-client-secret")
	os.Setenv("TEST_REDIRECT_URI", "http://localhost:8080/callback")
	defer os.Unsetenv("TEST_CLIENT_ID")
	defer os.Unsetenv("TEST_CLIENT_SECRET")
	defer os.Unsetenv("TEST_REDIRECT_URI")

	// Register a mock provider for testing
	mockProvider := &oauth.MockProvider{
		NameVal:            "test-provider",
		RequiredEnvVarsVal: []string{"TEST_CLIENT_ID", "TEST_CLIENT_SECRET", "TEST_REDIRECT_URI"},
		AuthURLVal:         "https://test-provider.example.com/auth",
	}
	_ = oauth.Register(mockProvider)

	handler := handlers.InitiateOAuthHandler()

	// Create input with provider and properly encoded state
	state := encodeOAuthState("test-nonce-123", "/dashboard", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "test-provider",
		State:    state,
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://test-provider.example.com/auth", resp.Location)
}

func TestInitiateOAuthHandler_MissingState(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	// Create input without state
	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    "",
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "state")
}

func TestInitiateOAuthHandler_InvalidState(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	// Create input with invalid state (not base64 encoded)
	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    "not-valid-base64!!!",
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid state")
}

func TestInitiateOAuthHandler_OpenRedirect(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	// Create state with open redirect URL
	state := encodeOAuthState("test-nonce", "https://evil.com/steal-token", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    state,
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "redirect")
}

func TestInitiateOAuthHandler_ProviderNotFound(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	// Create input with unknown provider and valid state
	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "unknown-provider",
		State:    state,
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Provider not found"))
}
