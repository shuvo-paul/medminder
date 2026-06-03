package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
)

type RegisterResult struct {
	User struct {
		ID            uuid.UUID
		Email         string
		DisplayName   string
		EmailVerified bool
	}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         struct {
		ID            uuid.UUID
		Email         string
		DisplayName   string
		EmailVerified bool
	}
}

type AuthService interface {
	Register(ctx context.Context, email, displayName, password string) (*RegisterResult, error)
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	SetPassword(ctx context.Context, userID uuid.UUID, password string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
	tokenSvc  TokenServiceInterface
}

var _ AuthService = (*authService)(nil)

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository, tokenSvc TokenServiceInterface) *authService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokenSvc:  tokenSvc,
	}
}

func (s *authService) Register(ctx context.Context, email, displayName, password string) (*RegisterResult, error) {
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.CreateUser(ctx, email, displayName, string(hashedPassword), false)
	if err != nil {
		return nil, err
	}

	result := &RegisterResult{}
	result.User.ID = user.ID
	result.User.Email = user.Email
	result.User.DisplayName = user.DisplayName
	result.User.EmailVerified = user.EmailVerified.Bool

	return result, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash := s.tokenSvc.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(RefreshTokenExpiry)
	if _, err := s.tokenRepo.CreateRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, err
	}

	result := &LoginResult{}
	result.AccessToken = accessToken
	result.RefreshToken = refreshToken
	result.User.ID = user.ID
	result.User.Email = user.Email
	result.User.DisplayName = user.DisplayName
	result.User.EmailVerified = user.EmailVerified.Bool

	return result, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrLogoutFailed
	}

	if err := s.tokenRepo.DeleteUserRefreshTokens(ctx, userID); err != nil {
		return errors.Join(ErrLogoutFailed, err)
	}

	return nil
}

func (s *authService) SetPassword(ctx context.Context, userID uuid.UUID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID.String(), string(hashedPassword)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("getting user: %w", ErrUserNotFound)
		}
		return fmt.Errorf("getting user: %w", err)
	}

	if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
		return ErrNoPasswordSet
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(currentPassword)); err != nil {
		return ErrWrongPassword
	}

	if currentPassword == newPassword {
		return ErrSamePassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID.String(), string(hashedPassword)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
