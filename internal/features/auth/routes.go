package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func RegisterRoutes(api huma.API, queries *db.Queries, jwtSecret string, emailClient email.EmailClient, frontendURL string) {
	// Repos (used to construct services)
	userRepo := repository.NewUserRepository(queries)
	tokenRepo := repository.NewRefreshTokenRepository(queries)
	evTokenRepo := repository.NewEmailVerificationTokenRepository(queries)
	tokenSvc := service.NewTokenService(jwtSecret)

	// Services
	authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)
	emailVerifSvc := service.NewEmailVerificationService(userRepo, evTokenRepo, tokenSvc, frontendURL)
	passwordResetSvc := service.NewPasswordResetService(
		userRepo,
		repository.NewPasswordResetTokenRepository(queries),
		tokenRepo,
		emailClient,
		frontendURL,
	)

	// Register (with email verification)
	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"auth"},
	}, handlers.RegisterHandler(authSvc, emailVerifSvc))

	// Login
	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Login user",
		Tags:        []string{"auth"},
	}, handlers.LoginHandler(authSvc))

	// Logout
	huma.Register(api, huma.Operation{
		OperationID: "logout-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout",
		Summary:     "Logout user",
		Tags:        []string{"auth"},
	}, handlers.LogoutHandler(authSvc, tokenSvc))

	// Email verification (unauthenticated)
	huma.Register(api, huma.Operation{
		OperationID: "verify-email",
		Method:      http.MethodPost,
		Path:        "/api/auth/email/verify",
		Summary:     "Verify email address",
		Tags:        []string{"auth"},
	}, handlers.VerifyEmailHandler(emailVerifSvc))

	// Resend verification (authenticated)
	huma.Register(api, huma.Operation{
		OperationID: "resend-verification",
		Method:      http.MethodPost,
		Path:        "/api/auth/email/resend-verification",
		Summary:     "Resend verification email",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ResendVerificationHandler(emailVerifSvc, tokenSvc))

	// Password reset routes (already use dto types)
	passwordResetDeps := handlers.NewPasswordResetDeps(passwordResetSvc)
	handlers.RegisterPasswordResetRoutes(api, passwordResetDeps)
}
