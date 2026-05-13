package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Set required env vars
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_REDIRECT_URI", "https://example.com/callback")
	defer func() {
		os.Unsetenv("GOOGLE_CLIENT_ID")
		os.Unsetenv("GOOGLE_CLIENT_SECRET")
		os.Unsetenv("GOOGLE_REDIRECT_URI")
	}()

	provider, err := New()
	require.NoError(t, err)
	assert.Equal(t, "test-client-id", provider.clientID)
	assert.Equal(t, "test-client-secret", provider.clientSecret)
	assert.Equal(t, "https://example.com/callback", provider.redirectURI)
}

func TestProvider_Name(t *testing.T) {
	provider := &googleProvider{}
	assert.Equal(t, "google", provider.Name())
}

func TestProvider_RequiredEnvVars(t *testing.T) {
	provider := &googleProvider{}
	envVars := provider.RequiredEnvVars()
	assert.Contains(t, envVars, "GOOGLE_CLIENT_ID")
	assert.Contains(t, envVars, "GOOGLE_CLIENT_SECRET")
	assert.Contains(t, envVars, "GOOGLE_REDIRECT_URI")
}

func TestProvider_AuthURL(t *testing.T) {
	provider := &googleProvider{
		clientID:    "test-client-id",
		redirectURI: "https://example.com/callback",
	}

	authURL := provider.AuthURL("test-state")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "redirect_uri=https%3A%2F%2Fexample.com%2Fcallback")
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "scope=openid+email+profile")
	assert.Contains(t, authURL, "state=test-state")
}

func TestProvider_ExchangeCode_Success(t *testing.T) {
	// Create a test server that returns a valid token response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-access-token",
			"refresh_token": "test-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"id_token":      "test-id-token",
		})
	}))
	defer server.Close()

	// Temporarily change token URL
	oldTokenURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = oldTokenURL }()

	provider := &googleProvider{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURI:  "https://example.com/callback",
	}

	resp, err := provider.ExchangeCode(context.Background(), "test-auth-code")
	require.NoError(t, err)
	assert.Equal(t, "test-access-token", resp.AccessToken)
	assert.Equal(t, "test-refresh-token", resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 3600, resp.ExpiresIn)
}

func TestProvider_ExchangeCode_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "The code was invalid",
		})
	}))
	defer server.Close()

	oldTokenURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = oldTokenURL }()

	provider := &googleProvider{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURI:  "https://example.com/callback",
	}

	_, err := provider.ExchangeCode(context.Background(), "invalid-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token exchange failed")
}

func TestProvider_GetUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             "google-user-123",
			"email":          "user@example.com",
			"verified_email": true,
			"name":           "Test User",
		})
	}))
	defer server.Close()

	oldUserInfoURL := userInfoURL
	userInfoURL = server.URL
	defer func() { userInfoURL = oldUserInfoURL }()

	provider := &googleProvider{}

	userInfo, err := provider.GetUserInfo(context.Background(), "test-access-token")
	require.NoError(t, err)
	assert.Equal(t, "google-user-123", userInfo.ProviderUserID)
	assert.Equal(t, "user@example.com", userInfo.Email)
	assert.True(t, userInfo.EmailVerified)
	assert.Equal(t, "Test User", userInfo.Name)
}

func TestProvider_GetUserInfo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	oldUserInfoURL := userInfoURL
	userInfoURL = server.URL
	defer func() { userInfoURL = oldUserInfoURL }()

	provider := &googleProvider{}

	_, err := provider.GetUserInfo(context.Background(), "invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userinfo request failed")
}

func TestTokenResponseWithExpiry_IsExpired(t *testing.T) {
	// Create a token that expires in 1 hour
	resp := &oauth.TokenResponse{
		AccessToken: "test",
		ExpiresIn:    3600,
	}
	expiryResp := NewTokenResponseWithExpiry(resp)
	assert.False(t, expiryResp.IsExpired())

	// Create an already expired token
	expiredResp := &oauth.TokenResponse{
		AccessToken: "test",
		ExpiresIn:    -1, // Already expired
	}
	expiryExpiredResp := NewTokenResponseWithExpiry(expiredResp)
	assert.True(t, expiryExpiredResp.IsExpired())
}
