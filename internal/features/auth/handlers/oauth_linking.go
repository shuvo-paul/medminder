package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/shuvo-paul/medminder/internal/common/auth"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

// OAuthLinkInitHandler returns a handler that initiates an OAuth account linking flow.
func OAuthLinkInitHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthLinkInitInput) (*dto.OAuthLinkInitResponse, error) {
	return func(ctx context.Context, input *dto.OAuthLinkInitInput) (*dto.OAuthLinkInitResponse, error) {
		userID, err := auth.ExtractUserID(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		_, err = oauth.GetProvider(input.Provider)
		if err != nil {
			return nil, huma.Error404NotFound("provider not found", err)
		}

		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			return nil, huma.Error500InternalServerError("failed to generate nonce", err)
		}
		nonce := hex.EncodeToString(nonceBytes)

		redirect := input.Body.Redirect
		if redirect != "" && !isValidRedirect(redirect) {
			return nil, huma.Error400BadRequest("invalid redirect URL", errors.New("redirect must be a relative path"))
		}
		if redirect == "" {
			redirect = "/account"
		}

		state := dto.OAuthState{
			Nonce:    nonce,
			Redirect: redirect,
			Purpose:  "link",
		}
		encodedState := state.Encode()

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

// OAuthAccountsHandler returns a handler that lists all OAuth accounts linked to the user.
func OAuthAccountsHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthAccountsInput) (*dto.OAuthAccountsResponse, error) {
	return func(ctx context.Context, input *dto.OAuthAccountsInput) (*dto.OAuthAccountsResponse, error) {
		userID, err := auth.ExtractUserID(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
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
func OAuthUnlinkHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthUnlinkInput) (*dto.OAuthUnlinkOutput, error) {
	return func(ctx context.Context, input *dto.OAuthUnlinkInput) (*dto.OAuthUnlinkOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
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
func OAuthLinkStatusHandler(deps *OAuthHandlerDeps) func(context.Context, *dto.OAuthLinkStatusInput) (*dto.OAuthLinkStatusResponse, error) {
	return func(ctx context.Context, input *dto.OAuthLinkStatusInput) (*dto.OAuthLinkStatusResponse, error) {
		userID, err := auth.ExtractUserID(input.Authorization, deps.TokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
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
