package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/repository"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

const (
	DefaultExpiryDays  = 30
	guestTokenByteSize = 32
)

type GuestAccessService interface {
	CreateToken(ctx context.Context, profileID uuid.UUID, label string, permissions []string, expiresInDays int) (*TokenResult, error)
	ListTokens(ctx context.Context, profileID uuid.UUID) ([]TokenResult, error)
	RevokeToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error
	Authenticate(ctx context.Context, rawToken string) (*AuthenticatedToken, error)
}

type TokenResult struct {
	ID         uuid.UUID
	Label      string
	ExpiresAt  time.Time
	RawToken   string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type AuthenticatedToken struct {
	TokenID     uuid.UUID
	ProfileID   uuid.UUID
	Permissions []string
}

type guestAccessService struct {
	repo        repository.GuestAccessRepository
	permChecker profileService.PermissionChecker
}

func NewGuestAccessService(repo repository.GuestAccessRepository, permChecker profileService.PermissionChecker) GuestAccessService {
	return &guestAccessService{
		repo:        repo,
		permChecker: permChecker,
	}
}

func (s *guestAccessService) CreateToken(ctx context.Context, profileID uuid.UUID, label string, permissions []string, expiresInDays int) (*TokenResult, error) {
	if expiresInDays <= 0 {
		expiresInDays = DefaultExpiryDays
	}

	if len(permissions) == 0 {
		permissions = []string{"medication:read", "reminder:read"}
	}

	permBytes, err := json.Marshal(permissions)
	if err != nil {
		return nil, ErrInvalidPermissions
	}

	tokenBytes := make([]byte, guestTokenByteSize)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	rawToken := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().AddDate(0, 0, expiresInDays)

	labelStr := repository.NewNullString(label)

	created, err := s.repo.Create(ctx, db.CreateGuestAccessTokenParams{
		ProfileID:   profileID,
		TokenHash:   tokenHash,
		Label:       labelStr,
		Permissions: permBytes,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, err
	}

	var lastUsedAt *time.Time
	if created.LastUsedAt.Valid {
		lastUsedAt = &created.LastUsedAt.Time
	}

	return &TokenResult{
		ID:         created.ID,
		Label:      label,
		ExpiresAt:  created.ExpiresAt,
		RawToken:   rawToken,
		CreatedAt:  created.CreatedAt,
		LastUsedAt: lastUsedAt,
	}, nil
}

func (s *guestAccessService) ListTokens(ctx context.Context, profileID uuid.UUID) ([]TokenResult, error) {
	tokens, err := s.repo.ListByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	results := make([]TokenResult, len(tokens))
	for i, t := range tokens {
		var lastUsedAt *time.Time
		if t.LastUsedAt.Valid {
			lastUsedAt = &t.LastUsedAt.Time
		}
		results[i] = TokenResult{
			ID:         t.ID,
			Label:      t.Label.String,
			ExpiresAt:  t.ExpiresAt,
			CreatedAt:  t.CreatedAt,
			LastUsedAt: lastUsedAt,
		}
	}
	return results, nil
}

func (s *guestAccessService) RevokeToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error {
	token, err := s.repo.GetByID(ctx, tokenID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrGuestTokenNotFound
		}
		return err
	}

	allowed, err := s.permChecker.HasAnyPermission(ctx, token.ProfileID, userID, []string{"profile:admin", "profile:owner"})
	if err != nil {
		return err
	}
	if !allowed {
		return ErrGuestTokenInsufficientPerms
	}

	return s.repo.Delete(ctx, tokenID)
}

func (s *guestAccessService) Authenticate(ctx context.Context, rawToken string) (*AuthenticatedToken, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := s.repo.GetByHash(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGuestTokenNotFound
		}
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, ErrGuestTokenExpired
	}

	if err := s.repo.UpdateLastUsedAt(ctx, token.ID); err != nil {
		return nil, err
	}

	var permissions []string
	if err := json.Unmarshal(token.Permissions, &permissions); err != nil {
		return nil, err
	}

	return &AuthenticatedToken{
		TokenID:     token.ID,
		ProfileID:   token.ProfileID,
		Permissions: permissions,
	}, nil
}
