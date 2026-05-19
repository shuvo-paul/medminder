package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

// ListOAuthProvidersHandler returns a handler that lists all configured OAuth providers.
// GET /api/auth/oauth/providers
func ListOAuthProvidersHandler() func(context.Context, *struct{}) (*dto.ProvidersResponse, error) {
	return func(ctx context.Context, input *struct{}) (*dto.ProvidersResponse, error) {
		providers := oauth.GetProviders()

		// Convert pkg/oauth.ProviderInfo to dto.OAuthProviderInfo
		providerInfos := make([]dto.OAuthProviderInfo, len(providers))
		for i, p := range providers {
			providerInfos[i] = dto.OAuthProviderInfo{
				ID:           p.ID,
				Name:         p.Name,
				IconURL:      p.IconURL,
				CallbackPath: p.CallbackPath,
			}
		}

		return &dto.ProvidersResponse{
			Providers: providerInfos,
		}, nil
	}
}

// InitiateOAuthHandler returns a handler that initiates OAuth login.
// GET /api/auth/oauth/{provider}
func InitiateOAuthHandler() func(context.Context, *dto.InitiateOAuthInput) (*dto.InitiateOAuthOutput, error) {
	return func(ctx context.Context, input *dto.InitiateOAuthInput) (*dto.InitiateOAuthOutput, error) {
		// Validate state is non-empty
		if input.State == "" {
			return nil, huma.Error400BadRequest("state parameter is required", errors.New("missing state parameter"))
		}

		// Parse and validate state to prevent open redirect
		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid state parameter", err)
		}

		// Validate redirect is relative or trusted (prevent open redirect)
		if state.Redirect != "" && !isValidRedirect(state.Redirect) {
			return nil, huma.Error400BadRequest("invalid redirect URL", errors.New("redirect must be a relative path"))
		}

		// Get the provider from registry
		provider, err := oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		// Generate the authorization URL with the state
		authURL := provider.AuthURL(input.State)

		return &dto.InitiateOAuthOutput{
			Location: authURL,
		}, nil
	}
}

// RegisterOAuthProviderRoutes registers the OAuth provider routes.
func RegisterOAuthProviderRoutes(api huma.API) {
	// List providers - unauthenticated
	huma.Register(api, huma.Operation{
		OperationID: "list-oauth-providers",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/providers",
		Summary:     "List configured OAuth providers",
		Tags:        []string{"auth"},
	}, ListOAuthProvidersHandler())

	// Initiate OAuth - unauthenticated
	huma.Register(api, huma.Operation{
		OperationID: "initiate-oauth",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/{provider}",
		Summary:     "Initiate OAuth login",
		Tags:        []string{"auth"},
	}, InitiateOAuthHandler())

	// OAuth callback - unauthenticated
	huma.Register(api, huma.Operation{
		OperationID: "oauth-callback",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/{provider}/callback",
		Summary:     "Handle OAuth callback",
		Tags:        []string{"auth"},
	}, OAuthCallbackHandler())
}

// isValidRedirect checks if the redirect URL is a relative path or a trusted domain.
// This prevents open redirect attacks.
func isValidRedirect(redirect string) bool {
	// Allow empty redirect (will use default)
	if redirect == "" {
		return true
	}

	// Allow relative paths starting with /
	if strings.HasPrefix(redirect, "/") {
		return true
	}

	// Allow relative paths without scheme (e.g., "dashboard")
	if !strings.Contains(redirect, "://") && !strings.HasPrefix(redirect, "//") {
		return true
	}

	return false
}

// OAuthCallbackHandler returns a handler that handles OAuth callback.
// GET /api/auth/oauth/{provider}/callback
func OAuthCallbackHandler() func(context.Context, *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
	return func(ctx context.Context, input *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
		// TODO: Implement OAuth callback - exchange code for tokens and complete authentication
		// This is a placeholder that returns the received parameters
		return &dto.OAuthCallbackOutput{
			Redirect: input.State,
		}, nil
	}
}
