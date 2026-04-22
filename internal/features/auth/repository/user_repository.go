package repository

import (
	"context"
	"database/sql"

	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
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
