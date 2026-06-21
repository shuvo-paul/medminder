package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	auditRepo "github.com/shuvo-paul/medminder/internal/features/audit/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

const oauthAuthCodeExpiry = 5 * time.Minute

// ListOAuthProvidersHandler returns a handler that lists all configured OAuth providers.
func ListOAuthProvidersHandler() func(context.Context, *struct{}) (*dto.ProvidersResponse, error) {
	return func(ctx context.Context, input *struct{}) (*dto.ProvidersResponse, error) {
		providers := oauth.GetProviders()

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
func InitiateOAuthHandler() func(context.Context, *dto.InitiateOAuthInput) (*dto.InitiateOAuthOutput, error) {
	return func(ctx context.Context, input *dto.InitiateOAuthInput) (*dto.InitiateOAuthOutput, error) {
		if input.State == "" {
			return nil, huma.Error400BadRequest("state parameter is required", errors.New("missing state parameter"))
		}

		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid state parameter", err)
		}

		if state.Redirect != "" && !isValidRedirect(state.Redirect) {
			return nil, huma.Error400BadRequest("invalid redirect URL", errors.New("redirect must be a relative path"))
		}

		provider, err := oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		authURL := provider.AuthURL(input.State)

		return &dto.InitiateOAuthOutput{
			Location: authURL,
			Status:   http.StatusFound,
		}, nil
	}
}

// OAuthHandlerDeps groups dependencies needed by OAuth handlers.
type OAuthHandlerDeps struct {
	AuthCodeRepo repository.OAuthAuthorizationCodeRepository
	OAuthSvc     service.OAuthService
	TokenSvc     service.TokenServiceInterface
	TokenRepo    repository.RefreshTokenRepository
	AuditRepo    auditRepo.AuditRepository
	FrontendURL  string
}

// RegisterOAuthProviderRoutes registers the OAuth provider routes.
func RegisterOAuthProviderRoutes(api huma.API, authCodeRepo repository.OAuthAuthorizationCodeRepository, oauthSvc service.OAuthService, tokenSvc service.TokenServiceInterface, tokenRepo repository.RefreshTokenRepository, auditRepository auditRepo.AuditRepository, frontendURL string) {
	deps := &OAuthHandlerDeps{
		AuthCodeRepo: authCodeRepo,
		OAuthSvc:     oauthSvc,
		TokenSvc:     tokenSvc,
		TokenRepo:    tokenRepo,
		AuditRepo:    auditRepository,
		FrontendURL:  frontendURL,
	}

	huma.Register(api, huma.Operation{
		OperationID: "list-oauth-providers",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/providers",
		Summary:     "List configured OAuth providers",
		Tags:        []string{"auth"},
	}, ListOAuthProvidersHandler())

	huma.Register(api, huma.Operation{
		OperationID: "initiate-oauth",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/{provider}",
		Summary:     "Initiate OAuth login",
		Tags:        []string{"auth"},
	}, InitiateOAuthHandler())

	huma.Register(api, huma.Operation{
		OperationID: "oauth-callback",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/{provider}/callback",
		Summary:     "Handle OAuth callback",
		Tags:        []string{"auth"},
	}, OAuthCallbackHandler(deps))

	huma.Register(api, huma.Operation{
		OperationID: "oauth-token-exchange",
		Method:      http.MethodPost,
		Path:        "/api/auth/oauth/token",
		Summary:     "Exchange OAuth authorization code for JWT tokens",
		Tags:        []string{"auth"},
	}, TokenExchangeHandler(deps))

	huma.Register(api, huma.Operation{
		OperationID: "oauth-link-init",
		Method:      http.MethodPost,
		Path:        "/api/auth/oauth/{provider}/init",
		Summary:     "Initiate OAuth provider linking",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, OAuthLinkInitHandler(deps))

	huma.Register(api, huma.Operation{
		OperationID: "list-oauth-accounts",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/accounts",
		Summary:     "List linked OAuth accounts",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, OAuthAccountsHandler(deps))

	huma.Register(api, huma.Operation{
		OperationID: "unlink-oauth-account",
		Method:      http.MethodDelete,
		Path:        "/api/auth/oauth/accounts/{provider}",
		Summary:     "Unlink an OAuth provider",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, OAuthUnlinkHandler(deps))

	huma.Register(api, huma.Operation{
		OperationID: "oauth-link-status",
		Method:      http.MethodGet,
		Path:        "/api/auth/oauth/accounts/{provider}/status",
		Summary:     "Check OAuth link status for a provider",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, OAuthLinkStatusHandler(deps))
}
