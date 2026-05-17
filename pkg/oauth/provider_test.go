package oauth_test

import (
	"errors"
	"testing"

	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/assert"
)

func TestGetProvider_NotFound(t *testing.T) {
	_, err := oauth.GetProvider("nonexistent")
	assert.True(t, errors.Is(err, oauth.ErrProviderNotFound))
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetProvider_Success(t *testing.T) {
	t.Setenv("TEST_VAR", "test-value")

	provider := &oauth.MockProvider{
		NameVal:            "test",
		RequiredEnvVarsVal: []string{"TEST_VAR"},
		AuthURLVal:         "https://example.com/auth",
	}

	err := oauth.Register(provider)
	assert.NoError(t, err)

	retrieved, err := oauth.GetProvider("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", retrieved.Name())
}

func TestGetProviders(t *testing.T) {
	t.Setenv("TEST_PROVIDER_VAR", "test-value")

	provider := &oauth.MockProvider{
		NameVal:            "testprovider",
		RequiredEnvVarsVal: []string{"TEST_PROVIDER_VAR"},
		AuthURLVal:         "https://example.com/auth",
	}
	_ = oauth.Register(provider)

	providers := oauth.GetProviders()
	assert.NotEmpty(t, providers)

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
	state := &oauth.OAuthState{
		Nonce:    "test-nonce-123",
		Redirect: "/dashboard",
		Purpose:  "login",
	}

	encoded := state.Encode()
	assert.NotEmpty(t, encoded)
	assert.NotEqual(t, "test-nonce-123", encoded)
}

func TestDecodeOAuthState(t *testing.T) {
	state := &oauth.OAuthState{
		Nonce:    "test-nonce-123",
		Redirect: "/dashboard",
		Purpose:  "login",
	}

	encoded := state.Encode()
	decoded, err := oauth.DecodeOAuthState(encoded)

	assert.NoError(t, err)
	assert.Equal(t, state.Nonce, decoded.Nonce)
	assert.Equal(t, state.Redirect, decoded.Redirect)
	assert.Equal(t, state.Purpose, decoded.Purpose)
}

func TestDecodeOAuthState_InvalidEncoding(t *testing.T) {
	_, err := oauth.DecodeOAuthState("not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state encoding")
}

func TestDecodeOAuthState_InvalidJSON(t *testing.T) {
	encoded := "bm90LWpzb24="
	_, err := oauth.DecodeOAuthState(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state JSON")
}

func TestRegister_MissingEnvVar(t *testing.T) {
	provider := &oauth.MockProvider{
		NameVal:            "missingenv",
		RequiredEnvVarsVal: []string{"NON_EXISTENT_ENV_VAR_12345"},
		AuthURLVal:         "https://example.com/auth",
	}

	err := oauth.Register(provider)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, oauth.ErrProviderConfigMissing))
	assert.Contains(t, err.Error(), "NON_EXISTENT_ENV_VAR_12345")
}
