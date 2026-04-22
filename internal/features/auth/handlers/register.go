package handlers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailExists = errors.New("email already exists")

type RegisterInput struct {
	Email       string `json:"email" minLength:"1" maxLength:"255" pattern:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
	DisplayName string `json:"display_name" minLength:"1" maxLength:"100"`
	Password    string `json:"password" minLength:"8"`
}

type RegisterOutput struct {
	User struct {
		ID            uuid.UUID `json:"id"`
		Email         string    `json:"email"`
		DisplayName   string    `json:"display_name"`
		EmailVerified bool      `json:"email_verified"`
	} `json:"user"`
}

func RegisterHandler(repo repository.UserRepository) func(context.Context, *RegisterInput) (*RegisterOutput, error) {
	return func(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
		if err := ValidateEmail(input.Email); err != nil {
			return nil, ErrInvalidEmail
		}
		if err := ValidatePassword(input.Password); err != nil {
			return nil, ErrInvalidPassword
		}
		if err := ValidateDisplayName(input.DisplayName); err != nil {
			return nil, ErrInvalidDisplayName
		}

		existingUser, err := repo.GetUserByEmail(ctx, input.Email)
		if err == nil && existingUser.Email != "" {
			return nil, ErrEmailExists
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), BcryptCost)
		if err != nil {
			return nil, err
		}

		user, err := repo.CreateUser(ctx, input.Email, input.DisplayName, string(hashedPassword))
		if err != nil {
			return nil, err
		}

		resp := &RegisterOutput{}
		resp.User.ID = user.ID
		resp.User.Email = user.Email
		resp.User.DisplayName = user.DisplayName
		resp.User.EmailVerified = user.EmailVerified.Bool

		return resp, nil
	}
}
