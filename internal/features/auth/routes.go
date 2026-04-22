package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

type registerOutput struct {
	Body struct {
		User struct {
			ID            uuid.UUID `json:"id"`
			Email         string    `json:"email"`
			DisplayName   string    `json:"display_name"`
			EmailVerified bool      `json:"email_verified"`
		} `json:"user"`
	}
}

type loginInput struct {
	Body struct {
		Email    string `json:"email" minLength:"1" maxLength:"255"`
		Password string `json:"password" minLength:"1"`
	}
}

type loginOutput struct {
	Body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID            uuid.UUID `json:"id"`
			Email         string    `json:"email"`
			DisplayName   string    `json:"display_name"`
			EmailVerified bool      `json:"email_verified"`
		} `json:"user"`
	}
}

type registerInput struct {
	Body struct {
		Email       string `json:"email" minLength:"1" maxLength:"255"`
		DisplayName string `json:"display_name" minLength:"1" maxLength:"100"`
		Password    string `json:"password" minLength:"8"`
	}
}

func RegisterRoutes(api huma.API, queries *db.Queries, jwtSecret string) {
	repo := repository.NewUserRepository(queries)
	tokenSvc := service.NewTokenService(jwtSecret)
	handler := handlers.RegisterHandler(repo, tokenSvc)

	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *registerInput) (*registerOutput, error) {
		resp, err := handler(ctx, &handlers.RegisterInput{
			Email:       input.Body.Email,
			DisplayName: input.Body.DisplayName,
			Password:    input.Body.Password,
		})
		if err != nil {
			if errors.Is(err, handlers.ErrInvalidEmail) {
				return nil, huma.Error400BadRequest("Invalid email format", err)
			}
			if errors.Is(err, handlers.ErrInvalidPassword) {
				return nil, huma.Error400BadRequest("Password must be at least 8 characters with 1 uppercase, 1 lowercase, and 1 number", err)
			}
			if errors.Is(err, handlers.ErrInvalidDisplayName) {
				return nil, huma.Error400BadRequest("Display name must be 1-100 characters", err)
			}
			if errors.Is(err, handlers.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already exists", err)
			}
			return nil, err
		}
		out := &registerOutput{}
		out.Body.User = resp.User
		return out, nil
	})

	loginHandler := handlers.LoginHandler(repo, tokenSvc)

	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Login user",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *loginInput) (*loginOutput, error) {
		resp, err := loginHandler(ctx, &handlers.LoginInput{
			Email:    input.Body.Email,
			Password: input.Body.Password,
		})
		if err != nil {
			if errors.Is(err, handlers.ErrInvalidCredentials) {
				return nil, huma.Error401Unauthorized("Invalid email or password", err)
			}
			if errors.Is(err, handlers.ErrInvalidEmail) {
				return nil, huma.Error400BadRequest("Invalid email format", err)
			}
			return nil, err
		}
		out := &loginOutput{}
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.User = resp.User
		return out, nil
	})
}
