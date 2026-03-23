package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
}
