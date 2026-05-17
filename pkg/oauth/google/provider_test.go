package google_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/shuvo-paul/medminder/pkg/oauth/google"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_REDIRECT_URI", "https://example.com/callback")
	defer func() {
		os.Unsetenv("GOOGLE_CLIENT_ID")
		os.Unsetenv("GOOGLE_CLIENT_SECRET")
		os.Unsetenv("GOOGLE_REDIRECT_URI")
	}()

	provider, err := google.New()
	require.NoError(t, err)
	assert.Equal(t, "test-client-id", provider.ClientID)
	assert.Equal(t, "test-client-secret", provider.ClientSecret)
	assert.Equal(t, "https://example.com/callback", provider.RedirectURI)
}

func TestProvider_Name(t *testing.T) {
	provider := &google.GoogleProvider{}
	assert.Equal(t, "google", provider.Name())
}

func TestProvider_RequiredEnvVars(t *testing.T) {
	provider := &google.GoogleProvider{}
	envVars := provider.RequiredEnvVars()
	assert.Contains(t, envVars, "GOOGLE_CLIENT_ID")
	assert.Contains(t, envVars, "GOOGLE_CLIENT_SECRET")
	assert.Contains(t, envVars, "GOOGLE_REDIRECT_URI")
}

func TestProvider_AuthURL(t *testing.T) {
	provider := &google.GoogleProvider{
		ClientID:    "test-client-id",
		RedirectURI: "https://example.com/callback",
	}

	authURL := provider.AuthURL("test-state")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "redirect_uri=https%3A%2F%2Fexample.com%2Fcallback")
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "scope=openid+email+profile")
	assert.Contains(t, authURL, "state=test-state")
}

func TestProvider_ExchangeCode_Success(t *testing.T) {
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

	oldTokenURL := google.TokenURL
	google.TokenURL = server.URL
	defer func() { google.TokenURL = oldTokenURL }()

	provider := &google.GoogleProvider{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "https://example.com/callback",
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

	oldTokenURL := google.TokenURL
	google.TokenURL = server.URL
	defer func() { google.TokenURL = oldTokenURL }()

	provider := &google.GoogleProvider{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "https://example.com/callback",
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

	oldUserInfoURL := google.UserInfoURL
	google.UserInfoURL = server.URL
	defer func() { google.UserInfoURL = oldUserInfoURL }()

	provider := &google.GoogleProvider{}

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

	oldUserInfoURL := google.UserInfoURL
	google.UserInfoURL = server.URL
	defer func() { google.UserInfoURL = oldUserInfoURL }()

	provider := &google.GoogleProvider{}

	_, err := provider.GetUserInfo(context.Background(), "invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userinfo request failed")
}

func TestTokenResponseWithExpiry_IsExpired(t *testing.T) {
	resp := &oauth.TokenResponse{
		AccessToken: "test",
		ExpiresIn:   3600,
	}
	expiryResp := google.NewTokenResponseWithExpiry(resp)
	assert.False(t, expiryResp.IsExpired())

	expiredResp := &oauth.TokenResponse{
		AccessToken: "test",
		ExpiresIn:   -1,
	}
	expiryExpiredResp := google.NewTokenResponseWithExpiry(expiredResp)
	assert.True(t, expiryExpiredResp.IsExpired())
}
