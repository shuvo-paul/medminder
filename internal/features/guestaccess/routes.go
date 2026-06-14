package guestaccess

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	authService "github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/handlers"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/repository"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/service"
	profileRepo "github.com/shuvo-paul/medminder/internal/features/profiles/repository"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
)

func RegisterRoutes(api huma.API, queries *db.Queries, tokenSvc authService.TokenServiceInterface) {
	guestRepo := repository.NewGuestAccessRepository(queries)
	permChecker := profileService.NewPermissionChecker(queries)
	guestSvc := service.NewGuestAccessService(guestRepo, permChecker)

	doseScheduleRepo := profileRepo.NewDoseScheduleRepository(queries)

	adminPerm := middleware.HumaRequireProfilePermission(permChecker, tokenSvc, "profile:admin", "profile:owner")

	huma.Register(api, huma.Operation{
		OperationID: "create-guest-access",
		Method:      "POST",
		Path:        "/api/profiles/{id}/guest-access",
		Summary:     "Create a guest access token",
		Tags:        []string{"guest-access"},
		Middlewares: []func(huma.Context, func(huma.Context)){adminPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.CreateGuestAccessHandler(guestSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "list-guest-access-tokens",
		Method:      "GET",
		Path:        "/api/profiles/{id}/guest-access",
		Summary:     "List active guest access tokens for a profile",
		Tags:        []string{"guest-access"},
		Middlewares: []func(huma.Context, func(huma.Context)){adminPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ListGuestAccessTokensHandler(guestSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "revoke-guest-access",
		Method:      "DELETE",
		Path:        "/api/guest-access/{tokenId}",
		Summary:     "Revoke a guest access token",
		Tags:        []string{"guest-access"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.RevokeGuestAccessHandler(guestSvc, tokenSvc, permChecker))

	huma.Register(api, huma.Operation{
		OperationID: "guest-list-medications",
		Method:      "GET",
		Path:        "/api/guest/{token}/medications",
		Summary:     "List medications via guest access token",
		Tags:        []string{"guest"},
	}, handlers.GuestListMedicationsHandler(guestSvc, doseScheduleRepo))

	huma.Register(api, huma.Operation{
		OperationID: "guest-get-reminder",
		Method:      "GET",
		Path:        "/api/guest/{token}/reminders/{reminderId}",
		Summary:     "Get a reminder via guest access token",
		Tags:        []string{"guest"},
	}, handlers.GuestGetReminderHandler(guestSvc, doseScheduleRepo))
}
