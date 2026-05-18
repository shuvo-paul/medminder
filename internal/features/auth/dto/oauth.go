package dto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// OAuthProvider represents an OAuth provider.
type OAuthProvider string

// Valid OAuth providers.
const (
	OAuthProviderGoogle OAuthProvider = "google"
)

// OAuthErrorCode represents OAuth error codes.
type OAuthErrorCode string

// OAuth error codes.
const (
	OAuthErrorInvalidCode   OAuthErrorCode = "invalid_code"
	OAuthErrorEmailExists   OAuthErrorCode = "email_exists"
	OAuthErrorAccountLocked OAuthErrorCode = "account_locked"
	OAuthErrorLinkFailed    OAuthErrorCode = "link_failed"
)

// OAuthState represents the state parameter in OAuth flows.
type OAuthState struct {
	Nonce    string `json:"nonce"`    // crypto.randomUUID() from client
	Redirect string `json:"redirect"` // e.g., "/dashboard"
	Purpose  string `json:"purpose"`  // "register" | "login" | "link"
}

// Encode encodes the OAuthState as a base64url-encoded string.
func (s *OAuthState) Encode() string {
	data, _ := json.Marshal(s)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeOAuthState decodes a base64url-encoded OAuthState string.
func DecodeOAuthState(encoded string) (*OAuthState, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid state encoding: %w", err)
	}

	var state OAuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid state JSON: %w", err)
	}

	return &state, nil
}

// OAuthTokenRequest represents the request body for token exchange.
type OAuthTokenRequest struct {
	Code  string `json:"code" maxLength:"256"`   // authorization code
	State string `json:"state" maxLength:"1024"` // base64url-encoded JSON: {nonce, redirect, purpose}
}

// OAuthTokenUserInfo represents user information in token response.
type OAuthTokenUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
}

// OAuthTokenResponse represents the response for token exchange.
type OAuthTokenResponse struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	TokenType    string             `json:"token_type"` // "Bearer"
	ExpiresIn    int                `json:"expires_in"` // 86400 (24h)
	User         OAuthTokenUserInfo `json:"user"`
}

// OAuthInitInput represents the input for initiating OAuth link flow.
type OAuthInitInput struct {
	Redirect string `json:"redirect"` // where to return after link (e.g., "/settings")
}

// OAuthInitResponse represents the response for initiating OAuth link flow.
type OAuthInitResponse struct {
	State string `json:"state"` // base64url-encoded JSON with nonce + redirect + purpose="link"
}

// OAuthLinkRequest represents the request to link an OAuth account.
type OAuthLinkRequest struct {
	Provider OAuthProvider `json:"provider" enum:"google"` // OAuth provider
}

// Validate validates the OAuthLinkRequest.
func (r *OAuthLinkRequest) Validate() error {
	if r.Provider != OAuthProviderGoogle {
		return errors.New("invalid provider: must be google")
	}
	return nil
}

// OAuthUnlinkResponse represents the response for unlinking an OAuth account.
type OAuthUnlinkResponse struct {
	Message string `json:"message"`
}

// OAuthErrorResponse represents an OAuth error response.
type OAuthErrorResponse struct {
	Error       string `json:"error"`       // "invalid_code" | "email_exists" | "account_locked" | "link_failed"
	Description string `json:"description"` // human-readable
	Provider    string `json:"provider"`    // e.g., "google" (for email_exists)
	Email       string `json:"email"`       // for email_exists
	Redirect    string `json:"redirect"`    // original redirect URL (echoed back)
}

// NewOAuthErrorResponse creates a new OAuthErrorResponse with error descriptions.
func NewOAuthErrorResponse(code OAuthErrorCode, provider, email, redirect string) *OAuthErrorResponse {
	resp := &OAuthErrorResponse{
		Error:    string(code),
		Provider: provider,
		Email:    email,
		Redirect: redirect,
	}

	switch code {
	case OAuthErrorInvalidCode:
		resp.Description = "Your sign-in session expired. Please try again."
	case OAuthErrorEmailExists:
		resp.Description = "An account with this email already exists. Sign in with your password to link your Google account."
	case OAuthErrorAccountLocked:
		resp.Description = "You must set a password before linking this account, or you will lose access to your account."
	case OAuthErrorLinkFailed:
		resp.Description = "Failed to link account. Please try again."
	default:
		resp.Description = "An error occurred. Please try again."
	}

	return resp
}

// OAuthUserInfo represents user information from an OAuth provider.
type OAuthUserInfo struct {
	ProviderUserID string // The provider's sub / user ID
	Email          string
	EmailVerified  bool   // true if provider verified the email
	Name           string // display name
}

// OAuthProviderInfo represents publicly available information about an OAuth provider.
type OAuthProviderInfo struct {
	ID           string `json:"id"`            // provider identifier (e.g., "google")
	Name         string `json:"name"`          // display name (e.g., "Google")
	IconURL      string `json:"icon_url"`      // path to provider icon
	CallbackPath string `json:"callback_path"` // OAuth callback path
}

// ProvidersResponse represents the response for listing OAuth providers.
type ProvidersResponse struct {
	Providers []OAuthProviderInfo `json:"providers"`
}

// ParseOAuthState parses the OAuth state from a string.
// This is a convenience function that wraps DecodeOAuthState.
func ParseOAuthState(stateStr string) (*OAuthState, error) {
	// Handle URL encoding (replace - with + and _ with /)
	stateStr = strings.ReplaceAll(stateStr, "-", "+")
	stateStr = strings.ReplaceAll(stateStr, "_", "/")

	// Add padding if needed
	padding := 4 - len(stateStr)%4
	if padding < 4 {
		stateStr += strings.Repeat("=", padding)
	}

	return DecodeOAuthState(stateStr)
}
