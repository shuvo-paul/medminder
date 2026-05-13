// Package oauth provides OAuth 2.0 authentication provider support.
//
// This package defines the Provider interface and a registry for OAuth providers.
// Adding a new provider (e.g., GitHub) requires only creating a new package
// under oauth/ and implementing the Provider interface.
package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Errors returned by the OAuth provider registry.
var (
	ErrProviderNotFound      = errors.New("oauth: provider not found")
	ErrProviderConfigMissing = errors.New("oauth: provider configuration missing")
)

// Config holds OAuth provider configuration from environment variables.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// TokenResponse represents the response from an OAuth token exchange.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents user information obtained from an OAuth provider.
type UserInfo struct {
	ProviderUserID string // The provider's sub / user ID
	Email          string
	EmailVerified  bool   // true if provider verified the email
	Name           string // display name default
}

// Provider defines the interface for OAuth providers.
// Implement this interface to add support for a new OAuth provider.
type Provider interface {
	// Name returns the provider's identifier (e.g., "google").
	Name() string

	// RequiredEnvVars returns the environment variable names required
	// for this provider to function (e.g., ["GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"]).
	RequiredEnvVars() []string

	// AuthURL returns the URL to redirect users to for authorization.
	// The state parameter should be included in the redirect.
	AuthURL(state string) string

	// ExchangeCode exchanges an authorization code for access and refresh tokens.
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)

	// GetUserInfo fetches user information using the provided access token.
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
}

// ProviderInfo represents publicly available information about a provider.
type ProviderInfo struct {
	ID          string `json:"id"`           // provider identifier (e.g., "google")
	Name        string `json:"name"`        // display name (e.g., "Google")
	IconURL     string `json:"icon_url"`    // path to provider icon
	CallbackPath string `json:"callback_path"` // OAuth callback path
}

// providerRegistry holds the registered OAuth providers.
type providerRegistry struct {
	mu         sync.RWMutex
	providers  map[string]Provider
	providerInfo map[string]ProviderInfo
}

// globalRegistry is the global provider registry.
var globalRegistry = &providerRegistry{
	providers:    make(map[string]Provider),
	providerInfo: make(map[string]ProviderInfo),
}

// Register registers an OAuth provider with the global registry.
// This is typically called from a provider's init() function.
// Returns ErrProviderConfigMissing if required environment variables are not set.
func Register(provider Provider) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := provider.Name()

	// Check required env vars
	for _, envVar := range provider.RequiredEnvVars() {
		if os.Getenv(envVar) == "" {
			return fmt.Errorf("%w: %s", ErrProviderConfigMissing, envVar)
		}
	}

	globalRegistry.providers[name] = provider

	// Store provider info for public API
	globalRegistry.providerInfo[name] = ProviderInfo{
		ID:           name,
		Name:         strings.Title(name),
		IconURL:      fmt.Sprintf("/assets/icons/%s.svg", name),
		CallbackPath: fmt.Sprintf("/api/auth/oauth/%s/callback", name),
	}

	return nil
}

// GetProvider returns the provider with the given name.
// Returns ErrProviderNotFound if the provider is not registered.
func GetProvider(name string) (Provider, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	provider, ok := globalRegistry.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	return provider, nil
}

// GetProviders returns a slice of ProviderInfo for all registered providers.
func GetProviders() []ProviderInfo {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(globalRegistry.providerInfo))
	for _, info := range globalRegistry.providerInfo {
		infos = append(infos, info)
	}
	return infos
}

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
		return nil, fmt.Errorf("oauth: invalid state encoding: %w", err)
	}

	var state OAuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("oauth: invalid state JSON: %w", err)
	}

	return &state, nil
}
