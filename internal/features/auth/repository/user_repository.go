package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash string, emailVerified bool) (db.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id string) (db.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	VerifyEmail(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string, emailVerified bool) (db.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:         email,
		DisplayName:   displayName,
		PasswordHash:  sql.NullString{String: passwordHash, Valid: true},
		EmailVerified: sql.NullBool{Bool: emailVerified, Valid: true},
	})
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (db.User, error) {
	return r.queries.GetUserByID(ctx, uuid.MustParse(id))
}

func (r *userRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           uuid.MustParse(id),
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
	})
}

func (r *userRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	return r.queries.VerifyUserEmail(ctx, id)
}
