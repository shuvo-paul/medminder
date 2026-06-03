package dto

import (
	"errors"
	"time"

	"github.com/shuvo-paul/medminder/pkg/oauth"
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

// OAuthState is the state parameter in OAuth flows.
// It is a type alias for pkg/oauth.OAuthState so all encoding/decoding
// logic lives in a single place.
type OAuthState = oauth.OAuthState

// OAuthTokenRequest represents the request body for token exchange.
type OAuthTokenRequest struct {
	Body struct {
		Code  string `json:"code" maxLength:"256"`   // authorization code
		State string `json:"state" maxLength:"1024"` // base64url-encoded JSON: {nonce, redirect, purpose}
	}
}

// OAuthTokenUserInfo represents user information in token response.
type OAuthTokenUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
}

// OAuthTokenResponseBody is the JSON body of an OAuth token exchange response.
type OAuthTokenResponseBody struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	TokenType    string             `json:"token_type"` // "Bearer"
	ExpiresIn    int                `json:"expires_in"` // 86400 (24h)
	User         OAuthTokenUserInfo `json:"user"`
}

// OAuthTokenResponse represents the response for token exchange.
type OAuthTokenResponse struct {
	Body OAuthTokenResponseBody
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

// OAuthLinkInitInput represents the input for initiating an OAuth link flow.
// This endpoint is authenticated — the Authorization header contains the Bearer token.
type OAuthLinkInitInput struct {
	Authorization string `header:"Authorization"`
	Provider      string `path:"provider" doc:"OAuth provider identifier (e.g., google)"`
	Body          struct {
		Redirect string `json:"redirect" doc:"Where to return after linking (e.g., /profile)"`
	}
}

// OAuthLinkInitResponse represents the response for initiating an OAuth link flow.
type OAuthLinkInitResponse struct {
	Body struct {
		State string `json:"state" doc:"Base64-encoded state with purpose=link"`
	}
}

// SetPasswordInput represents the input for setting a password for an OAuth-only user.
// This endpoint is authenticated — the Authorization header contains the Bearer token.
type SetPasswordInput struct {
	Authorization string `header:"Authorization"`
	Body          struct {
		Password string `json:"password" minLength:"8" doc:"New password (8+ chars, 1 uppercase, 1 lowercase, 1 number)"`
	}
}

// SetPasswordOutput represents the output for setting a password.
type SetPasswordOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// InitiateOAuthInput represents the input for initiating OAuth login.
// Path parameter: provider - the OAuth provider name (e.g., "google")
// Query parameter: state - the OAuth state parameter from the client
type InitiateOAuthInput struct {
	Provider string `path:"provider" doc:"OAuth provider identifier (e.g., google)"` // OAuth provider identifier
	State    string `query:"state"`                                                  // OAuth state parameter (required, non-empty)
}

// InitiateOAuthOutput represents the output for initiating OAuth login.
// Returns a 302 redirect to the provider's authorization URL.
type InitiateOAuthOutput struct {
	Location string `header:"Location"` // Redirect URL
	Status   int    `status:"302"`      // HTTP redirect status code
}

// OAuthCallbackInput represents the input for OAuth callback.
// Path parameter: provider - the OAuth provider name (e.g., "google")
// Query parameters: code - authorization code, state - OAuth state parameter
type OAuthCallbackInput struct {
	Provider string `path:"provider"` // OAuth provider identifier
	Code     string `query:"code"`    // Authorization code from provider
	State    string `query:"state"`   // OAuth state parameter
	Error    string `query:"error"`   // Error from provider (if any)
}

// OAuthCallbackOutput represents the output for OAuth callback.
type OAuthCallbackOutput struct {
	Redirect string `header:"Location"` // Redirect URL after callback
	Status   int    `status:"302"`      // HTTP redirect status code
}

// OAuthAccountsInput represents the input for listing linked OAuth accounts.
// This endpoint is authenticated via Bearer token in Authorization header.
type OAuthAccountsInput struct {
	Authorization string `header:"Authorization"`
}

// OAuthLinkedAccount represents a single linked OAuth provider account.
type OAuthLinkedAccount struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	CreatedAt      time.Time `json:"created_at"`
	ProviderName   string    `json:"provider_name"`
}

// OAuthAccountsResponse represents the response for listing linked OAuth accounts.
type OAuthAccountsResponse struct {
	Body struct {
		Accounts    []OAuthLinkedAccount `json:"accounts"`
		HasPassword bool                 `json:"has_password"`
	}
}

// OAuthUnlinkInput represents the input for unlinking an OAuth provider.
type OAuthUnlinkInput struct {
	Authorization string `header:"Authorization"`
	Provider      string `path:"provider" doc:"OAuth provider identifier (e.g., google)"`
}

// OAuthUnlinkOutput represents the output for unlinking an OAuth provider.
type OAuthUnlinkOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// OAuthLinkStatusInput represents the input for checking OAuth link status.
type OAuthLinkStatusInput struct {
	Authorization string `header:"Authorization"`
	Provider      string `path:"provider" doc:"OAuth provider identifier (e.g., google)"`
}

// OAuthLinkStatusResponse represents the response for checking OAuth link status.
type OAuthLinkStatusResponse struct {
	Body struct {
		Linked      bool `json:"linked"`
		CanUnlink   bool `json:"can_unlink"`
		HasPassword bool `json:"has_password"`
	}
}

// ParseOAuthState parses the OAuth state from a string.
func ParseOAuthState(stateStr string) (*OAuthState, error) {
	return oauth.DecodeOAuthState(stateStr)
}
