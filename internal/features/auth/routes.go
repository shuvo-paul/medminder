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
	}, func(ctx context.Context, input *handlers.RegisterInput) (*registerOutput, error) {
		resp, err := handler(ctx, input)
		if err != nil {
			if errors.Is(err, handlers.ErrInvalidInput) {
				return nil, huma.Error400BadRequest("Invalid input", err)
			}
			if errors.Is(err, handlers.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already exists", err)
			}
			return nil, err
		}
		out := &registerOutput{}
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.User = resp.User
		return out, nil
	})
}
