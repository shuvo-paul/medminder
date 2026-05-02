package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func RegisterRoutes(api huma.API, queries *db.Queries, jwtSecret string, emailClient email.EmailClient, frontendURL string) {
	// Repos (used to construct services)
	userRepo := repository.NewUserRepository(queries)
	tokenRepo := repository.NewRefreshTokenRepository(queries)
	tokenSvc := service.NewTokenService(jwtSecret)

	// Services
	authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)
	passwordResetSvc := service.NewPasswordResetService(
		userRepo,
		repository.NewPasswordResetTokenRepository(queries),
		tokenRepo,
		emailClient,
		frontendURL,
	)

	// Register
	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"auth"},
	}, handlers.RegisterHandler(authSvc))

	// Login
	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Login user",
		Tags:        []string{"auth"},
	}, handlers.LoginHandler(authSvc))

	// Logout
	logoutHandler := handlers.LogoutHandler(authSvc)

	huma.Register(api, huma.Operation{
		OperationID: "logout-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout",
		Summary:     "Logout user",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *dto.LogoutInput) (*dto.LogoutOutput, error) {
		authHeader := input.Header.Authorization
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return nil, huma.Error401Unauthorized("Invalid authorization header", nil)
		}
		tokenString := authHeader[7:]

		claims, err := tokenSvc.ValidateAccessToken(tokenString)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		userIDStr, ok := claims["sub"].(string)
		if !ok {
			return nil, huma.Error401Unauthorized("Invalid access token", nil)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid user ID in token", nil)
		}

		if err := logoutHandler(ctx, userID); err != nil {
			return nil, err
		}

		return &dto.LogoutOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "logged out successfully"}}, nil
	})

	// Password reset routes (already use dto types)
	passwordResetDeps := handlers.NewPasswordResetDeps(passwordResetSvc)
	handlers.RegisterPasswordResetRoutes(api, passwordResetDeps)
}
