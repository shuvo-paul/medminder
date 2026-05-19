package handlers

import (
	"context"
	"errors"
	"net/http"

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

		// Get the provider from registry
		provider, err := oauth.GetProvider(input.Provider)
		if err != nil {
			if errors.Is(err, oauth.ErrProviderNotFound) {
				return nil, huma.Error404NotFound("provider not found", err)
			}
			return nil, huma.Error500InternalServerError("failed to get provider", err)
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
		Method:     http.MethodGet,
		Path:       "/api/auth/oauth/providers",
		Summary:    "List configured OAuth providers",
		Tags:       []string{"auth"},
	}, ListOAuthProvidersHandler())

	// Initiate OAuth - unauthenticated
	huma.Register(api, huma.Operation{
		OperationID: "initiate-oauth",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/{provider}",
		Summary:     "Initiate OAuth login",
		Tags:        []string{"auth"},
	}, InitiateOAuthHandler())
}
