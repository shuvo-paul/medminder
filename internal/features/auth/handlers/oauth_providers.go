package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/shuvo-paul/medminder/internal/common/log"
	auditRepo "github.com/shuvo-paul/medminder/internal/features/audit/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

// oauthAuthCodeExpiry is the lifetime of an internal authorization code.
const oauthAuthCodeExpiry = 5 * time.Minute

// logCodeRejected logs an oauth_code_rejected audit event for invalid or misused
// authorization codes. userID is nil (anonymous) because the code hasn't been
// associated with a user at this point.
func logCodeRejected(ctx context.Context, auditRepo auditRepo.AuditRepository, reason, provider string) {
	ip := middleware.IPFromContext(ctx)
	ua := middleware.UserAgentFromContext(ctx)
	metadata := map[string]string{"reason": reason}
	if provider != "" {
		metadata["provider"] = provider
	}
	if err := auditRepo.LogEvent(ctx, "oauth_code_rejected", uuid.NullUUID{}, ip, ua, metadata); err != nil {
		log.Warn("audit_log_failed", log.F("event", "oauth_code_rejected"), log.F("error", err.Error()))
	}
}

// logOAuthLoginFailed logs an oauth_login_failed audit event.
func logOAuthLoginFailed(ctx context.Context, auditRepo auditRepo.AuditRepository, provider, email string) {
	ip := middleware.IPFromContext(ctx)
	ua := middleware.UserAgentFromContext(ctx)
	if err := auditRepo.LogEvent(ctx, "oauth_login_failed", uuid.NullUUID{}, ip, ua, map[string]string{
		"provider": provider,
		"reason":   "email_exists",
		"email":    email,
	}); err != nil {
		log.Warn("audit_log_failed", log.F("event", "oauth_login_failed"), log.F("error", err.Error()))
	}
}

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

	// Initiate OAuth link - authenticated
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

	// List linked OAuth accounts - authenticated
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

	// Unlink OAuth account - authenticated
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

	// OAuth link status - authenticated
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
			return handleProviderError(input, deps.FrontendURL)
		}

		// --- Mode 2: Success (authorization code received) ---
		if input.Code == "" {
			return nil, huma.Error400BadRequest("missing authorization code", errors.New("code parameter is required"))
		}

		// Parse and validate state
		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return &dto.OAuthCallbackOutput{
				Redirect: deps.FrontendURL + "/login?oauth_error=invalid_state",
				Status:   http.StatusFound,
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

		if state.Purpose == "link" {
			// Link flow: find the original code (created by OAuthLinkInitHandler)
			// to associate the provider account with the authenticated user.
			originalCode, err := deps.AuthCodeRepo.GetAuthorizationCodeByNonceAndPurpose(ctx, state.Nonce, "link")
			if err != nil || !originalCode.UserID.Valid {
				return nil, huma.Error500InternalServerError("failed to find original authorization code for linking", err)
			}

			_, err = deps.AuthCodeRepo.CreateAuthorizationCodeWithUserInfoForLink(
				ctx,
				uuid.New(),
				codeHash,
				originalCode.UserID.UUID,
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
				return nil, huma.Error500InternalServerError("failed to store authorization code for linking", err)
			}
		} else {
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
		}

		// Encode original state base64url to pass back to frontend
		encodedState := state.Encode()

		// 302 redirect to frontend callback route (NO JWT tokens in URL)
		redirectURL := fmt.Sprintf("%s/auth/callback?code=%s&state=%s",
			deps.FrontendURL,
			url.QueryEscape(internalCode),
			url.QueryEscape(encodedState),
		)

		return &dto.OAuthCallbackOutput{
			Redirect: redirectURL,
			Status:   http.StatusFound,
		}, nil
	}
}

// TokenExchangeHandler returns a handler that exchanges an internal OAuth authorization
// code for JWT access and refresh tokens.
// POST /api/auth/oauth/token
func TokenExchangeHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
	return func(ctx context.Context, input *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
		// Parse and validate state
		state, err := dto.ParseOAuthState(input.Body.State)
		if err != nil {
			return nil, huma.Error401Unauthorized("invalid state", err)
		}

		// Hash the received code
		codeHash := hashCode(input.Body.Code)

		// Look up and lock the authorization code (prevents concurrent redemption)
		authCodeInfo, err := deps.AuthCodeRepo.GetAndLockAuthorizationCode(ctx, codeHash)
		if err != nil {
			logCodeRejected(ctx, deps.AuditRepo, "invalid_code", "")
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("invalid or expired authorization code"))
		}

		// Verify nonce matches (CSRF protection)
		if authCodeInfo.Nonce != state.Nonce {
			// Mark as used to prevent replay attempts
			_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
			logCodeRejected(ctx, deps.AuditRepo, "nonce_mismatch", authCodeInfo.Provider)
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("state nonce mismatch"))
		}

		var oauthUser *service.OAuthUser

		if authCodeInfo.Purpose == "link" {
			// Link flow: associate the provider account with the existing user.
			if !authCodeInfo.UserID.Valid {
				_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
				return nil, huma.Error500InternalServerError("invalid authorization code", errors.New("missing user binding"))
			}

			if err := deps.OAuthSvc.LinkOAuthAccount(ctx, authCodeInfo.UserID.UUID, authCodeInfo.Provider, authCodeInfo.ProviderUserID); err != nil {
				_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)

				if errors.Is(err, service.ErrAccountWillBeLocked) {
					return nil, huma.Error403Forbidden(string(dto.OAuthErrorAccountLocked),
						fmt.Errorf("account will be locked out"))
				}
				if errors.Is(err, service.ErrProviderAlreadyLinked) {
					return nil, huma.Error409Conflict(string(dto.OAuthErrorLinkFailed),
						fmt.Errorf("already linked to another account"))
				}
				log.Error("oauth_token_exchange_link_account_failed",
					log.F("error", err.Error()),
					log.F("provider", authCodeInfo.Provider),
					log.F("provider_user_id", authCodeInfo.ProviderUserID),
				)
				return nil, huma.Error500InternalServerError("failed to link account", err)
			}

			// Fetch the user to generate tokens
			user, err := deps.OAuthSvc.GetUserByOAuth(ctx, authCodeInfo.Provider, authCodeInfo.ProviderUserID)
			if err != nil {
				return nil, huma.Error500InternalServerError("failed to get user", err)
			}
			oauthUser = user
		} else {
			// Login/register flow: get or create user by OAuth
			oauthUser, err = deps.OAuthSvc.GetOrCreateUserByOAuth(ctx, authCodeInfo.Provider, &oauth.UserInfo{
				ProviderUserID: authCodeInfo.ProviderUserID,
				Email:          authCodeInfo.ProviderEmail,
				EmailVerified:  authCodeInfo.ProviderEmailVerified,
				Name:           authCodeInfo.ProviderName,
			})
			if err != nil {
				// Mark code as used even on error to prevent further attempts
				_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)

				if errors.Is(err, service.ErrEmailExists) {
					logOAuthLoginFailed(ctx, deps.AuditRepo, authCodeInfo.Provider, authCodeInfo.ProviderEmail)
					return nil, huma.Error409Conflict(string(dto.OAuthErrorEmailExists),
						fmt.Errorf(`{"email":"%s","provider":"%s"}`, authCodeInfo.ProviderEmail, authCodeInfo.Provider))
				}
				log.Error("oauth_token_exchange_get_or_create_user_failed",
					log.F("error", err.Error()),
					log.F("provider", authCodeInfo.Provider),
					log.F("provider_user_id", authCodeInfo.ProviderUserID),
					log.F("provider_email", authCodeInfo.ProviderEmail),
				)
				return nil, huma.Error500InternalServerError("authentication failed", err)
			}
		}

		// Mark authorization code as used (one-time use)
		_, err = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
		if err != nil {
			log.Error("oauth_token_exchange_mark_code_used_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to mark code as used", err)
		}

		// Generate JWT access token
		accessToken, err := deps.TokenSvc.GenerateAccessToken(oauthUser.ID, oauthUser.Email)
		if err != nil {
			log.Error("oauth_token_exchange_generate_access_token_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to generate access token", err)
		}

		// Generate refresh token
		refreshToken, err := deps.TokenSvc.GenerateRefreshToken()
		if err != nil {
			log.Error("oauth_token_exchange_generate_refresh_token_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to generate refresh token", err)
		}

		// Store refresh token hash in database
		refreshTokenHash := deps.TokenSvc.HashRefreshToken(refreshToken)
		refreshExpiry := time.Now().Add(service.RefreshTokenExpiry)
		if _, err := deps.TokenRepo.CreateRefreshToken(ctx, oauthUser.ID, refreshTokenHash, refreshExpiry); err != nil {
			log.Error("oauth_token_exchange_store_refresh_token_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to store refresh token", err)
		}

		return &dto.OAuthTokenResponse{
			Body: dto.OAuthTokenResponseBody{
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
			},
		}, nil
	}
}

// handleProviderError handles the error case where the OAuth provider returned an error
// (e.g., user denied access). It attempts to parse the state to redirect back to the
// original frontend page, falling back to /login if state is absent or malformed.
func handleProviderError(input *dto.OAuthCallbackInput, frontendURL string) (*dto.OAuthCallbackOutput, error) {
	if input.State == "" {
		return &dto.OAuthCallbackOutput{
			Redirect: frontendURL + "/login?oauth_error=cancelled",
			Status:   http.StatusFound,
		}, nil
	}

	state, err := dto.ParseOAuthState(input.State)
	if err != nil || state.Redirect == "" {
		return &dto.OAuthCallbackOutput{
			Redirect: frontendURL + "/login?oauth_error=cancelled",
			Status:   http.StatusFound,
		}, nil
	}

	redirectURL := fmt.Sprintf("%s%s?oauth_error=cancelled&provider=%s",
		frontendURL,
		state.Redirect,
		url.QueryEscape(input.Provider),
	)

	return &dto.OAuthCallbackOutput{
		Redirect: redirectURL,
		Status:   http.StatusFound,
	}, nil
}

// ExtractUserIDFromAuth extracts the authenticated user ID from the Authorization header.
// It parses the Bearer token, validates it, and returns the user ID from the "sub" claim.
func ExtractUserIDFromAuth(authHeader string, tokenSvc service.TokenServiceInterface) (uuid.UUID, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return uuid.Nil, huma.Error401Unauthorized("Invalid authorization header", nil)
	}
	tokenString := authHeader[7:]

	claims, err := tokenSvc.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, huma.Error401Unauthorized("Invalid or expired access token", err)
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, huma.Error401Unauthorized("Invalid access token", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, huma.Error401Unauthorized("Invalid user ID in token", nil)
	}

	if userID == uuid.Nil {
		return uuid.Nil, huma.Error401Unauthorized("Invalid user ID", nil)
	}

	return userID, nil
}

// OAuthLinkInitHandler returns a handler that initiates an OAuth account linking flow.
// POST /api/auth/oauth/{provider}/init (authenticated)
//
// This stores an authorization code binding the authenticated user to a nonce with
// purpose="link", so the OAuth callback flow can associate the provider account
// with the correct user.
func OAuthLinkInitHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthLinkInitInput) (*dto.OAuthLinkInitResponse, error) {
	return func(ctx context.Context, input *dto.OAuthLinkInitInput) (*dto.OAuthLinkInitResponse, error) {
		// Extract authenticated user ID from Bearer token
		userID, err := ExtractUserIDFromAuth(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, err
		}

		// Verify the provider exists
		_, err = oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		// Generate a cryptographically random nonce
		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			return nil, huma.Error500InternalServerError("failed to generate nonce", err)
		}
		nonce := hex.EncodeToString(nonceBytes)

		// Validate redirect if provided
		redirect := input.Body.Redirect
		if redirect != "" && !isValidRedirect(redirect) {
			return nil, huma.Error400BadRequest("invalid redirect URL", errors.New("redirect must be a relative path"))
		}
		if redirect == "" {
			redirect = "/profile"
		}

		state := dto.OAuthState{
			Nonce:    nonce,
			Redirect: redirect,
			Purpose:  "link",
		}
		encodedState := state.Encode()

		// Store authorization code binding user_id to this nonce.
		// The code_hash uses the encoded state hash as a placeholder since
		// no provider auth code exists yet. The callback handler MUST use
		// GetAuthorizationCodeByNonceAndPurpose (not GetAuthorizationCodeByHash)
		// to find this record, since code_hash is a placeholder, not the provider code.
		codeID := uuid.New()
		codeHash := hashCode(encodedState)
		expiresAt := time.Now().Add(oauthAuthCodeExpiry)

		_, err = deps.AuthCodeRepo.CreateAuthorizationCode(
			ctx,
			codeID,
			codeHash,
			uuid.NullUUID{UUID: userID, Valid: true},
			nonce,
			"link",
			expiresAt,
		)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to store authorization code", err)
		}

		return &dto.OAuthLinkInitResponse{
			Body: struct {
				State string `json:"state" doc:"Base64-encoded state with purpose=link"`
			}{State: encodedState},
		}, nil
	}
}

// SetPasswordHandler returns a handler that sets a password for an OAuth-only user.
// POST /api/auth/password/set (authenticated)
//
// This allows users who registered via OAuth to add a password, enabling
// password-based login and preventing lock-out when unlinking providers.
func SetPasswordHandler(authSvc service.AuthService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.SetPasswordInput) (*dto.SetPasswordOutput, error) {
	return func(ctx context.Context, input *dto.SetPasswordInput) (*dto.SetPasswordOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		if err := ValidatePassword(input.Body.Password); err != nil {
			return nil, huma.Error400BadRequest("invalid password", err)
		}

		if err := authSvc.SetPassword(ctx, userID, input.Body.Password); err != nil {
			return nil, huma.Error500InternalServerError("failed to set password", err)
		}

		return &dto.SetPasswordOutput{
			Body: struct {
				Message string `json:"message" doc:"Success message"`
			}{Message: "password set successfully"},
		}, nil
	}
}

// OAuthAccountsHandler returns a handler that lists all OAuth accounts linked to the user.
// GET /api/auth/oauth/accounts (authenticated)
func OAuthAccountsHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthAccountsInput) (*dto.OAuthAccountsResponse, error) {
	return func(ctx context.Context, input *dto.OAuthAccountsInput) (*dto.OAuthAccountsResponse, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, err
		}

		providers, err := deps.OAuthSvc.GetUserOAuthProviders(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get linked accounts", err)
		}

		hasPassword, err := deps.OAuthSvc.HasPassword(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get user info", err)
		}

		allProviders := oauth.GetProviders()
		providerNames := make(map[string]string)
		for _, p := range allProviders {
			providerNames[p.ID] = p.Name
		}

		accounts := make([]dto.OAuthLinkedAccount, len(providers))
		for i, p := range providers {
			providerName := providerNames[p]
			if providerName == "" {
				providerName = p
			}
			accounts[i] = dto.OAuthLinkedAccount{
				ID:           "",
				Provider:     p,
				ProviderName: providerName,
			}
		}

		return &dto.OAuthAccountsResponse{
			Body: struct {
				Accounts    []dto.OAuthLinkedAccount `json:"accounts"`
				HasPassword bool                     `json:"has_password"`
			}{Accounts: accounts, HasPassword: hasPassword},
		}, nil
	}
}

// OAuthUnlinkHandler returns a handler that unlinks an OAuth provider from the user.
// DELETE /api/auth/oauth/accounts/{provider} (authenticated)
func OAuthUnlinkHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthUnlinkInput) (*dto.OAuthUnlinkOutput, error) {
	return func(ctx context.Context, input *dto.OAuthUnlinkInput) (*dto.OAuthUnlinkOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, err
		}

		if err := deps.OAuthSvc.UnlinkOAuthAccount(ctx, userID, input.Provider); err != nil {
			if errors.Is(err, repository.ErrOAuthAccountNotFound) {
				return nil, huma.Error404NotFound("account not linked", err)
			}
			if errors.Is(err, service.ErrAccountWillBeLocked) {
				return nil, huma.Error403Forbidden("cannot unlink last login method", err)
			}
			return nil, huma.Error500InternalServerError("failed to unlink account", err)
		}

		return &dto.OAuthUnlinkOutput{
			Body: struct {
				Message string `json:"message" doc:"Success message"`
			}{Message: "provider unlinked successfully"},
		}, nil
	}
}

// OAuthLinkStatusHandler returns a handler that checks the OAuth link status for a provider.
// GET /api/auth/oauth/accounts/{provider}/status (authenticated)
func OAuthLinkStatusHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthLinkStatusInput) (*dto.OAuthLinkStatusResponse, error) {
	return func(ctx context.Context, input *dto.OAuthLinkStatusInput) (*dto.OAuthLinkStatusResponse, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, err
		}

		providers, err := deps.OAuthSvc.GetUserOAuthProviders(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get linked accounts", err)
		}

		linked := false
		for _, p := range providers {
			if p == input.Provider {
				linked = true
				break
			}
		}

		canUnlink := false
		if linked {
			canUnlink, err = deps.OAuthSvc.CanUnlinkProvider(ctx, userID, input.Provider)
			if err != nil {
				return nil, huma.Error500InternalServerError("failed to check unlink status", err)
			}
		}

		hasPassword, err := deps.OAuthSvc.HasPassword(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get user info", err)
		}

		return &dto.OAuthLinkStatusResponse{
			Body: struct {
				Linked      bool `json:"linked"`
				CanUnlink   bool `json:"can_unlink"`
				HasPassword bool `json:"has_password"`
			}{
				Linked:      linked,
				CanUnlink:   canUnlink,
				HasPassword: hasPassword,
			},
		}, nil
	}
}

// ChangePasswordHandler returns a handler that changes the authenticated user's password.
// PUT /api/auth/password (authenticated)
func ChangePasswordHandler(authSvc service.AuthService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.ChangePasswordInput) (*dto.ChangePasswordOutput, error) {
	return func(ctx context.Context, input *dto.ChangePasswordInput) (*dto.ChangePasswordOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		if err := ValidatePassword(input.Body.NewPassword); err != nil {
			return nil, huma.Error400BadRequest("invalid new password", err)
		}

		if err := authSvc.ChangePassword(ctx, userID, input.Body.CurrentPassword, input.Body.NewPassword); err != nil {
			if errors.Is(err, service.ErrNoPasswordSet) {
				return nil, huma.Error400BadRequest("no password set", err)
			}
			if errors.Is(err, service.ErrWrongPassword) {
				return nil, huma.Error403Forbidden("current password is incorrect", err)
			}
			return nil, huma.Error500InternalServerError("failed to change password", err)
		}

		return &dto.ChangePasswordOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "password changed successfully"},
		}, nil
	}
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
