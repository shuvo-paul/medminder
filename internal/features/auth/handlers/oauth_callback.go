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
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

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

// isValidRedirect checks if the redirect URL is a relative path or a trusted domain.
func isValidRedirect(redirect string) bool {
	if redirect == "" {
		return true
	}
	if strings.HasPrefix(redirect, "/") {
		return true
	}
	if !strings.Contains(redirect, "://") && !strings.HasPrefix(redirect, "//") {
		return true
	}
	return false
}

// OAuthCallbackHandler returns a handler that handles the OAuth callback from the provider.
func OAuthCallbackHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
	return func(ctx context.Context, input *dto.OAuthCallbackInput) (*dto.OAuthCallbackOutput, error) {
		if input.Error != "" {
			return handleProviderError(input, deps.FrontendURL)
		}

		if input.Code == "" {
			return nil, huma.Error400BadRequest("missing authorization code", errors.New("code parameter is required"))
		}

		state, err := dto.ParseOAuthState(input.State)
		if err != nil {
			return &dto.OAuthCallbackOutput{
				Redirect: deps.FrontendURL + "/login?oauth_error=invalid_state",
				Status:   http.StatusFound,
			}, nil
		}

		provider, err := oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		tokenResp, err := provider.ExchangeCode(ctx, input.Code)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to exchange authorization code", err)
		}

		userInfo, err := provider.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get user info", err)
		}

		internalCode, codeHash, err := generateAuthCode()
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to generate authorization code", err)
		}

		expiresAt := time.Now().Add(oauthAuthCodeExpiry)

		if state.Purpose == "link" {
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

		encodedState := state.Encode()

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

// handleProviderError handles the error case where the OAuth provider returned an error.
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

// TokenExchangeHandler returns a handler that exchanges an internal OAuth authorization
// code for JWT access and refresh tokens.
func TokenExchangeHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
	return func(ctx context.Context, input *dto.OAuthTokenRequest) (*dto.OAuthTokenResponse, error) {
		state, err := dto.ParseOAuthState(input.Body.State)
		if err != nil {
			return nil, huma.Error401Unauthorized("invalid state", err)
		}

		codeHash := hashCode(input.Body.Code)

		authCodeInfo, err := deps.AuthCodeRepo.GetAndLockAuthorizationCode(ctx, codeHash)
		if err != nil {
			logCodeRejected(ctx, deps.AuditRepo, "invalid_code", "")
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("invalid or expired authorization code"))
		}

		if authCodeInfo.Nonce != state.Nonce {
			_, _ = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
			logCodeRejected(ctx, deps.AuditRepo, "nonce_mismatch", authCodeInfo.Provider)
			return nil, huma.Error401Unauthorized(string(dto.OAuthErrorInvalidCode),
				fmt.Errorf("state nonce mismatch"))
		}

		var oauthUser *service.OAuthUser

		if authCodeInfo.Purpose == "link" {
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

			user, err := deps.OAuthSvc.GetUserByOAuth(ctx, authCodeInfo.Provider, authCodeInfo.ProviderUserID)
			if err != nil {
				return nil, huma.Error500InternalServerError("failed to get user", err)
			}
			oauthUser = user
		} else {
			oauthUser, err = deps.OAuthSvc.GetOrCreateUserByOAuth(ctx, authCodeInfo.Provider, &oauth.UserInfo{
				ProviderUserID: authCodeInfo.ProviderUserID,
				Email:          authCodeInfo.ProviderEmail,
				EmailVerified:  authCodeInfo.ProviderEmailVerified,
				Name:           authCodeInfo.ProviderName,
			})
			if err != nil {
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

		_, err = deps.AuthCodeRepo.MarkAuthorizationCodeAsUsed(ctx, codeHash)
		if err != nil {
			log.Error("oauth_token_exchange_mark_code_used_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to mark code as used", err)
		}

		accessToken, err := deps.TokenSvc.GenerateAccessToken(oauthUser.ID, oauthUser.Email)
		if err != nil {
			log.Error("oauth_token_exchange_generate_access_token_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to generate access token", err)
		}

		refreshToken, err := deps.TokenSvc.GenerateRefreshToken()
		if err != nil {
			log.Error("oauth_token_exchange_generate_refresh_token_failed",
				log.F("error", err.Error()),
				log.F("user_id", oauthUser.ID.String()),
			)
			return nil, huma.Error500InternalServerError("failed to generate refresh token", err)
		}

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

// generateAuthCode creates a cryptographically random authorization code and its SHA-256 hash.
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
