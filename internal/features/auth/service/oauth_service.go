package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/pkg/oauth"
)

type OAuthService interface {
	GetOrCreateUserByOAuth(ctx context.Context, provider string, userInfo *oauth.UserInfo) (*User, error)
	GetUserByOAuth(ctx context.Context, provider, providerUserID string) (*User, error)
}

type oauthService struct {
	userRepo         repository.UserRepository
	oauthAccountRepo repository.OAuthAccountRepository
}

func NewOAuthService(userRepo repository.UserRepository, oauthAccountRepo repository.OAuthAccountRepository) OAuthService {
	return &oauthService{
		userRepo:         userRepo,
		oauthAccountRepo: oauthAccountRepo,
	}
}

type User struct {
	ID            uuid.UUID
	Email         string
	DisplayName   string
	EmailVerified bool
}

func (s *oauthService) GetOrCreateUserByOAuth(ctx context.Context, provider string, userInfo *oauth.UserInfo) (*User, error) {
	existingAccount, err := s.oauthAccountRepo.GetOAuthAccountByProviderAndUserID(ctx, provider, userInfo.ProviderUserID)
	if err == nil {
		user, err := s.userRepo.GetUserByID(ctx, existingAccount.UserID.String())
		if err != nil {
			return nil, err
		}
		return &User{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			EmailVerified: user.EmailVerified.Bool,
		}, nil
	}
	if !errors.Is(err, repository.ErrOAuthAccountNotFound) {
		return nil, err
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err == nil && existingUser.Email != "" {
		log.Info("oauth login failed: email exists", log.F("provider", provider), log.F("email", userInfo.Email))
		return nil, ErrEmailExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	newUser, err := s.userRepo.CreateUser(ctx, userInfo.Email, userInfo.Name, "", userInfo.EmailVerified)
	if err != nil {
		return nil, err
	}

	_, err = s.oauthAccountRepo.CreateOAuthAccount(ctx, uuid.New(), newUser.ID, provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, err
	}

	log.Info("oauth connected", log.F("provider", provider), log.F("user_id", newUser.ID.String()), log.F("email", userInfo.Email))

	return &User{
		ID:            newUser.ID,
		Email:         newUser.Email,
		DisplayName:   newUser.DisplayName,
		EmailVerified: newUser.EmailVerified.Bool,
	}, nil
}

func (s *oauthService) GetUserByOAuth(ctx context.Context, provider, providerUserID string) (*User, error) {
	account, err := s.oauthAccountRepo.GetOAuthAccountByProviderAndUserID(ctx, provider, providerUserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, account.UserID.String())
	if err != nil {
		return nil, err
	}

	return &User{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerified.Bool,
	}, nil
}
