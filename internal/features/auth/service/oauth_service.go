package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/log"
	db "github.com/shuvo-paul/medminder/internal/database/sqlc"
	auditRepo "github.com/shuvo-paul/medminder/internal/features/audit/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/middleware"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

type OAuthService interface {
	GetOrCreateUserByOAuth(ctx context.Context, provider string, userInfo *oauth.UserInfo) (*OAuthUser, error)
	GetUserByOAuth(ctx context.Context, provider, providerUserID string) (*OAuthUser, error)
	LinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error
	UnlinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider string) error
	GetUserOAuthProviders(ctx context.Context, userID uuid.UUID) ([]string, error)
	CanUnlinkProvider(ctx context.Context, userID uuid.UUID, provider string) (bool, error)
}

type oauthService struct {
	userRepo         repository.UserRepository
	oauthAccountRepo repository.OAuthAccountRepository
	auditRepo        auditRepo.AuditRepository
}

func NewOAuthService(userRepo repository.UserRepository, oauthAccountRepo repository.OAuthAccountRepository, auditRepository auditRepo.AuditRepository) OAuthService {
	return &oauthService{
		userRepo:         userRepo,
		oauthAccountRepo: oauthAccountRepo,
		auditRepo:        auditRepository,
	}
}

type OAuthUser struct {
	ID            uuid.UUID
	Email         string
	DisplayName   string
	EmailVerified bool
}

func (s *oauthService) GetOrCreateUserByOAuth(ctx context.Context, provider string, userInfo *oauth.UserInfo) (*OAuthUser, error) {
	if userInfo == nil {
		return nil, fmt.Errorf("userInfo is required: %w", ErrOAuthProviderError)
	}

	existingAccount, err := s.oauthAccountRepo.GetOAuthAccountByProviderAndUserID(ctx, provider, userInfo.ProviderUserID)
	if err == nil {
		// Existing OAuth account — this is a login.
		user, err := s.userRepo.GetUserByID(ctx, existingAccount.UserID.String())
		if err != nil {
			return nil, err
		}
		s.logAudit(ctx, "oauth_login", existingAccount.UserID, map[string]string{
			"provider": provider,
		})
		return s.toOAuthUser(&user), nil
	}
	if !errors.Is(err, repository.ErrOAuthAccountNotFound) {
		return nil, err
	}

	_, err = s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err == nil {
		log.Info("oauth_login_failed", log.F("provider", provider), log.F("email", userInfo.Email), log.F("reason", "email_exists"))
		return nil, ErrEmailExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	newUser, err := s.userRepo.CreateUser(ctx, userInfo.Email, userInfo.Name, "", userInfo.EmailVerified)
	if err != nil {
		return nil, err
	}

	_, err = s.oauthAccountRepo.CreateOAuthAccount(ctx, uuid.New(), newUser.ID, provider, userInfo.ProviderUserID)
	if err != nil {
		// Best effort cleanup: verify email to mark user as verified despite OAuth failure
		if verifyErr := s.userRepo.VerifyEmail(ctx, newUser.ID); verifyErr != nil {
			log.Error("oauth_user_cleanup_failed", log.F("user_id", newUser.ID.String()), log.F("error", verifyErr.Error()))
		}
		return nil, ErrOAuthProviderFailed
	}

	s.logAudit(ctx, "oauth_connected", newUser.ID, map[string]string{
		"provider": provider,
	})

	return s.toCreateUserRow(&newUser), nil
}

func (s *oauthService) GetUserByOAuth(ctx context.Context, provider, providerUserID string) (*OAuthUser, error) {
	account, err := s.oauthAccountRepo.GetOAuthAccountByProviderAndUserID(ctx, provider, providerUserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, account.UserID.String())
	if err != nil {
		return nil, err
	}

	return s.toOAuthUser(&user), nil
}

func (s *oauthService) toOAuthUser(user *db.User) *OAuthUser {
	return &OAuthUser{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerified.Bool,
	}
}

func (s *oauthService) toCreateUserRow(user *db.CreateUserRow) *OAuthUser {
	return &OAuthUser{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerified.Bool,
	}
}

// logAudit records an OAuth audit event with IP and User-Agent from the request context.
func (s *oauthService) logAudit(ctx context.Context, eventType string, userID uuid.UUID, metadata map[string]string) {
	ip := middleware.IPFromContext(ctx)
	ua := middleware.UserAgentFromContext(ctx)
	_ = s.auditRepo.LogEvent(ctx, eventType, uuid.NullUUID{UUID: userID, Valid: true}, ip, ua, metadata)
}

// LinkOAuthAccount links an OAuth account to an existing user.
// If the user already has this provider linked with a different account, it replaces the link.
// Returns ErrAccountWillBeLocked if linking would leave the user with no login method.
func (s *oauthService) LinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	// Check if user already has this provider linked
	existingAccount, err := s.oauthAccountRepo.GetOAuthAccountByUserIDAndProvider(ctx, userID, provider)
	accountExists := !errors.Is(err, repository.ErrOAuthAccountNotFound)

	// Get user to check password_hash
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return err
	}

	// Check dead-end prevention: if password_hash is NULL and GetOAuthAccountsByUserID
	// returns only this provider being linked (or nothing), reject
	oauthAccounts, err := s.oauthAccountRepo.GetOAuthAccountsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	hasPassword := user.PasswordHash.Valid && user.PasswordHash.String != ""
	providerCount := len(oauthAccounts)

	if accountExists {
		// existingAccount is the current link - we're replacing it
		// Dead-end check: if no password and (only this provider OR upserting same provider)
		if !hasPassword && providerCount <= 1 {
			return ErrAccountWillBeLocked
		}
		// Create new link first, then delete old one to avoid partial failure
		// where user loses OAuth link if create succeeds but delete fails.
		// Brief window of duplicate links is acceptable; loss of link is not.
		_, err = s.oauthAccountRepo.CreateOAuthAccount(ctx, uuid.New(), userID, provider, providerUserID)
		if err != nil {
			return err
		}
		_, err = s.oauthAccountRepo.DeleteOAuthAccount(ctx, existingAccount.ID)
		if err != nil {
			return err
		}
	}
	// No existing link - new linking
	// Dead-end check: if no password and no other providers, reject
	if !hasPassword && providerCount == 0 {
		return ErrAccountWillBeLocked
	}

	// Create new OAuth account link
	_, err = s.oauthAccountRepo.CreateOAuthAccount(ctx, uuid.New(), userID, provider, providerUserID)
	if err != nil {
		return err
	}

	s.logAudit(ctx, "oauth_linked", userID, map[string]string{
		"provider": provider,
	})

	return nil
}

// UnlinkOAuthAccount removes an OAuth account link from a user.
// Returns ErrOAuthAccountNotFound if the user doesn't have this provider linked.
// Returns ErrAccountWillBeLocked if unlinking would leave the user with no login method.
func (s *oauthService) UnlinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider string) error {
	// Get the OAuth account for this user + provider
	_, err := s.oauthAccountRepo.GetOAuthAccountByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, repository.ErrOAuthAccountNotFound) {
			return repository.ErrOAuthAccountNotFound
		}
		return err
	}

	// Get user to check password_hash
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return err
	}

	// Dead-end prevention: if password_hash is NULL, reject
	if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
		return ErrAccountWillBeLocked
	}

	// Check: if this is the user's ONLY login method (no other OAuth providers), reject
	oauthAccounts, err := s.oauthAccountRepo.GetOAuthAccountsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// If only one provider and we're removing it, check if user has password
	if len(oauthAccounts) == 1 && oauthAccounts[0].Provider == provider {
		if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
			return ErrAccountWillBeLocked
		}
	}

	// Delete the OAuth account
	err = s.oauthAccountRepo.DeleteOAuthAccountByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		return err
	}

	s.logAudit(ctx, "oauth_unlinked", userID, map[string]string{
		"provider": provider,
	})

	return nil
}

// GetUserOAuthProviders returns the list of OAuth providers linked to a user.
func (s *oauthService) GetUserOAuthProviders(ctx context.Context, userID uuid.UUID) ([]string, error) {
	accounts, err := s.oauthAccountRepo.GetOAuthAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	providers := make([]string, len(accounts))
	for i, account := range accounts {
		providers[i] = account.Provider
	}
	return providers, nil
}

// CanUnlinkProvider returns true if removing the specified provider won't lock the user out.
func (s *oauthService) CanUnlinkProvider(ctx context.Context, userID uuid.UUID, provider string) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return false, err
	}

	// If user has a password, they can always unlink
	if user.PasswordHash.Valid && user.PasswordHash.String != "" {
		return true, nil
	}

	// No password - check if they have multiple OAuth providers
	accounts, err := s.oauthAccountRepo.GetOAuthAccountsByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	// Count providers other than the one being unlinked
	otherProviders := 0
	for _, account := range accounts {
		if account.Provider != provider {
			otherProviders++
		}
	}

	return otherProviders > 0, nil
}
