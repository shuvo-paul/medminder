package auth

import (
	"database/sql"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	auditRepo "github.com/shuvo-paul/medminder/internal/features/audit/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

func RegisterRoutes(api huma.API, dbConn *sql.DB, queries *db.Queries, auditRepository auditRepo.AuditRepository, jwtSecret string, emailClient email.EmailClient, frontendURL string, profileSvc profileService.ProfileService) {
	// Repos (used to construct services)
	userRepo := repository.NewUserRepository(queries)
	tokenRepo := repository.NewRefreshTokenRepository(queries)
	evTokenRepo := repository.NewEmailVerificationTokenRepository(queries)
	emailChangeRepo := repository.NewEmailChangeRepository(queries)
	tokenSvc := service.NewTokenService(jwtSecret)

	// Services
	authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)
	emailVerifSvc := service.NewEmailVerificationService(userRepo, evTokenRepo, tokenSvc, frontendURL)
	emailChangeSvc := service.NewEmailChangeService(userRepo, emailChangeRepo, tokenSvc, frontendURL)
	passwordResetSvc := service.NewPasswordResetService(
		userRepo,
		repository.NewPasswordResetTokenRepository(queries),
		tokenRepo,
		emailClient,
		frontendURL,
	)

	// OAuth
	oauthAccountRepo := repository.NewOAuthAccountRepository(queries)
	oauthAuthCodeRepo := repository.NewOAuthAuthorizationCodeRepository(dbConn, queries)
	oauthSvc := service.NewOAuthService(userRepo, oauthAccountRepo, auditRepository)

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

	// Request email change (authenticated)
	huma.Register(api, huma.Operation{
		OperationID: "request-email-change",
		Method:      http.MethodPost,
		Path:        "/api/auth/email/change/request",
		Summary:     "Request an email address change",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.RequestEmailChangeHandler(emailChangeSvc, tokenSvc))

	// Cancel email change (authenticated)
	huma.Register(api, huma.Operation{
		OperationID: "cancel-email-change",
		Method:      http.MethodPost,
		Path:        "/api/auth/email/change/cancel",
		Summary:     "Cancel a pending email address change",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.CancelEmailChangeHandler(emailChangeSvc, tokenSvc))

	// Get pending email change (authenticated)
	huma.Register(api, huma.Operation{
		OperationID: "get-pending-email-change",
		Method:      http.MethodGet,
		Path:        "/api/auth/email/change/pending",
		Summary:     "Get pending email change request",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.GetPendingEmailChangeHandler(emailChangeSvc, tokenSvc))

	// Verify updated email (unauthenticated - token in body)
	huma.Register(api, huma.Operation{
		OperationID: "verify-updated-email",
		Method:      http.MethodPost,
		Path:        "/api/auth/email/verify-updated",
		Summary:     "Verify updated email address from email change request",
		Tags:        []string{"auth"},
	}, handlers.VerifyUpdatedEmailHandler(emailChangeSvc))

	// Password reset routes (already use dto types)
	passwordResetDeps := handlers.NewPasswordResetDeps(passwordResetSvc)
	handlers.RegisterPasswordResetRoutes(api, passwordResetDeps)

	// OAuth provider routes (unauthenticated)
	handlers.RegisterOAuthProviderRoutes(api, oauthAuthCodeRepo, oauthSvc, tokenSvc, tokenRepo, auditRepository, frontendURL)

	// Set password (authenticated) — for OAuth-only users to add a password
	huma.Register(api, huma.Operation{
		OperationID: "set-password",
		Method:      http.MethodPost,
		Path:        "/api/auth/password/set",
		Summary:     "Set a password for an OAuth-only user",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.SetPasswordHandler(authSvc, tokenSvc))

	// Change password (authenticated) — requires current password
	huma.Register(api, huma.Operation{
		OperationID: "change-password",
		Method:      http.MethodPut,
		Path:        "/api/auth/password",
		Summary:     "Change password",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ChangePasswordHandler(authSvc, tokenSvc))

	// Delete account (authenticated) — requires password confirmation
	huma.Register(api, huma.Operation{
		OperationID: "delete-account",
		Method:      http.MethodDelete,
		Path:        "/api/auth/account",
		Summary:     "Delete the authenticated user's account",
		Tags:        []string{"auth"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.DeleteAccountHandler(authSvc, profileSvc, tokenSvc))
}
