package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	require.NoError(t, err)
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

func TestParseOAuthState(t *testing.T) {
	state := &OAuthState{
		Nonce:    "test-nonce",
		Redirect: "/settings",
		Purpose:  "link",
	}

	encoded := state.Encode()
	decoded, err := ParseOAuthState(encoded)

	require.NoError(t, err)
	assert.Equal(t, state.Nonce, decoded.Nonce)
	assert.Equal(t, state.Redirect, decoded.Redirect)
	assert.Equal(t, state.Purpose, decoded.Purpose)
}

func TestParseOAuthState_URLEncoded(t *testing.T) {
	// Test with URL-safe base64 encoding (- and _ instead of + and /)
	state := &OAuthState{
		Nonce:    "test-nonce",
		Redirect: "/settings",
		Purpose:  "link",
	}

	encoded := state.Encode()
	// Replace standard base64 chars with URL-safe ones
	encoded = replaceChars(encoded)
	decoded, err := ParseOAuthState(encoded)

	require.NoError(t, err)
	assert.Equal(t, state.Nonce, decoded.Nonce)
}

func replaceChars(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '+':
			result[i] = '-'
		case '/':
			result[i] = '_'
		default:
			result[i] = s[i]
		}
	}
	return string(result)
}

func TestOAuthLinkRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		provider OAuthProvider
		wantErr bool
	}{
		{"google valid", "google", false},
		{"github valid", "github", false},
		{"apple valid", "apple", false},
		{"invalid provider", "invalid", true},
		{"empty provider", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &OAuthLinkRequest{Provider: tt.provider}
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewOAuthErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		code        OAuthErrorCode
		provider    string
		email       string
		redirect    string
		wantError   string
		wantDesc    string
		wantEmail   string
		wantRedirect string
	}{
		{
			name: "invalid_code",
			code: OAuthErrorInvalidCode,
			provider: "google",
			email: "",
			redirect: "/dashboard",
			wantError: "invalid_code",
			wantDesc: "Your sign-in session expired. Please try again.",
			wantEmail: "",
			wantRedirect: "/dashboard",
		},
		{
			name: "email_exists",
			code: OAuthErrorEmailExists,
			provider: "google",
			email: "user@example.com",
			redirect: "/login",
			wantError: "email_exists",
			wantDesc: "An account with this email already exists",
			wantEmail: "user@example.com",
			wantRedirect: "/login",
		},
		{
			name: "account_locked",
			code: OAuthErrorAccountLocked,
			provider: "google",
			email: "",
			redirect: "/settings",
			wantError: "account_locked",
			wantDesc: "You must set a password before linking",
			wantEmail: "",
			wantRedirect: "/settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		resp := NewOAuthErrorResponse(tt.code, tt.provider, tt.email, tt.redirect)
		assert.Equal(t, tt.wantError, resp.Error)
		assert.Contains(t, resp.Description, tt.wantDesc)
		assert.Equal(t, tt.provider, resp.Provider)
		assert.Equal(t, tt.wantEmail, resp.Email)
		assert.Equal(t, tt.wantRedirect, tt.redirect)
	})
	}
}
