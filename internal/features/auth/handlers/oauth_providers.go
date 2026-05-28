package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

// oauthAuthCodeExpiry is the lifetime of an internal authorization code.
const oauthAuthCodeExpiry = 5 * time.Minute

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

// OAuthHandlerDeps groups dependencies needed by OAuth handlers.
type OAuthHandlerDeps struct {
	AuthCodeRepo repository.OAuthAuthorizationCodeRepository
	OAuthSvc     service.OAuthService
	TokenSvc     service.TokenServiceInterface
	TokenRepo    repository.RefreshTokenRepository
}

// RegisterOAuthProviderRoutes registers the OAuth provider routes.
func RegisterOAuthProviderRoutes(api huma.API, authCodeRepo repository.OAuthAuthorizationCodeRepository, oauthSvc service.OAuthService, tokenSvc service.TokenServiceInterface, tokenRepo repository.RefreshTokenRepository) {
	deps := &OAuthHandlerDeps{
		AuthCodeRepo: authCodeRepo,
		OAuthSvc:     oauthSvc,
		TokenSvc:     tokenSvc,
		TokenRepo:    tokenRepo,
	}

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
	}, OAuthCallbackHandler(deps))

	// Token exchange - unauthenticated
	huma.Register(api, huma.Operation{
		OperationID: "oauth-token-exchange",
		Method:      http.MethodPost,
		Path:        "/api/auth/oauth/token",
		Summary:     "Exchange OAuth authorization code for JWT tokens",
		Tags:        []string{"auth"},
	}, TokenExchangeHandler(deps))
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

// OAuthCallbackHandler returns a handler that handles the OAuth callback from the provider.
// GET /api/auth/oauth/{provider}/callback
//
// Two modes:
//   - Error mode (provider returns ?error=...): redirects to frontend with error info
//   - Success mode (provider returns ?code=...): exchanges code, stores user info internally,
//     and redirects to frontend callback route with an internal authorization code.
func OAuthCallbackHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
	return func(ctx context.Context, input *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
		// --- Mode 1: Provider error ---
		if input.Error != "" {
			return handleProviderError(input)
		}

		// --- Mode 2: Success (authorization code received) ---
		if input.Code == "" {
			return nil, huma.Error400BadRequest("missing authorization code", errors.New("code parameter is required"))
		}

		// Parse and validate state
		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return &dto.OAuthCallbackOutput{
				Redirect: "/login?oauth_error=invalid_state",
			}, nil
		}

		// Get the provider from registry
		provider, err := oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		// Exchange the authorization code for tokens
		tokenResp, err := provider.ExchangeCode(ctx, input.Code)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to exchange authorization code", err)
		}

		// Fetch user info from provider
		userInfo, err := provider.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get user info", err)
		}

		// Generate internal authorization code
		internalCode, codeHash, err := generateAuthCode()
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to generate authorization code", err)
		}

		// Store in database (5-minute expiry, SHA-256 hashed)
		expiresAt := time.Now().Add(oauthAuthCodeExpiry)
		_, err = deps.AuthCodeRepo.CreateAuthorizationCodeWithUserInfo(
			ctx,
			uuid.New(),
			codeHash,
			state.Nonce,
			state.Purpose,
			expiresAt,
			input.Provider,
			userInfo.ProviderUserID,
			userInfo.Email,
			userInfo.Name,
			userInfo.EmailVerified,
		)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to store authorization code", err)
		}

		// Encode original state to pass back to frontend
		encodedState, err := json.Marshal(state)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to encode state", err)
		}

		// 302 redirect to frontend callback route (NO JWT tokens in URL)
		redirectURL := fmt.Sprintf("/auth/callback?code=%s&state=%s",
			url.QueryEscape(internalCode),
			url.QueryEscape(string(encodedState)),
		)

		return &dto.OAuthCallbackOutput{
			Redirect: redirectURL,
		}, nil
	}
}

// TokenExchangeHandler returns a handler that exchanges an internal OAuth authorization
// code for JWT access and refresh tokens.
// POST /api/auth/oauth/token
func TokenExchangeHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
	return func(ctx context.Context, input *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
		// Parse and validate state
		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return nil, huma.Error401Unauthorized("invalid state", err)
		}

		// Hash the received code
		codeHash := hashCode(input.Code)

		// Look up and lock the authorization code (prevents concurrent redemption)
		authCodeInfo, err := deps.AuthCodeRepo.GetAndLockAuthorizationCode(ctx, codeHash)
		if err != nil {
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("invalid or expired authorization code"))
		}

		// Verify nonce matches (CSRF protection)
		if authCodeInfo.Nonce != state.Nonce {
			// Mark as used to prevent replay attempts
			_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("state nonce mismatch"))
		}

		// Get or create user by OAuth
		oauthUser, err := deps.OAuthSvc.GetOrCreateUserByOAuth(ctx, authCodeInfo.Provider, &oauth.UserInfo{
			ProviderUserID: authCodeInfo.ProviderUserID,
			Email:          authCodeInfo.ProviderEmail,
			EmailVerified:  authCodeInfo.ProviderEmailVerified,
			Name:           authCodeInfo.ProviderName,
		})
		if err != nil {
			// Mark code as used even on error to prevent further attempts
			_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)

			if errors.Is(err, service.ErrEmailExists) {
				return nil, huma.Error409Conflict(string(dto.OAuthErrorEmailExists),
					fmt.Errorf("email already exists"))
			}
			return nil, huma.Error500InternalServerError("authentication failed", err)
		}

		// Mark authorization code as used (one-time use)
		_, err = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to mark code as used", err)
		}

		// Generate JWT access token
		accessToken, err := deps.TokenSvc.GenerateAccessToken(oauthUser.ID, oauthUser.Email)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to generate access token", err)
		}

		// Generate refresh token
		refreshToken, err := deps.TokenSvc.GenerateRefreshToken()
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to generate refresh token", err)
		}

		// Store refresh token hash in database
		refreshTokenHash := deps.TokenSvc.HashRefreshToken(refreshToken)
		refreshExpiry := time.Now().Add(service.RefreshTokenExpiry)
		if _, err := deps.TokenRepo.CreateRefreshToken(ctx, oauthUser.ID, refreshTokenHash, refreshExpiry); err != nil {
			return nil, huma.Error500InternalServerError("failed to store refresh token", err)
		}

		return &dto.OAuthTokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int(service.AccessTokenExpiry.Seconds()),
			User: dto.OAuthTokenUserInfo{
				ID:            oauthUser.ID.String(),
				Email:         oauthUser.Email,
				DisplayName:   oauthUser.DisplayName,
				EmailVerified: oauthUser.EmailVerified,
			},
		}, nil
	}
}

// handleProviderError handles the error case where the OAuth provider returned an error
// (e.g., user denied access). It attempts to parse the state to redirect back to the
// original frontend page, falling back to /login if state is absent or malformed.
func handleProviderError(input *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
	if input.State == "" {
		return &dto.OAuthCallbackOutput{
			Redirect: "/login?oauth_error=cancelled",
		}, nil
	}

	state, err := dto.ParseOAuthState(input.State)
	if err != nil || state.Redirect == "" {
		return &dto.OAuthCallbackOutput{
			Redirect: "/login?oauth_error=cancelled",
		}, nil
	}

	redirectURL := fmt.Sprintf("%s?oauth_error=cancelled&provider=%s",
		state.Redirect,
		url.QueryEscape(input.Provider),
	)

	return &dto.OAuthCallbackOutput{
		Redirect: redirectURL,
	}, nil
}

// generateAuthCode creates a cryptographically random authorization code and its SHA-256 hash.
// Returns the raw code (for the frontend) and the hash (for database storage).
func generateAuthCode() (rawCode, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random code: %w", err)
	}
	rawCode = hex.EncodeToString(bytes)
	hash = hashCode(rawCode)
	return rawCode, hash, nil
}

// hashCode returns the SHA-256 hex hash of the given code.
func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
