package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

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

func RegisterHandler(svc service.AuthService) func(context.Context, *RegisterInput) (*RegisterOutput, error) {
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

		result, err := svc.Register(ctx, input.Email, input.DisplayName, input.Password)
		if err != nil {
			if errors.Is(err, service.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already exists", err)
			}
			return nil, err
		}

		resp := &RegisterOutput{}
		resp.User.ID = result.User.ID
		resp.User.Email = result.User.Email
		resp.User.DisplayName = result.User.DisplayName
		resp.User.EmailVerified = result.User.EmailVerified

		return resp, nil
	}
}
