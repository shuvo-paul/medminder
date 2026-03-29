package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:         email,
		DisplayName:   displayName,
		PasswordHash:  sql.NullString{String: passwordHash, Valid: true},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	})
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *userRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}
